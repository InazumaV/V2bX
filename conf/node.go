package conf

type CertConfig struct {
	CertMode         string            `mapstructure:"CertMode"` // none, file, http, dns
	RejectUnknownSni bool              `mapstructure:"RejectUnknownSni"`
	CertDomain       string            `mapstructure:"CertDomain"`
	CertFile         string            `mapstructure:"CertFile"`
	KeyFile          string            `mapstructure:"KeyFile"`
	Provider         string            `mapstructure:"Provider"` // alidns, cloudflare, gandi, godaddy....
	Email            string            `mapstructure:"Email"`
	DNSEnv           map[string]string `mapstructure:"DNSEnv"`
}

type FallBackConfig struct {
	SNI              string `mapstructure:"SNI"`
	Alpn             string `mapstructure:"Alpn"`
	Path             string `mapstructure:"Path"`
	Dest             string `mapstructure:"Dest"`
	ProxyProtocolVer uint64 `mapstructure:"ProxyProtocolVer"`
}

type IpReportConfig struct {
	Url          string `mapstructure:"Url"`
	Token        string `mapstructure:"Token"`
	Periodic     int    `mapstructure:"Periodic"`
	Timeout      int    `mapstructure:"Timeout"`
	EnableIpSync bool   `mapstructure:"EnableIpSync"`
}

type DynamicSpeedLimitConfig struct {
	Periodic   int    `mapstructure:"Periodic"`
	Traffic    int64  `mapstructure:"Traffic"`
	SpeedLimit uint64 `mapstructure:"SpeedLimit"`
	ExpireTime int    `mapstructure:"ExpireTime"`
}

type ControllerConfig struct {
	ListenIP                string                   `mapstructure:"ListenIP"`
	SendIP                  string                   `mapstructure:"SendIP"`
	UpdatePeriodic          int                      `mapstructure:"UpdatePeriodic"`
	EnableDNS               bool                     `mapstructure:"EnableDNS"`
	DNSType                 string                   `mapstructure:"DNSType"`
	DisableUploadTraffic    bool                     `mapstructure:"DisableUploadTraffic"`
	DisableGetRule          bool                     `mapstructure:"DisableGetRule"`
	EnableProxyProtocol     bool                     `mapstructure:"EnableProxyProtocol"`
	EnableFallback          bool                     `mapstructure:"EnableFallback"`
	DisableIVCheck          bool                     `mapstructure:"DisableIVCheck"`
	DisableSniffing         bool                     `mapstructure:"DisableSniffing"`
	FallBackConfigs         []*FallBackConfig        `mapstructure:"FallBackConfigs"`
	EnableIpRecorder        bool                     `mapstructure:"EnableIpRecorder"`
	IpRecorderConfig        *IpReportConfig          `mapstructure:"IpRecorderConfig"`
	EnableDynamicSpeedLimit bool                     `mapstructure:"EnableDynamicSpeedLimit"`
	DynamicSpeedLimitConfig *DynamicSpeedLimitConfig `mapstructure:"DynamicSpeedLimitConfig"`
	CertConfig              *CertConfig              `mapstructure:"CertConfig"`
}

type ApiConfig struct {
	APIHost     string `mapstructure:"ApiHost"`
	NodeID      int    `mapstructure:"NodeID"`
	Key         string `mapstructure:"ApiKey"`
	NodeType    string `mapstructure:"NodeType"`
	EnableVless bool   `mapstructure:"EnableVless"`
	EnableXTLS  bool   `mapstructure:"EnableXTLS"`
	//EnableSS2022        bool    `mapstructure:"EnableSS2022"`
	Timeout             int     `mapstructure:"Timeout"`
	SpeedLimit          float64 `mapstructure:"SpeedLimit"`
	DeviceLimit         int     `mapstructure:"DeviceLimit"`
	RuleListPath        string  `mapstructure:"RuleListPath"`
	DisableCustomConfig bool    `mapstructure:"DisableCustomConfig"`
}

type NodeConfig struct {
	ApiConfig        *ApiConfig        `mapstructure:"ApiConfig"`
	ControllerConfig *ControllerConfig `mapstructure:"ControllerConfig"`
}
