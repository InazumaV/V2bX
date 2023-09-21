package sing

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/inazumav/sing-box/inbound"
	F "github.com/sagernet/sing/common/format"

	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/conf"
	"github.com/goccy/go-json"
	"github.com/inazumav/sing-box/option"
)

type WsNetworkConfig struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

func getInboundOptions(tag string, info *panel.NodeInfo, c *conf.Options) (option.Inbound, error) {
	addr, err := netip.ParseAddr(c.ListenIP)
	if err != nil {
		return option.Inbound{}, fmt.Errorf("the listen ip not vail")
	}
	var domainStrategy option.DomainStrategy
	if c.SingOptions.EnableDNS {
		domainStrategy = c.SingOptions.DomainStrategy
	}
	listen := option.ListenOptions{
		Listen:        (*option.ListenAddress)(&addr),
		ListenPort:    uint16(info.Common.ServerPort),
		ProxyProtocol: c.SingOptions.EnableProxyProtocol,
		TCPFastOpen:   c.SingOptions.TCPFastOpen,
		InboundOptions: option.InboundOptions{
			SniffEnabled:             c.SingOptions.SniffEnabled,
			SniffOverrideDestination: c.SingOptions.SniffOverrideDestination,
			DomainStrategy:           domainStrategy,
		},
	}
	var tls option.InboundTLSOptions
	switch info.Security {
	case panel.Tls:
		if c.CertConfig == nil {
			return option.Inbound{}, fmt.Errorf("the CertConfig is not vail")
		}
		switch c.CertConfig.CertMode {
		case "none", "":
			break // disable
		default:
			tls.Enabled = true
			tls.CertificatePath = c.CertConfig.CertFile
			tls.KeyPath = c.CertConfig.KeyFile
		}
	case panel.Reality:
		tls.Enabled = true
		v := info.VAllss
		tls.ServerName = v.TlsSettings.ServerName
		dest, _ := strconv.Atoi(v.TlsSettings.ServerPort)
		mtd, _ := time.ParseDuration(v.RealityConfig.MaxTimeDiff)
		tls.Reality = &option.InboundRealityOptions{
			Enabled:    true,
			ShortID:    []string{v.TlsSettings.ShortId},
			PrivateKey: v.TlsSettings.PrivateKey,
			Handshake: option.InboundRealityHandshakeOptions{
				ServerOptions: option.ServerOptions{
					Server:     tls.ServerName,
					ServerPort: uint16(dest),
				},
			},
			MaxTimeDifference: option.Duration(mtd),
		}
	}
	in := option.Inbound{
		Tag: tag,
	}
	switch info.Type {
	case "vmess", "vless":
		n := info.VAllss
		t := option.V2RayTransportOptions{
			Type: n.Network,
		}
		switch n.Network {
		case "tcp":
			t.Type = ""
		case "ws":
			var (
				path    string
				ed      int
				headers map[string]option.Listable[string]
			)
			if len(n.NetworkSettings) != 0 {
				network := WsNetworkConfig{}
				err := json.Unmarshal(n.NetworkSettings, &network)
				if err != nil {
					return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
				}
				var u *url.URL
				u, err = url.Parse(network.Path)
				if err != nil {
					return option.Inbound{}, fmt.Errorf("parse path error: %s", err)
				}
				path = u.Path
				ed, _ = strconv.Atoi(u.Query().Get("ed"))
				headers = make(map[string]option.Listable[string], len(network.Headers))
				for k, v := range network.Headers {
					headers[k] = option.Listable[string]{
						v,
					}
				}
			}
			t.WebsocketOptions = option.V2RayWebsocketOptions{
				Path:                path,
				EarlyDataHeaderName: "Sec-WebSocket-Protocol",
				MaxEarlyData:        uint32(ed),
				Headers:             headers,
			}
		case "grpc":
			if len(n.NetworkSettings) != 0 {
				err := json.Unmarshal(n.NetworkSettings, &t.GRPCOptions)
				if err != nil {
					return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
				}
			}
		}
		if info.Type == "vless" {
			in.Type = "vless"
			in.VLESSOptions = option.VLESSInboundOptions{
				ListenOptions: listen,
				TLS:           &tls,
				Transport:     &t,
			}
		} else {
			in.Type = "vmess"
			in.VMessOptions = option.VMessInboundOptions{
				ListenOptions: listen,
				TLS:           &tls,
				Transport:     &t,
			}
		}
	case "shadowsocks":
		in.Type = "shadowsocks"
		n := info.Shadowsocks
		var keyLength int
		switch n.Cipher {
		case "2022-blake3-aes-128-gcm":
			keyLength = 16
		case "2022-blake3-aes-256-gcm":
			keyLength = 32
		default:
			keyLength = 16
		}
		in.ShadowsocksOptions = option.ShadowsocksInboundOptions{
			ListenOptions: listen,
			Method:        n.Cipher,
		}
		p := make([]byte, keyLength)
		_, _ = rand.Read(p)
		randomPasswd := string(p)
		if strings.Contains(n.Cipher, "2022") {
			in.ShadowsocksOptions.Password = n.ServerKey
			randomPasswd = base64.StdEncoding.EncodeToString([]byte(randomPasswd))
		}
		in.ShadowsocksOptions.Users = []option.ShadowsocksUser{{
			Password: randomPasswd,
		}}
	case "trojan":
		in.Type = "trojan"
		in.TrojanOptions = option.TrojanInboundOptions{
			ListenOptions: listen,
			TLS:           &tls,
		}
		if c.SingOptions.FallBackConfigs != nil {
			// fallback handling
			fallback := c.SingOptions.FallBackConfigs.FallBack
			fallbackPort, err := strconv.Atoi(fallback.ServerPort)
			if err == nil {
				in.TrojanOptions.Fallback = &option.ServerOptions{
					Server:     fallback.Server,
					ServerPort: uint16(fallbackPort),
				}
			}
			fallbackForALPNMap := c.SingOptions.FallBackConfigs.FallBackForALPN
			fallbackForALPN := make(map[string]*option.ServerOptions, len(fallbackForALPNMap))
			if err := processFallback(c, fallbackForALPN); err == nil {
				in.TrojanOptions.FallbackForALPN = fallbackForALPN
			}
		}
	case "hysteria":
		in.Type = "hysteria"
		in.HysteriaOptions = option.HysteriaInboundOptions{
			ListenOptions: listen,
			UpMbps:        info.Hysteria.UpMbps,
			DownMbps:      info.Hysteria.DownMbps,
			Obfs:          info.Hysteria.Obfs,
			TLS:           &tls,
		}
	}
	return in, nil
}

func (b *Box) AddNode(tag string, info *panel.NodeInfo, config *conf.Options) error {
	err := updateDNSConfig(info)
	if err != nil {
		return fmt.Errorf("build dns error: %s", err)
	}
	c, err := getInboundOptions(tag, info, config)
	if err != nil {
		return err
	}
	in, err := inbound.New(
		context.Background(),
		b.router,
		b.logFactory.NewLogger(F.ToString("inbound/", c.Type, "[", tag, "]")),
		c,
		nil,
	)
	if err != nil {
		return fmt.Errorf("init inbound error： %s", err)
	}
	err = in.Start()
	if err != nil {
		return fmt.Errorf("start inbound error: %s", err)
	}
	b.inbounds[tag] = in
	err = b.router.AddInbound(in)
	if err != nil {
		return fmt.Errorf("add inbound error: %s", err)
	}
	return nil
}

func (b *Box) DelNode(tag string) error {
	err := b.inbounds[tag].Close()
	if err != nil {
		return fmt.Errorf("close inbound error: %s", err)
	}
	err = b.router.DelInbound(tag)
	if err != nil {
		return fmt.Errorf("delete inbound error: %s", err)
	}
	return nil
}
