package sing

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"

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
	listen := option.ListenOptions{
		//ProxyProtocol: true,
		Listen:     (*option.ListenAddress)(&addr),
		ListenPort: uint16(info.Port),
	}
	tls := option.InboundTLSOptions{
		Enabled:         info.Tls,
		CertificatePath: c.CertConfig.CertFile,
		KeyPath:         c.CertConfig.KeyFile,
		ServerName:      info.ServerName,
	}
	in := option.Inbound{
		Tag: tag,
	}
	switch info.Type {
	case "v2ray":
		t := option.V2RayTransportOptions{
			Type: info.Network,
		}
		switch info.Network {
		case "tcp":
			t.Type = ""
		case "ws":
			network := WsNetworkConfig{}
			err := json.Unmarshal(info.NetworkSettings, &network)
			if err != nil {
				return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
			}
			var u *url.URL
			u, err = url.Parse(network.Path)
			if err != nil {
				return option.Inbound{}, fmt.Errorf("parse path error: %s", err)
			}
			ed, _ := strconv.Atoi(u.Query().Get("ed"))
			h := make(map[string]option.Listable[string], len(network.Headers))
			for k, v := range network.Headers {
				h[k] = option.Listable[string]{
					v,
				}
			}
			t.WebsocketOptions = option.V2RayWebsocketOptions{
				Path:                u.Path,
				EarlyDataHeaderName: "Sec-WebSocket-Protocol",
				MaxEarlyData:        uint32(ed),
				Headers:             h,
			}
		case "grpc":
			err := json.Unmarshal(info.NetworkSettings, &t.GRPCOptions)
			if err != nil {
				return option.Inbound{}, fmt.Errorf("decode NetworkSettings error: %s", err)
			}
		}
		if info.ExtraConfig.EnableVless == "true" {
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
		var keyLength int
		switch info.Cipher {
		case "2022-blake3-aes-128-gcm":
			keyLength = 16
		case "2022-blake3-aes-256-gcm":
			keyLength = 32
		default:
			keyLength = 16
		}
		in.ShadowsocksOptions = option.ShadowsocksInboundOptions{
			ListenOptions: listen,
			Method:        info.Cipher,
		}
		p := make([]byte, keyLength)
		_, _ = rand.Read(p)
		randomPasswd := hex.EncodeToString(p)
		if strings.Contains(info.Cipher, "2022") {
			in.ShadowsocksOptions.Password = info.ServerKey
		}
		in.ShadowsocksOptions.Users = []option.ShadowsocksUser{{
			Password: base64.StdEncoding.EncodeToString([]byte(randomPasswd)),
		}}
	}
	return in, nil
}

func (b *Box) AddNode(tag string, info *panel.NodeInfo, config *conf.Options) error {
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
	b.inbounds[tag] = in
	err = in.Start()
	if err != nil {
		return fmt.Errorf("start inbound error: %s", err)
	}
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
