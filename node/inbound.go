package node

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/common/file"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/node/lego"
	"github.com/goccy/go-json"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/core"
	coreConf "github.com/xtls/xray-core/infra/conf"
)

// BuildInbound build Inbound config for different protocol
func BuildInbound(config *conf.ControllerConfig, nodeInfo *panel.NodeInfo, tag string) (*core.InboundHandlerConfig, error) {
	inbound := &coreConf.InboundDetourConfig{}
	// Set network protocol
	t := coreConf.TransportProtocol(nodeInfo.Network)
	inbound.StreamSetting = &coreConf.StreamConfig{Network: &t}
	var err error
	switch nodeInfo.NodeType {
	case "v2ray":
		err = buildV2ray(config, nodeInfo, inbound)
	case "trojan":
		err = buildTrojan(config, nodeInfo, inbound)
	case "shadowsocks":
		err = buildShadowsocks(config, nodeInfo, inbound)
	default:
		return nil, fmt.Errorf("unsupported node type: %s, Only support: V2ray, Trojan, Shadowsocks", nodeInfo.NodeType)
	}
	if err != nil {
		return nil, err
	}
	// Set server port
	inbound.PortList = &coreConf.PortList{
		Range: []coreConf.PortRange{{From: uint32(nodeInfo.ServerPort), To: uint32(nodeInfo.ServerPort)}},
	}
	// Set Listen IP address
	ipAddress := net.ParseAddress(config.ListenIP)
	inbound.ListenOn = &coreConf.Address{Address: ipAddress}
	// Set SniffingConfig
	sniffingConfig := &coreConf.SniffingConfig{
		Enabled:      true,
		DestOverride: &coreConf.StringList{"http", "tls"},
	}
	if config.DisableSniffing {
		sniffingConfig.Enabled = false
	}
	inbound.SniffingConfig = sniffingConfig
	if *inbound.StreamSetting.Network == "tcp" {
		if inbound.StreamSetting.TCPSettings != nil {
			inbound.StreamSetting.TCPSettings.AcceptProxyProtocol = config.EnableProxyProtocol
		} else {
			tcpSetting := &coreConf.TCPConfig{
				AcceptProxyProtocol: config.EnableProxyProtocol,
			} //Enable proxy protocol
			inbound.StreamSetting.TCPSettings = tcpSetting
		}
	} else if *inbound.StreamSetting.Network == "ws" {
		inbound.StreamSetting.WSSettings = &coreConf.WebSocketConfig{
			AcceptProxyProtocol: config.EnableProxyProtocol} //Enable proxy protocol
	}
	// Set TLS and XTLS settings
	if nodeInfo.Tls != 0 {
		if config.CertConfig.CertMode != "none" {
			// Normal tls
			inbound.StreamSetting.Security = "tls"
			certFile, keyFile, err := getCertFile(config.CertConfig)
			if err != nil {
				return nil, err
			}
			inbound.StreamSetting.TLSSettings = &coreConf.TLSConfig{
				Certs: []*coreConf.TLSCertConfig{
					{
						CertFile:     certFile,
						KeyFile:      keyFile,
						OcspStapling: 3600,
					},
				},
				RejectUnknownSNI: config.CertConfig.RejectUnknownSni,
			}
		} else if config.EnableReality {
			// Reality
			inbound.StreamSetting.Security = "reality"
			inbound.StreamSetting.REALITYSettings = &coreConf.REALITYConfig{
				Dest:         config.RealityConfig.Dest,
				Xver:         config.RealityConfig.Xver,
				ServerNames:  config.RealityConfig.ServerNames,
				PrivateKey:   config.RealityConfig.PrivateKey,
				MinClientVer: config.RealityConfig.MinClientVer,
				MaxClientVer: config.RealityConfig.MaxClientVer,
				MaxTimeDiff:  config.RealityConfig.MaxTimeDiff,
				ShortIds:     config.RealityConfig.ShortIds,
			}
		}
	}
	// Support ProxyProtocol for any transport protocol
	if *inbound.StreamSetting.Network != "tcp" &&
		*inbound.StreamSetting.Network != "ws" &&
		config.EnableProxyProtocol {
		sockoptConfig := &coreConf.SocketConfig{
			AcceptProxyProtocol: config.EnableProxyProtocol,
		} //Enable proxy protocol
		inbound.StreamSetting.SocketSettings = sockoptConfig
	}
	inbound.Tag = tag
	return inbound.Build()
}

func buildV2ray(config *conf.ControllerConfig, nodeInfo *panel.NodeInfo, inbound *coreConf.InboundDetourConfig) error {
	if config.EnableVless {
		//Set vless
		inbound.Protocol = "vless"
		if config.EnableFallback {
			// Set fallback
			fallbackConfigs, err := buildVlessFallbacks(config.FallBackConfigs)
			if err != nil {
				return err
			}
			s, err := json.Marshal(&coreConf.VLessInboundConfig{
				Decryption: "none",
				Fallbacks:  fallbackConfigs,
			})
			if err != nil {
				return fmt.Errorf("marshal vless fallback config error: %s", err)
			}
			inbound.Settings = (*json.RawMessage)(&s)
		} else {
			var err error
			s, err := json.Marshal(&coreConf.VLessInboundConfig{
				Decryption: "none",
			})
			if err != nil {
				return fmt.Errorf("marshal vless config error: %s", err)
			}
			inbound.Settings = (*json.RawMessage)(&s)
		}
	} else {
		// Set vmess
		inbound.Protocol = "vmess"
		var err error
		s, err := json.Marshal(&coreConf.VMessInboundConfig{})
		if err != nil {
			return fmt.Errorf("marshal vmess settings error: %s", err)
		}
		inbound.Settings = (*json.RawMessage)(&s)
	}
	if len(nodeInfo.NetworkSettings) == 0 {
		return nil
	}
	switch nodeInfo.Network {
	case "tcp":
		err := json.Unmarshal(nodeInfo.NetworkSettings, &inbound.StreamSetting.TCPSettings)
		if err != nil {
			return fmt.Errorf("unmarshal tcp settings error: %s", err)
		}
	case "ws":
		err := json.Unmarshal(nodeInfo.NetworkSettings, &inbound.StreamSetting.WSSettings)
		if err != nil {
			return fmt.Errorf("unmarshal ws settings error: %s", err)
		}
	case "grpc":
		err := json.Unmarshal(nodeInfo.NetworkSettings, &inbound.StreamSetting.GRPCConfig)
		if err != nil {
			return fmt.Errorf("unmarshal grpc settings error: %s", err)
		}
	}
	return nil
}

func buildTrojan(config *conf.ControllerConfig, nodeInfo *panel.NodeInfo, inbound *coreConf.InboundDetourConfig) error {
	inbound.Protocol = "trojan"
	if config.EnableFallback {
		// Set fallback
		fallbackConfigs, err := buildTrojanFallbacks(config.FallBackConfigs)
		if err != nil {
			return err
		}
		s, err := json.Marshal(&coreConf.TrojanServerConfig{
			Fallbacks: fallbackConfigs,
		})
		inbound.Settings = (*json.RawMessage)(&s)
		if err != nil {
			return fmt.Errorf("marshal trojan fallback config error: %s", err)
		}
	} else {
		s := []byte("{}")
		inbound.Settings = (*json.RawMessage)(&s)
	}
	t := coreConf.TransportProtocol(nodeInfo.Network)
	inbound.StreamSetting = &coreConf.StreamConfig{Network: &t}
	return nil
}

func buildShadowsocks(config *conf.ControllerConfig, nodeInfo *panel.NodeInfo, inbound *coreConf.InboundDetourConfig) error {
	inbound.Protocol = "shadowsocks"
	settings := &coreConf.ShadowsocksServerConfig{
		Cipher: nodeInfo.Cipher,
	}
	p := make([]byte, 32)
	_, err := rand.Read(p)
	if err != nil {
		return fmt.Errorf("generate random password error: %s", err)
	}
	randomPasswd := hex.EncodeToString(p)
	cipher := nodeInfo.Cipher
	if nodeInfo.ServerKey != "" {
		settings.Password = nodeInfo.ServerKey
		randomPasswd = base64.StdEncoding.EncodeToString([]byte(randomPasswd))
		cipher = ""
	}
	defaultSSuser := &coreConf.ShadowsocksUserConfig{
		Cipher:   cipher,
		Password: randomPasswd,
	}
	settings.Users = append(settings.Users, defaultSSuser)
	settings.NetworkList = &coreConf.NetworkList{"tcp", "udp"}
	settings.IVCheck = true
	if config.DisableIVCheck {
		settings.IVCheck = false
	}
	t := coreConf.TransportProtocol("tcp")
	inbound.StreamSetting = &coreConf.StreamConfig{Network: &t}
	s, err := json.Marshal(settings)
	inbound.Settings = (*json.RawMessage)(&s)
	if err != nil {
		return fmt.Errorf("marshal shadowsocks settings error: %s", err)
	}
	return nil
}

func getCertFile(certConfig *conf.CertConfig) (certFile string, keyFile string, err error) {
	if certConfig.CertFile == "" || certConfig.KeyFile == "" {
		return "", "", fmt.Errorf("cert file path or key file path not exist")
	}
	switch certConfig.CertMode {
	case "file":
		return certConfig.CertFile, certConfig.KeyFile, nil
	case "dns", "http":
		if file.IsExist(certConfig.CertFile) && file.IsExist(certConfig.KeyFile) {
			return certConfig.CertFile, certConfig.KeyFile, nil
		}
		l, err := lego.New(certConfig)
		if err != nil {
			return "", "", fmt.Errorf("create lego object error: %s", err)
		}
		err = l.CreateCert()
		if err != nil {
			return "", "", fmt.Errorf("create cert error: %s", err)
		}
		return certConfig.CertFile, certConfig.KeyFile, nil
	}
	return "", "", fmt.Errorf("unsupported certmode: %s", certConfig.CertMode)
}

func buildVlessFallbacks(fallbackConfigs []conf.FallBackConfig) ([]*coreConf.VLessInboundFallback, error) {
	if fallbackConfigs == nil {
		return nil, fmt.Errorf("you must provide FallBackConfigs")
	}
	vlessFallBacks := make([]*coreConf.VLessInboundFallback, len(fallbackConfigs))
	for i, c := range fallbackConfigs {
		if c.Dest == "" {
			return nil, fmt.Errorf("dest is required for fallback fialed")
		}
		var dest json.RawMessage
		dest, err := json.Marshal(c.Dest)
		if err != nil {
			return nil, fmt.Errorf("marshal dest %s config fialed: %s", dest, err)
		}
		vlessFallBacks[i] = &coreConf.VLessInboundFallback{
			Name: c.SNI,
			Alpn: c.Alpn,
			Path: c.Path,
			Dest: dest,
			Xver: c.ProxyProtocolVer,
		}
	}
	return vlessFallBacks, nil
}

func buildTrojanFallbacks(fallbackConfigs []conf.FallBackConfig) ([]*coreConf.TrojanInboundFallback, error) {
	if fallbackConfigs == nil {
		return nil, fmt.Errorf("you must provide FallBackConfigs")
	}

	trojanFallBacks := make([]*coreConf.TrojanInboundFallback, len(fallbackConfigs))
	for i, c := range fallbackConfigs {

		if c.Dest == "" {
			return nil, fmt.Errorf("dest is required for fallback fialed")
		}

		var dest json.RawMessage
		dest, err := json.Marshal(c.Dest)
		if err != nil {
			return nil, fmt.Errorf("marshal dest %s config fialed: %s", dest, err)
		}
		trojanFallBacks[i] = &coreConf.TrojanInboundFallback{
			Name: c.SNI,
			Alpn: c.Alpn,
			Path: c.Path,
			Dest: dest,
			Xver: c.ProxyProtocolVer,
		}
	}
	return trojanFallBacks, nil
}
