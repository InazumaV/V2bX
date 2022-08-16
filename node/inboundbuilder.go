// Package node the InbounderConfig used by add inbound
package node

import (
	"encoding/json"
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/node/legoCmd"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/uuid"
	"github.com/xtls/xray-core/core"
	coreConf "github.com/xtls/xray-core/infra/conf"
)

//InboundBuilder build Inbound config for different protocol
func InboundBuilder(config *conf.ControllerConfig, nodeInfo *panel.NodeInfo, tag string) (*core.InboundHandlerConfig, error) {
	var proxySetting interface{}
	if nodeInfo.NodeType == "V2ray" {
		defer func() {
			nodeInfo.V2ray = nil
		}()
		if nodeInfo.EnableVless {
			nodeInfo.V2ray.Inbounds[0].Protocol = "vless"
			// Enable fallback
			if config.EnableFallback {
				fallbackConfigs, err := buildVlessFallbacks(config.FallBackConfigs)
				if err == nil {
					proxySetting = &coreConf.VLessInboundConfig{
						Decryption: "none",
						Fallbacks:  fallbackConfigs,
					}
				} else {
					return nil, err
				}
			} else {
				proxySetting = &coreConf.VLessInboundConfig{
					Decryption: "none",
				}
			}
		} else {
			nodeInfo.V2ray.Inbounds[0].Protocol = "vmess"
			proxySetting = &coreConf.VMessInboundConfig{}
		}
	} else if nodeInfo.NodeType == "Trojan" {
		defer func() {
			nodeInfo.V2ray = nil
			nodeInfo.Trojan = nil
		}()
		nodeInfo.V2ray = &panel.V2rayConfig{}
		nodeInfo.V2ray.Inbounds = make([]coreConf.InboundDetourConfig, 1)
		nodeInfo.V2ray.Inbounds[0].Protocol = "trojan"
		// Enable fallback
		if config.EnableFallback {
			fallbackConfigs, err := buildTrojanFallbacks(config.FallBackConfigs)
			if err == nil {
				proxySetting = &coreConf.TrojanServerConfig{
					Fallbacks: fallbackConfigs,
				}
			} else {
				return nil, err
			}
		} else {
			proxySetting = &coreConf.TrojanServerConfig{}
		}
		nodeInfo.V2ray.Inbounds[0].PortList = &coreConf.PortList{
			Range: []coreConf.PortRange{{From: uint32(nodeInfo.Trojan.LocalPort), To: uint32(nodeInfo.Trojan.LocalPort)}},
		}
		t := coreConf.TransportProtocol(nodeInfo.Trojan.TransportProtocol)
		nodeInfo.V2ray.Inbounds[0].StreamSetting = &coreConf.StreamConfig{Network: &t}
	} else if nodeInfo.NodeType == "Shadowsocks" {
		defer func() {
			nodeInfo.V2ray = nil
		}()
		nodeInfo.V2ray = &panel.V2rayConfig{}
		nodeInfo.V2ray.Inbounds = []coreConf.InboundDetourConfig{{Protocol: "shadowsocks"}}
		proxySetting = &coreConf.ShadowsocksServerConfig{}
		randomPasswd := uuid.New()
		defaultSSuser := &coreConf.ShadowsocksUserConfig{
			Cipher:   "aes-128-gcm",
			Password: randomPasswd.String(),
		}
		proxySetting, _ := proxySetting.(*coreConf.ShadowsocksServerConfig)
		proxySetting.Users = append(proxySetting.Users, defaultSSuser)
		proxySetting.NetworkList = &coreConf.NetworkList{"tcp", "udp"}
		proxySetting.IVCheck = true
		if config.DisableIVCheck {
			proxySetting.IVCheck = false
		}
		nodeInfo.V2ray.Inbounds[0].PortList = &coreConf.PortList{
			Range: []coreConf.PortRange{{From: uint32(nodeInfo.SS.Port), To: uint32(nodeInfo.SS.Port)}},
		}
		t := coreConf.TransportProtocol(nodeInfo.SS.TransportProtocol)
		nodeInfo.V2ray.Inbounds[0].StreamSetting = &coreConf.StreamConfig{Network: &t}
	} else {
		return nil, fmt.Errorf("unsupported node type: %s, Only support: V2ray, Trojan, Shadowsocks", nodeInfo.NodeType)
	} /*else if nodeInfo.NodeType == "dokodemo-door" {
		nodeInfo.V2ray = &api.V2rayConfig{}
		nodeInfo.V2ray.Inbounds = make([]coreConf.InboundDetourConfig, 1)
		nodeInfo.V2ray.Inbounds[0].Protocol = "dokodemo-door"
		proxySetting = struct {
			Host        string   `json:"address"`
			NetworkList []string `json:"network"`
		}{
			Host:        "v1.mux.cool",
			NetworkList: []string{"tcp", "udp"},
		}
	}*/
	// Build Listen IP address
	ipAddress := net.ParseAddress(config.ListenIP)
	nodeInfo.V2ray.Inbounds[0].ListenOn = &coreConf.Address{Address: ipAddress}
	// SniffingConfig
	sniffingConfig := &coreConf.SniffingConfig{
		Enabled:      true,
		DestOverride: &coreConf.StringList{"http", "tls"},
	}
	if config.DisableSniffing {
		sniffingConfig.Enabled = false
	}
	nodeInfo.V2ray.Inbounds[0].SniffingConfig = sniffingConfig

	var setting json.RawMessage

	// Build Protocol and Protocol setting

	setting, err := json.Marshal(proxySetting)
	if err != nil {
		return nil, fmt.Errorf("marshal proxy %s config fialed: %s", nodeInfo.NodeType, err)
	}
	if *nodeInfo.V2ray.Inbounds[0].StreamSetting.Network == "tcp" {
		if nodeInfo.V2ray.Inbounds[0].StreamSetting.TCPSettings != nil {
			nodeInfo.V2ray.Inbounds[0].StreamSetting.TCPSettings.AcceptProxyProtocol = config.EnableProxyProtocol
		} else {
			tcpSetting := &coreConf.TCPConfig{
				AcceptProxyProtocol: config.EnableProxyProtocol,
			}
			nodeInfo.V2ray.Inbounds[0].StreamSetting.TCPSettings = tcpSetting
		}
	} else if *nodeInfo.V2ray.Inbounds[0].StreamSetting.Network == "ws" {
		nodeInfo.V2ray.Inbounds[0].StreamSetting.WSSettings = &coreConf.WebSocketConfig{
			AcceptProxyProtocol: config.EnableProxyProtocol}
	}
	// Build TLS and XTLS settings
	if nodeInfo.EnableTls && config.CertConfig.CertMode != "none" {
		nodeInfo.V2ray.Inbounds[0].StreamSetting.Security = nodeInfo.TLSType
		certFile, keyFile, err := getCertFile(config.CertConfig)
		if err != nil {
			return nil, err
		}
		if nodeInfo.TLSType == "tls" {
			tlsSettings := &coreConf.TLSConfig{
				RejectUnknownSNI: config.CertConfig.RejectUnknownSni,
			}
			tlsSettings.Certs = append(tlsSettings.Certs, &coreConf.TLSCertConfig{CertFile: certFile, KeyFile: keyFile, OcspStapling: 3600})

			nodeInfo.V2ray.Inbounds[0].StreamSetting.TLSSettings = tlsSettings
		} else if nodeInfo.TLSType == "xtls" {
			xtlsSettings := &coreConf.XTLSConfig{
				RejectUnknownSNI: config.CertConfig.RejectUnknownSni,
			}
			xtlsSettings.Certs = append(xtlsSettings.Certs, &coreConf.XTLSCertConfig{
				CertFile:     certFile,
				KeyFile:      keyFile,
				OcspStapling: 3600})
			nodeInfo.V2ray.Inbounds[0].StreamSetting.XTLSSettings = xtlsSettings
		}
	}
	// Support ProxyProtocol for any transport protocol
	if *nodeInfo.V2ray.Inbounds[0].StreamSetting.Network != "tcp" &&
		*nodeInfo.V2ray.Inbounds[0].StreamSetting.Network != "ws" &&
		config.EnableProxyProtocol {
		sockoptConfig := &coreConf.SocketConfig{
			AcceptProxyProtocol: config.EnableProxyProtocol,
		}
		nodeInfo.V2ray.Inbounds[0].StreamSetting.SocketSettings = sockoptConfig
	}
	nodeInfo.V2ray.Inbounds[0].Settings = &setting
	nodeInfo.V2ray.Inbounds[0].Tag = tag
	return nodeInfo.V2ray.Inbounds[0].Build()
}

func getCertFile(certConfig *conf.CertConfig) (certFile string, keyFile string, err error) {
	if certConfig.CertMode == "file" {
		if certConfig.CertFile == "" || certConfig.KeyFile == "" {
			return "", "", fmt.Errorf("cert file path or key file path not exist")
		}
		return certConfig.CertFile, certConfig.KeyFile, nil
	} else if certConfig.CertMode == "dns" {
		lego, err := legoCmd.New()
		if err != nil {
			return "", "", err
		}
		certPath, keyPath, err := lego.DNSCert(certConfig.CertDomain, certConfig.Email, certConfig.Provider, certConfig.DNSEnv)
		if err != nil {
			return "", "", err
		}
		return certPath, keyPath, err
	} else if certConfig.CertMode == "http" {
		lego, err := legoCmd.New()
		if err != nil {
			return "", "", err
		}
		certPath, keyPath, err := lego.HTTPCert(certConfig.CertDomain, certConfig.Email)
		if err != nil {
			return "", "", err
		}
		return certPath, keyPath, err
	}

	return "", "", fmt.Errorf("unsupported certmode: %s", certConfig.CertMode)
}

func buildVlessFallbacks(fallbackConfigs []*conf.FallBackConfig) ([]*coreConf.VLessInboundFallback, error) {
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

func buildTrojanFallbacks(fallbackConfigs []*conf.FallBackConfig) ([]*coreConf.TrojanInboundFallback, error) {
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
