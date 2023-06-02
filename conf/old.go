package conf

import "log"

type OldConfig struct {
	NodesConfig []*struct {
		ApiConfig        *OldApiConfig        `yaml:"ApiConfig"`
		ControllerConfig *OldControllerConfig `yaml:"ControllerConfig"`
	} `yaml:"Nodes"`
}

type OldControllerConfig struct {
	ListenIP                string                   `yaml:"ListenIP"`
	SendIP                  string                   `yaml:"SendIP"`
	EnableDNS               bool                     `yaml:"EnableDNS"`
	DNSType                 string                   `yaml:"DNSType"`
	DisableUploadTraffic    bool                     `yaml:"DisableUploadTraffic"`
	DisableGetRule          bool                     `yaml:"DisableGetRule"`
	EnableProxyProtocol     bool                     `yaml:"EnableProxyProtocol"`
	EnableFallback          bool                     `yaml:"EnableFallback"`
	DisableIVCheck          bool                     `yaml:"DisableIVCheck"`
	DisableSniffing         bool                     `yaml:"DisableSniffing"`
	FallBackConfigs         []*FallBackConfig        `yaml:"FallBackConfigs"`
	EnableIpRecorder        bool                     `yaml:"EnableIpRecorder"`
	IpRecorderConfig        *IpReportConfig          `yaml:"IpRecorderConfig"`
	EnableDynamicSpeedLimit bool                     `yaml:"EnableDynamicSpeedLimit"`
	DynamicSpeedLimitConfig *DynamicSpeedLimitConfig `yaml:"DynamicSpeedLimitConfig"`
	CertConfig              *CertConfig              `yaml:"CertConfig"`
}

type OldApiConfig struct {
	APIHost             string `yaml:"ApiHost"`
	NodeID              int    `yaml:"NodeID"`
	Key                 string `yaml:"ApiKey"`
	NodeType            string `yaml:"NodeType"`
	EnableVless         bool   `yaml:"EnableVless"`
	Timeout             int    `yaml:"Timeout"`
	SpeedLimit          int    `yaml:"SpeedLimit"`
	DeviceLimit         int    `yaml:"DeviceLimit"`
	RuleListPath        string `yaml:"RuleListPath"`
	DisableCustomConfig bool   `yaml:"DisableCustomConfig"`
}

func migrateOldConfig(c *Conf, old *OldConfig) {
	changed := false
	for i, n := range c.NodesConfig {
		if i >= len(old.NodesConfig) {
			break
		}
		// node option
		if old.NodesConfig[i].ApiConfig.EnableVless {
			n.ControllerConfig.EnableVless = true
			changed = true
		}
		// limit config
		if old.NodesConfig[i].ApiConfig.SpeedLimit != 0 {
			n.ControllerConfig.LimitConfig.SpeedLimit = old.NodesConfig[i].ApiConfig.SpeedLimit
			changed = true
		}
		if old.NodesConfig[i].ApiConfig.DeviceLimit != 0 {
			n.ControllerConfig.LimitConfig.IPLimit = old.NodesConfig[i].ApiConfig.DeviceLimit
			changed = true
		}
		if old.NodesConfig[i].ControllerConfig.EnableDynamicSpeedLimit {
			n.ControllerConfig.LimitConfig.EnableDynamicSpeedLimit = true
			changed = true
		}
		if old.NodesConfig[i].ControllerConfig.DynamicSpeedLimitConfig != nil {
			n.ControllerConfig.LimitConfig.DynamicSpeedLimitConfig =
				old.NodesConfig[i].ControllerConfig.DynamicSpeedLimitConfig
			changed = true
		}
		if old.NodesConfig[i].ControllerConfig.EnableIpRecorder {
			n.ControllerConfig.LimitConfig.EnableIpRecorder = true
			changed = true
		}
		if old.NodesConfig[i].ControllerConfig.IpRecorderConfig != nil {
			n.ControllerConfig.LimitConfig.IpRecorderConfig =
				old.NodesConfig[i].ControllerConfig.IpRecorderConfig
			changed = true
		}
	}
	if changed {
		log.Println("Warning: This config file is old.")
	}
}
