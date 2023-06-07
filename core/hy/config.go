package hy

import (
	"errors"
	"fmt"
	"github.com/yosuke-furukawa/json5/encoding/json5"
	"regexp"
	"strconv"
)

const (
	mbpsToBps   = 125000
	minSpeedBPS = 16384

	DefaultALPN = "hysteria"

	DefaultStreamReceiveWindow     = 16777216                           // 16 MB
	DefaultConnectionReceiveWindow = DefaultStreamReceiveWindow * 5 / 2 // 40 MB

	DefaultMaxIncomingStreams = 1024

	DefaultMMDBFilename = "GeoLite2-Country.mmdb"

	ServerMaxIdleTimeoutSec     = 60
	DefaultClientIdleTimeoutSec = 20

	DefaultClientHopIntervalSec = 10
)

var rateStringRegexp = regexp.MustCompile(`^(\d+)\s*([KMGT]?)([Bb])ps$`)

type serverConfig struct {
	Listen   string `json:"listen"`
	Protocol string `json:"protocol"`
	ACME     struct {
		Domains                 []string `json:"domains"`
		Email                   string   `json:"email"`
		DisableHTTPChallenge    bool     `json:"disable_http"`
		DisableTLSALPNChallenge bool     `json:"disable_tlsalpn"`
		AltHTTPPort             int      `json:"alt_http_port"`
		AltTLSALPNPort          int      `json:"alt_tlsalpn_port"`
	} `json:"acme"`
	CertFile string `json:"cert"`
	KeyFile  string `json:"key"`
	// Optional below
	Up         string `json:"up"`
	UpMbps     int    `json:"up_mbps"`
	Down       string `json:"down"`
	DownMbps   int    `json:"down_mbps"`
	DisableUDP bool   `json:"disable_udp"`
	ACL        string `json:"acl"`
	MMDB       string `json:"mmdb"`
	Obfs       string `json:"obfs"`
	Auth       struct {
		Mode   string           `json:"mode"`
		Config json5.RawMessage `json:"config"`
	} `json:"auth"`
	ALPN                string `json:"alpn"`
	PrometheusListen    string `json:"prometheus_listen"`
	ReceiveWindowConn   uint64 `json:"recv_window_conn"`
	ReceiveWindowClient uint64 `json:"recv_window_client"`
	MaxConnClient       int    `json:"max_conn_client"`
	DisableMTUDiscovery bool   `json:"disable_mtu_discovery"`
	Resolver            string `json:"resolver"`
	ResolvePreference   string `json:"resolve_preference"`
	SOCKS5Outbound      struct {
		Server   string `json:"server"`
		User     string `json:"user"`
		Password string `json:"password"`
	} `json:"socks5_outbound"`
	BindOutbound struct {
		Address string `json:"address"`
		Device  string `json:"device"`
	} `json:"bind_outbound"`
}

func (c *serverConfig) Speed() (uint64, uint64, error) {
	var up, down uint64
	if len(c.Up) > 0 {
		up = stringToBps(c.Up)
		if up == 0 {
			return 0, 0, errors.New("invalid speed format")
		}
	} else {
		up = uint64(c.UpMbps) * mbpsToBps
	}
	if len(c.Down) > 0 {
		down = stringToBps(c.Down)
		if down == 0 {
			return 0, 0, errors.New("invalid speed format")
		}
	} else {
		down = uint64(c.DownMbps) * mbpsToBps
	}
	return up, down, nil
}

func (c *serverConfig) Check() error {
	if len(c.Listen) == 0 {
		return errors.New("missing listen address")
	}
	if len(c.ACME.Domains) == 0 && (len(c.CertFile) == 0 || len(c.KeyFile) == 0) {
		return errors.New("need either ACME info or cert/key files")
	}
	if len(c.ACME.Domains) > 0 && (len(c.CertFile) > 0 || len(c.KeyFile) > 0) {
		return errors.New("cannot use both ACME and cert/key files, they are mutually exclusive")
	}
	if up, down, err := c.Speed(); err != nil || (up != 0 && up < minSpeedBPS) || (down != 0 && down < minSpeedBPS) {
		return errors.New("invalid speed")
	}
	if (c.ReceiveWindowConn != 0 && c.ReceiveWindowConn < 65536) ||
		(c.ReceiveWindowClient != 0 && c.ReceiveWindowClient < 65536) {
		return errors.New("invalid receive window size")
	}
	if c.MaxConnClient < 0 {
		return errors.New("invalid max connections per client")
	}
	return nil
}

func (c *serverConfig) Fill() {
	if len(c.ALPN) == 0 {
		c.ALPN = DefaultALPN
	}
	if c.ReceiveWindowConn == 0 {
		c.ReceiveWindowConn = DefaultStreamReceiveWindow
	}
	if c.ReceiveWindowClient == 0 {
		c.ReceiveWindowClient = DefaultConnectionReceiveWindow
	}
	if c.MaxConnClient == 0 {
		c.MaxConnClient = DefaultMaxIncomingStreams
	}
	if len(c.MMDB) == 0 {
		c.MMDB = DefaultMMDBFilename
	}
}

func (c *serverConfig) String() string {
	return fmt.Sprintf("%+v", *c)
}

type Relay struct {
	Listen  string `json:"listen"`
	Remote  string `json:"remote"`
	Timeout int    `json:"timeout"`
}

func (r *Relay) Check() error {
	if len(r.Listen) == 0 {
		return errors.New("missing relay listen address")
	}
	if len(r.Remote) == 0 {
		return errors.New("missing relay remote address")
	}
	if r.Timeout != 0 && r.Timeout < 4 {
		return errors.New("invalid relay timeout")
	}
	return nil
}

type clientConfig struct {
	Server   string `json:"server"`
	Protocol string `json:"protocol"`
	Up       string `json:"up"`
	UpMbps   int    `json:"up_mbps"`
	Down     string `json:"down"`
	DownMbps int    `json:"down_mbps"`
	// Optional below
	Retry            int  `json:"retry"`
	RetryInterval    *int `json:"retry_interval"`
	QuitOnDisconnect bool `json:"quit_on_disconnect"`
	HandshakeTimeout int  `json:"handshake_timeout"`
	IdleTimeout      int  `json:"idle_timeout"`
	HopInterval      int  `json:"hop_interval"`
	SOCKS5           struct {
		Listen     string `json:"listen"`
		Timeout    int    `json:"timeout"`
		DisableUDP bool   `json:"disable_udp"`
		User       string `json:"user"`
		Password   string `json:"password"`
	} `json:"socks5"`
	HTTP struct {
		Listen   string `json:"listen"`
		Timeout  int    `json:"timeout"`
		User     string `json:"user"`
		Password string `json:"password"`
		Cert     string `json:"cert"`
		Key      string `json:"key"`
	} `json:"http"`
	TUN struct {
		Name                     string `json:"name"`
		Timeout                  int    `json:"timeout"`
		MTU                      uint32 `json:"mtu"`
		TCPSendBufferSize        string `json:"tcp_sndbuf"`
		TCPReceiveBufferSize     string `json:"tcp_rcvbuf"`
		TCPModerateReceiveBuffer bool   `json:"tcp_autotuning"`
	} `json:"tun"`
	TCPRelays []Relay `json:"relay_tcps"`
	TCPRelay  Relay   `json:"relay_tcp"` // deprecated, but we still support it for backward compatibility
	UDPRelays []Relay `json:"relay_udps"`
	UDPRelay  Relay   `json:"relay_udp"` // deprecated, but we still support it for backward compatibility
	TCPTProxy struct {
		Listen  string `json:"listen"`
		Timeout int    `json:"timeout"`
	} `json:"tproxy_tcp"`
	UDPTProxy struct {
		Listen  string `json:"listen"`
		Timeout int    `json:"timeout"`
	} `json:"tproxy_udp"`
	TCPRedirect struct {
		Listen  string `json:"listen"`
		Timeout int    `json:"timeout"`
	} `json:"redirect_tcp"`
	ACL                 string `json:"acl"`
	MMDB                string `json:"mmdb"`
	Obfs                string `json:"obfs"`
	Auth                []byte `json:"auth"`
	AuthString          string `json:"auth_str"`
	ALPN                string `json:"alpn"`
	ServerName          string `json:"server_name"`
	Insecure            bool   `json:"insecure"`
	CustomCA            string `json:"ca"`
	ReceiveWindowConn   uint64 `json:"recv_window_conn"`
	ReceiveWindow       uint64 `json:"recv_window"`
	DisableMTUDiscovery bool   `json:"disable_mtu_discovery"`
	FastOpen            bool   `json:"fast_open"`
	LazyStart           bool   `json:"lazy_start"`
	Resolver            string `json:"resolver"`
	ResolvePreference   string `json:"resolve_preference"`
}

func (c *clientConfig) Speed() (uint64, uint64, error) {
	var up, down uint64
	if len(c.Up) > 0 {
		up = stringToBps(c.Up)
		if up == 0 {
			return 0, 0, errors.New("invalid speed format")
		}
	} else {
		up = uint64(c.UpMbps) * mbpsToBps
	}
	if len(c.Down) > 0 {
		down = stringToBps(c.Down)
		if down == 0 {
			return 0, 0, errors.New("invalid speed format")
		}
	} else {
		down = uint64(c.DownMbps) * mbpsToBps
	}
	return up, down, nil
}

func (c *clientConfig) Fill() {
	if len(c.ALPN) == 0 {
		c.ALPN = DefaultALPN
	}
	if c.ReceiveWindowConn == 0 {
		c.ReceiveWindowConn = DefaultStreamReceiveWindow
	}
	if c.ReceiveWindow == 0 {
		c.ReceiveWindow = DefaultConnectionReceiveWindow
	}
	if len(c.MMDB) == 0 {
		c.MMDB = DefaultMMDBFilename
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = DefaultClientIdleTimeoutSec
	}
	if c.HopInterval == 0 {
		c.HopInterval = DefaultClientHopIntervalSec
	}
}

func (c *clientConfig) String() string {
	return fmt.Sprintf("%+v", *c)
}

func stringToBps(s string) uint64 {
	if s == "" {
		return 0
	}
	m := rateStringRegexp.FindStringSubmatch(s)
	if m == nil {
		return 0
	}
	var n uint64
	switch m[2] {
	case "K":
		n = 1 << 10
	case "M":
		n = 1 << 20
	case "G":
		n = 1 << 30
	case "T":
		n = 1 << 40
	default:
		n = 1
	}
	v, _ := strconv.ParseUint(m[1], 10, 64)
	n = v * n
	if m[3] == "b" {
		// Bits, need to convert to bytes
		n = n >> 3
	}
	return n
}
