package hy

import (
	"crypto/tls"
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/limiter"
	"github.com/apernet/hysteria/core/sockopt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/quic-go/quic-go"

	"github.com/apernet/hysteria/core/acl"
	"github.com/apernet/hysteria/core/cs"
	"github.com/apernet/hysteria/core/pktconns"
	"github.com/apernet/hysteria/core/pmtud"
	"github.com/apernet/hysteria/core/transport"
	"github.com/sirupsen/logrus"
)

var serverPacketConnFuncFactoryMap = map[string]pktconns.ServerPacketConnFuncFactory{
	"":             pktconns.NewServerUDPConnFunc,
	"udp":          pktconns.NewServerUDPConnFunc,
	"wechat":       pktconns.NewServerWeChatConnFunc,
	"wechat-video": pktconns.NewServerWeChatConnFunc,
	"faketcp":      pktconns.NewServerFakeTCPConnFunc,
}

type Server struct {
	tag     string
	l       *limiter.Limiter
	counter *UserTrafficCounter
	users   sync.Map
	running atomic.Bool
	*cs.Server
}

func NewServer(tag string, l *limiter.Limiter) *Server {
	return &Server{
		tag: tag,
		l:   l,
	}
}

func (s *Server) runServer(node *panel.NodeInfo, c *conf.ControllerConfig) error {
	/*if c.HyOptions == nil {
		return errors.New("hy options is not vail")
	}*/
	// Resolver
	if len(c.HyOptions.Resolver) > 0 {
		err := setResolver(c.HyOptions.Resolver)
		if err != nil {
			return fmt.Errorf("set resolver error: %s", err)
		}
	}
	// tls config
	kpl, err := newKeypairLoader(c.CertConfig.CertFile, c.CertConfig.KeyFile)
	if err != nil {
		return fmt.Errorf("load cert error: %s", err)
	}
	tlsConfig := &tls.Config{
		GetCertificate: kpl.GetCertificateFunc(),
		NextProtos:     []string{DefaultALPN},
		MinVersion:     tls.VersionTLS13,
	}
	// QUIC config
	quicConfig := &quic.Config{
		InitialStreamReceiveWindow:     DefaultStreamReceiveWindow,
		MaxStreamReceiveWindow:         DefaultStreamReceiveWindow,
		InitialConnectionReceiveWindow: DefaultConnectionReceiveWindow,
		MaxConnectionReceiveWindow:     DefaultConnectionReceiveWindow,
		MaxIncomingStreams:             int64(DefaultMaxIncomingStreams),
		MaxIdleTimeout:                 ServerMaxIdleTimeoutSec * time.Second,
		KeepAlivePeriod:                0, // Keep alive should solely be client's responsibility
		DisablePathMTUDiscovery:        false,
		EnableDatagrams:                true,
	}
	if !quicConfig.DisablePathMTUDiscovery && pmtud.DisablePathMTUDiscovery {
		logrus.Info("Path MTU Discovery is not yet supported on this platform")
	}
	// Resolve preference
	if len(c.HyOptions.ResolvePreference) > 0 {
		pref, err := transport.ResolvePreferenceFromString(c.HyOptions.Resolver)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to parse the resolve preference")
		}
		transport.DefaultServerTransport.ResolvePreference = pref
	}
	/*// SOCKS5 outbound
	if config.SOCKS5Outbound.Server != "" {
		transport.DefaultServerTransport.SOCKS5Client = transport.NewSOCKS5Client(config.SOCKS5Outbound.Server,
			config.SOCKS5Outbound.User, config.SOCKS5Outbound.Password)
	}*/
	// Bind outbound
	if c.HyOptions.SendDevice != "" {
		iface, err := net.InterfaceByName(c.HyOptions.SendDevice)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to find the interface")
		}
		transport.DefaultServerTransport.LocalUDPIntf = iface
		sockopt.BindDialer(transport.DefaultServerTransport.Dialer, iface)
	}
	if c.SendIP != "" {
		ip := net.ParseIP(c.SendIP)
		if ip == nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("Failed to parse the address")
		}
		transport.DefaultServerTransport.Dialer.LocalAddr = &net.TCPAddr{IP: ip}
		transport.DefaultServerTransport.LocalUDPAddr = &net.UDPAddr{IP: ip}
	}
	// ACL
	var aclEngine *acl.Engine
	/*if len(config.ACL) > 0 {
		aclEngine, err = acl.LoadFromFile(config.ACL, func(addr string) (*net.IPAddr, error) {
			ipAddr, _, err := transport.DefaultServerTransport.ResolveIPAddr(addr)
			return ipAddr, err
		},
			func() (*geoip2.Reader, error) {
				return loadMMDBReader(config.MMDB)
			})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"file":  config.ACL,
			}).Fatal("Failed to parse ACL")
		}
		aclEngine.DefaultAction = acl.ActionDirect
	}*/
	// Prometheus
	s.counter = NewUserTrafficCounter()
	// Packet conn
	pktConnFuncFactory := serverPacketConnFuncFactoryMap[""]
	if pktConnFuncFactory == nil {
		return fmt.Errorf("unsopport protocol")
	}
	pktConnFunc := pktConnFuncFactory(node.HyObfs)
	addr := fmt.Sprintf("%s:%d", c.ListenIP, node.Port)
	pktConn, err := pktConnFunc(addr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"addr":  addr,
		}).Fatal("Failed to listen on the UDP address")
	}
	// Server
	up, down := SpeedTrans(node.UpMbps, node.DownMbps)
	s.Server, err = cs.NewServer(tlsConfig, quicConfig, pktConn,
		transport.DefaultServerTransport, up, down, false, aclEngine,
		s.connectFunc, s.disconnectFunc, tcpRequestFunc, tcpErrorFunc, udpRequestFunc, udpErrorFunc, s.counter)
	if err != nil {
		return fmt.Errorf("new server error: %s", err)
	}
	logrus.WithField("addr", addr).Info("Server up and running")
	go func() {
		s.running.Store(true)
		defer func() {
			s.running.Store(false)
		}()
		err = s.Server.Serve()
		if err != nil {
			logrus.WithField("addr", addr).Errorf("serve error: %s", err)
		}
	}()
	return nil
}

func (s *Server) authByUser(addr net.Addr, auth []byte, sSend uint64, sRecv uint64) (bool, string) {
	if _, r := s.l.CheckLimit(string(auth), addr.String(), false); r {
		return false, "device limited"
	}
	if _, ok := s.users.Load(string(auth)); ok {
		return true, "Done"
	}
	return false, "Failed"
}

func (s *Server) connectFunc(addr net.Addr, auth []byte, sSend uint64, sRecv uint64) (bool, string) {
	s.l.ConnLimiter.AddConnCount(addr.String(), string(auth), false)
	ok, msg := s.authByUser(addr, auth, sSend, sRecv)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"src": defaultIPMasker.Mask(addr.String()),
		}).Info("Authentication failed, client rejected")
		return false, msg
	}
	logrus.WithFields(logrus.Fields{
		"src":  defaultIPMasker.Mask(addr.String()),
		"Uuid": string(auth),
		"Tag":  s.tag,
	}).Info("Client connected")
	return ok, msg
}

func (s *Server) disconnectFunc(addr net.Addr, auth []byte, err error) {
	s.l.ConnLimiter.DelConnCount(addr.String(), string(auth))
	logrus.WithFields(logrus.Fields{
		"src":   defaultIPMasker.Mask(addr.String()),
		"error": err,
	}).Info("Client disconnected")
}

func tcpRequestFunc(addr net.Addr, auth []byte, reqAddr string, action acl.Action, arg string) {
	logrus.WithFields(logrus.Fields{
		"src":    defaultIPMasker.Mask(addr.String()),
		"dst":    defaultIPMasker.Mask(reqAddr),
		"action": actionToString(action, arg),
	}).Debug("TCP request")
}

func tcpErrorFunc(addr net.Addr, auth []byte, reqAddr string, err error) {
	if err != io.EOF {
		logrus.WithFields(logrus.Fields{
			"src":   defaultIPMasker.Mask(addr.String()),
			"dst":   defaultIPMasker.Mask(reqAddr),
			"error": err,
		}).Info("TCP error")
	} else {
		logrus.WithFields(logrus.Fields{
			"src": defaultIPMasker.Mask(addr.String()),
			"dst": defaultIPMasker.Mask(reqAddr),
		}).Debug("TCP EOF")
	}
}

func udpRequestFunc(addr net.Addr, auth []byte, sessionID uint32) {
	logrus.WithFields(logrus.Fields{
		"src":     defaultIPMasker.Mask(addr.String()),
		"session": sessionID,
	}).Debug("UDP request")
}

func udpErrorFunc(addr net.Addr, auth []byte, sessionID uint32, err error) {
	if err != io.EOF {
		logrus.WithFields(logrus.Fields{
			"src":     defaultIPMasker.Mask(addr.String()),
			"session": sessionID,
			"error":   err,
		}).Info("UDP error")
	} else {
		logrus.WithFields(logrus.Fields{
			"src":     defaultIPMasker.Mask(addr.String()),
			"session": sessionID,
		}).Debug("UDP EOF")
	}
}

func actionToString(action acl.Action, arg string) string {
	switch action {
	case acl.ActionDirect:
		return "Direct"
	case acl.ActionProxy:
		return "Proxy"
	case acl.ActionBlock:
		return "Block"
	case acl.ActionHijack:
		return "Hijack to " + arg
	default:
		return "Unknown"
	}
}
