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

	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/goccy/go-json"
	"github.com/inazumav/sing-box/option"
)

type WsNetworkConfig struct {
	Path string `json:"path"`
}

func getInboundOptions(tag string, info *panel.NodeInfo, c *conf.ControllerConfig) (option.Inbound, error) {
	addr, _ := netip.ParseAddr("0.0.0.0")
	listen := option.ListenOptions{
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
		in.Type = "vmess"
		t := option.V2RayTransportOptions{
			Type: info.Network,
		}
		switch info.Network {
		case "tcp":
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
			t.WebsocketOptions = option.V2RayWebsocketOptions{
				Path:                u.Path,
				EarlyDataHeaderName: "Sec-WebSocket-Protocol",
				MaxEarlyData:        uint32(ed),
			}
		case "grpc":
			t.GRPCOptions = option.V2RayGRPCOptions{
				ServiceName: info.ServerName,
			}
		}
		in.VMessOptions = option.VMessInboundOptions{
			ListenOptions: listen,
			TLS:           &tls,
			Transport:     &t,
		}
	case "shadowsocks":
		in.Type = "shadowsocks"
		p := make([]byte, 32)
		_, _ = rand.Read(p)
		randomPasswd := hex.EncodeToString(p)
		if strings.Contains(info.Cipher, "2022") {
			randomPasswd = base64.StdEncoding.EncodeToString([]byte(randomPasswd))
		}
		in.ShadowsocksOptions = option.ShadowsocksInboundOptions{
			ListenOptions: listen,
			Method:        info.Cipher,
			Password:      info.ServerKey,
			Users: []option.ShadowsocksUser{
				{
					Password: randomPasswd,
				},
			},
		}
	}
	return in, nil
}

func (b *Box) AddNode(tag string, info *panel.NodeInfo, config *conf.ControllerConfig) error {
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
