package conf

type CertConfig struct {
	CertMode         string            `yaml:"CertMode"` // none, file, http, dns
	RejectUnknownSni bool              `yaml:"RejectUnknownSni"`
	CertDomain       string            `yaml:"CertDomain"`
	CertFile         string            `yaml:"CertFile"`
	KeyFile          string            `yaml:"KeyFile"`
	Provider         string            `yaml:"Provider"` // alidns, cloudflare, gandi, godaddy....
	Email            string            `yaml:"Email"`
	DNSEnv           map[string]string `yaml:"DNSEnv"`
}

type FallBackConfig struct {
	SNI              string `yaml:"SNI"`
	Alpn             string `yaml:"Alpn"`
	Path             string `yaml:"Path"`
	Dest             string `yaml:"Dest"`
	ProxyProtocolVer uint64 `yaml:"ProxyProtocolVer"`
}

type IpReportConfig struct {
	Url          string `yaml:"Url"`
	Token        string `yaml:"Token"`
	Periodic     int    `yaml:"Periodic"`
	Timeout      int    `yaml:"Timeout"`
	EnableIpSync bool   `yaml:"EnableIpSync"`
}

type DynamicSpeedLimitConfig struct {
	Periodic   int    `yaml:"Periodic"`
	Traffic    int64  `yaml:"Traffic"`
	SpeedLimit uint64 `yaml:"SpeedLimit"`
	ExpireTime int    `yaml:"ExpireTime"`
}

type ControllerConfig struct {
	ListenIP                string                   `yaml:"ListenIP"`
	SendIP                  string                   `yaml:"SendIP"`
	UpdatePeriodic          int                      `yaml:"UpdatePeriodic"`
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

type ApiConfig struct {
	APIHost     string `yaml:"ApiHost"`
	NodeID      int    `yaml:"NodeID"`
	Key         string `yaml:"ApiKey"`
	NodeType    string `yaml:"NodeType"`
	EnableVless bool   `yaml:"EnableVless"`
	EnableXTLS  bool   `yaml:"EnableXTLS"`
	//EnableSS2022        bool    `yaml:"EnableSS2022"`
	Timeout             int     `yaml:"Timeout"`
	SpeedLimit          float64 `yaml:"SpeedLimit"`
	DeviceLimit         int     `yaml:"DeviceLimit"`
	RuleListPath        string  `yaml:"RuleListPath"`
	DisableCustomConfig bool    `yaml:"DisableCustomConfig"`
}

type NodeConfig struct {
	ApiConfig        *ApiConfig        `yaml:"ApiConfig"`
	ControllerConfig *ControllerConfig `yaml:"ControllerConfig"`
}
