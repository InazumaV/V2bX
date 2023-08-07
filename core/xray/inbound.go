package xray

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/conf"
	"github.com/goccy/go-json"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/core"
	coreConf "github.com/xtls/xray-core/infra/conf"
)

// BuildInbound build Inbound config for different protocol
func buildInbound(config *conf.Options, nodeInfo *panel.NodeInfo, tag string) (*core.InboundHandlerConfig, error) {
	in := &coreConf.InboundDetourConfig{}
	// Set network protocol
	t := coreConf.TransportProtocol(nodeInfo.Network)
	in.StreamSetting = &coreConf.StreamConfig{Network: &t}
	var err error
	switch nodeInfo.Type {
	case "v2ray":
		err = buildV2ray(config, nodeInfo, in)
	case "trojan":
		err = buildTrojan(config, in)
	case "shadowsocks":
		err = buildShadowsocks(config, nodeInfo, in)
	default:
		return nil, fmt.Errorf("unsupported node type: %s, Only support: V2ray, Trojan, Shadowsocks", nodeInfo.Type)
	}
	if err != nil {
		return nil, err
	}
	// Set server port
	in.PortList = &coreConf.PortList{
		Range: []coreConf.PortRange{{From: uint32(nodeInfo.Port), To: uint32(nodeInfo.Port)}},
	}
	// Set Listen IP address
	ipAddress := net.ParseAddress(config.ListenIP)
	in.ListenOn = &coreConf.Address{Address: ipAddress}
	// Set SniffingConfig
	sniffingConfig := &coreConf.SniffingConfig{
		Enabled:      true,
		DestOverride: &coreConf.StringList{"http", "tls"},
	}
	if config.XrayOptions.DisableSniffing {
		sniffingConfig.Enabled = false
	}
	in.SniffingConfig = sniffingConfig
	if *in.StreamSetting.Network == "tcp" {
		if in.StreamSetting.TCPSettings != nil {
			in.StreamSetting.TCPSettings.AcceptProxyProtocol = config.XrayOptions.EnableProxyProtocol
		} else {
			tcpSetting := &coreConf.TCPConfig{
				AcceptProxyProtocol: config.XrayOptions.EnableProxyProtocol,
			} //Enable proxy protocol
			in.StreamSetting.TCPSettings = tcpSetting
		}
	} else if *in.StreamSetting.Network == "ws" {
		in.StreamSetting.WSSettings = &coreConf.WebSocketConfig{
			AcceptProxyProtocol: config.XrayOptions.EnableProxyProtocol} //Enable proxy protocol
	}
	// Set TLS or Reality settings
	if nodeInfo.Tls {
		if config.CertConfig == nil {
			return nil, errors.New("the CertConfig is not vail")
		}
		switch config.CertConfig.CertMode {
		case "none", "":
			break // disable
		case "reality":
			// Reality
			in.StreamSetting.Security = "reality"
			d, err := json.Marshal(config.CertConfig.RealityConfig.Dest)
			if err != nil {
				return nil, fmt.Errorf("marshal reality dest error: %s", err)
			}
			if len(config.CertConfig.RealityConfig.ShortIds) == 0 {
				config.CertConfig.RealityConfig.ShortIds = []string{""}
			}
			in.StreamSetting.REALITYSettings = &coreConf.REALITYConfig{
				Dest:         d,
				Xver:         config.CertConfig.RealityConfig.Xver,
				ServerNames:  config.CertConfig.RealityConfig.ServerNames,
				PrivateKey:   config.CertConfig.RealityConfig.PrivateKey,
				MinClientVer: config.CertConfig.RealityConfig.MinClientVer,
				MaxClientVer: config.CertConfig.RealityConfig.MaxClientVer,
				MaxTimeDiff:  config.CertConfig.RealityConfig.MaxTimeDiff,
				ShortIds:     config.CertConfig.RealityConfig.ShortIds,
			}
			break
		case "remote":
			if nodeInfo.ExtraConfig.EnableReality == "true" {
				rc := nodeInfo.ExtraConfig.RealityConfig
				in.StreamSetting.Security = "reality"
				d, err := json.Marshal(rc.Dest)
				if err != nil {
					return nil, fmt.Errorf("marshal reality dest error: %s", err)
				}
				if len(rc.ShortIds) == 0 {
					rc.ShortIds = []string{""}
				}
				Xver, _ := strconv.ParseUint(rc.Xver, 10, 64)
				MaxTimeDiff, _ := strconv.ParseUint(rc.Xver, 10, 64)
				in.StreamSetting.REALITYSettings = &coreConf.REALITYConfig{
					Dest:         d,
					Xver:         Xver,
					ServerNames:  rc.ServerNames,
					PrivateKey:   rc.PrivateKey,
					MinClientVer: rc.MinClientVer,
					MaxClientVer: rc.MaxClientVer,
					MaxTimeDiff:  MaxTimeDiff,
					ShortIds:     rc.ShortIds,
				}
				break
			}
		default:
			{
				// Normal tls
				in.StreamSetting.Security = "tls"
				in.StreamSetting.TLSSettings = &coreConf.TLSConfig{
					Certs: []*coreConf.TLSCertConfig{
						{
							CertFile:     config.CertConfig.CertFile,
							KeyFile:      config.CertConfig.KeyFile,
							OcspStapling: 3600,
						},
					},
					RejectUnknownSNI: config.CertConfig.RejectUnknownSni,
				}
			}
		}
	}
	// Support ProxyProtocol for any transport protocol
	if *in.StreamSetting.Network != "tcp" &&
		*in.StreamSetting.Network != "ws" &&
		config.XrayOptions.EnableProxyProtocol {
		socketConfig := &coreConf.SocketConfig{
			AcceptProxyProtocol: config.XrayOptions.EnableProxyProtocol,
			TFO:                 config.XrayOptions.EnableTFO,
		} //Enable proxy protocol
		in.StreamSetting.SocketSettings = socketConfig
	}
	in.Tag = tag
	return in.Build()
}

func buildV2ray(config *conf.Options, nodeInfo *panel.NodeInfo, inbound *coreConf.InboundDetourConfig) error {
	if nodeInfo.ExtraConfig.EnableVless == "true" {
		//Set vless
		inbound.Protocol = "vless"
		if config.XrayOptions.EnableFallback {
			// Set fallback
			fallbackConfigs, err := buildVlessFallbacks(config.XrayOptions.FallBackConfigs)
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
	default:
		return errors.New("the network type is not vail")
	}
	return nil
}

func buildTrojan(config *conf.Options, inbound *coreConf.InboundDetourConfig) error {
	inbound.Protocol = "trojan"
	if config.XrayOptions.EnableFallback {
		// Set fallback
		fallbackConfigs, err := buildTrojanFallbacks(config.XrayOptions.FallBackConfigs)
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
	t := coreConf.TransportProtocol("tcp")
	inbound.StreamSetting = &coreConf.StreamConfig{Network: &t}
	return nil
}

func buildShadowsocks(config *conf.Options, nodeInfo *panel.NodeInfo, inbound *coreConf.InboundDetourConfig) error {
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
	if config.XrayOptions.DisableIVCheck {
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

func buildVlessFallbacks(fallbackConfigs []conf.FallBackConfigForXray) ([]*coreConf.VLessInboundFallback, error) {
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

func buildTrojanFallbacks(fallbackConfigs []conf.FallBackConfigForXray) ([]*coreConf.TrojanInboundFallback, error) {
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
