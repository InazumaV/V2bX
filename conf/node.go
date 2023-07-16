package conf

type NodeConfig struct {
	ApiConfig        *ApiConfig        `yaml:"ApiConfig"`
	ControllerConfig *ControllerConfig `yaml:"ControllerConfig"`
}

type ApiConfig struct {
	APIHost      string `yaml:"ApiHost"`
	NodeID       int    `yaml:"NodeID"`
	Key          string `yaml:"ApiKey"`
	NodeType     string `yaml:"NodeType"`
	Timeout      int    `yaml:"Timeout"`
	RuleListPath string `yaml:"RuleListPath"`
}

type ControllerConfig struct {
	ListenIP    string      `yaml:"ListenIP"`
	SendIP      string      `yaml:"SendIP"`
	XrayOptions XrayOptions `yaml:"XrayOptions"`
	HyOptions   HyOptions   `yaml:"HyOptions"`
	LimitConfig LimitConfig `yaml:"LimitConfig"`
	CertConfig  *CertConfig `yaml:"CertConfig"`
}

type RealityConfig struct {
	Dest         interface{} `yaml:"Dest" json:"Dest"`
	Xver         uint64      `yaml:"Xver" json:"Xver"`
	ServerNames  []string    `yaml:"ServerNames" json:"ServerNames"`
	PrivateKey   string      `yaml:"PrivateKey" json:"PrivateKey"`
	MinClientVer string      `yaml:"MinClientVer" json:"MinClientVer"`
	MaxClientVer string      `yaml:"MaxClientVer" json:"MaxClientVer"`
	MaxTimeDiff  uint64      `yaml:"MaxTimeDiff" json:"MaxTimeDiff"`
	ShortIds     []string    `yaml:"ShortIds" json:"ShortIds"`
}

type XrayOptions struct {
	EnableProxyProtocol bool             `yaml:"EnableProxyProtocol"`
	EnableDNS           bool             `yaml:"EnableDNS"`
	DNSType             string           `yaml:"DNSType"`
	EnableUot           bool             `yaml:"EnableUot"`
	EnableTFO           bool             `yaml:"EnableTFO"`
	EnableVless         bool             `yaml:"EnableVless"`
	DisableIVCheck      bool             `yaml:"DisableIVCheck"`
	DisableSniffing     bool             `yaml:"DisableSniffing"`
	EnableFallback      bool             `yaml:"EnableFallback"`
	FallBackConfigs     []FallBackConfig `yaml:"FallBackConfigs"`
}

type HyOptions struct {
	Resolver          string `yaml:"Resolver"`
	ResolvePreference string `yaml:"ResolvePreference"`
	SendDevice        string `yaml:"SendDevice"`
}

type LimitConfig struct {
	EnableRealtime          bool                     `yaml:"EnableRealtime"`
	SpeedLimit              int                      `yaml:"SpeedLimit"`
	IPLimit                 int                      `yaml:"DeviceLimit"`
	ConnLimit               int                      `yaml:"ConnLimit"`
	EnableIpRecorder        bool                     `yaml:"EnableIpRecorder"`
	IpRecorderConfig        *IpReportConfig          `yaml:"IpRecorderConfig"`
	EnableDynamicSpeedLimit bool                     `yaml:"EnableDynamicSpeedLimit"`
	DynamicSpeedLimitConfig *DynamicSpeedLimitConfig `yaml:"DynamicSpeedLimitConfig"`
}

type FallBackConfig struct {
	SNI              string `yaml:"SNI"`
	Alpn             string `yaml:"Alpn"`
	Path             string `yaml:"Path"`
	Dest             string `yaml:"Dest"`
	ProxyProtocolVer uint64 `yaml:"ProxyProtocolVer"`
}

type RecorderConfig struct {
	Url     string `yaml:"Url"`
	Token   string `yaml:"Token"`
	Timeout int    `yaml:"Timeout"`
}

type RedisConfig struct {
	Address  string `yaml:"Address"`
	Password string `yaml:"Password"`
	Db       int    `yaml:"Db"`
	Expiry   int    `json:"Expiry"`
}

type IpReportConfig struct {
	Periodic       int             `yaml:"Periodic"`
	Type           string          `yaml:"Type"`
	RecorderConfig *RecorderConfig `yaml:"RecorderConfig"`
	RedisConfig    *RedisConfig    `yaml:"RedisConfig"`
	EnableIpSync   bool            `yaml:"EnableIpSync"`
}

type DynamicSpeedLimitConfig struct {
	Periodic   int   `yaml:"Periodic"`
	Traffic    int64 `yaml:"Traffic"`
	SpeedLimit int   `yaml:"SpeedLimit"`
	ExpireTime int   `yaml:"ExpireTime"`
}

type CertConfig struct {
	CertMode         string            `yaml:"CertMode"` // none, file, http, dns
	RejectUnknownSni bool              `yaml:"RejectUnknownSni"`
	CertDomain       string            `yaml:"CertDomain"`
	CertFile         string            `yaml:"CertFile"`
	KeyFile          string            `yaml:"KeyFile"`
	Provider         string            `yaml:"Provider"` // alidns, cloudflare, gandi, godaddy....
	Email            string            `yaml:"Email"`
	DNSEnv           map[string]string `yaml:"DNSEnv"`
	RealityConfig    *RealityConfig    `yaml:"RealityConfig"`
}
