package conf

import "github.com/goccy/go-json"

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
	ListenIP             string           `yaml:"ListenIP"`
	SendIP               string           `yaml:"SendIP"`
	EnableDNS            bool             `yaml:"EnableDNS"`
	DNSType              string           `yaml:"DNSType"`
	EnableVless          bool             `yaml:"EnableVless"`
	EnableXtls           bool             `json:"EnableXtls"`
	LimitConfig          LimitConfig      `yaml:"LimitConfig"`
	DisableUploadTraffic bool             `yaml:"DisableUploadTraffic"`
	DisableGetRule       bool             `yaml:"DisableGetRule"`
	EnableProxyProtocol  bool             `yaml:"EnableProxyProtocol"`
	DisableIVCheck       bool             `yaml:"DisableIVCheck"`
	DisableSniffing      bool             `yaml:"DisableSniffing"`
	EnableReality        bool             `yaml:"EnableReality"`
	RealityConfig        RealityConfig    `yaml:"RealityConfig"`
	EnableFallback       bool             `yaml:"EnableFallback"`
	FallBackConfigs      []FallBackConfig `yaml:"FallBackConfigs"`
	CertConfig           *CertConfig      `yaml:"CertConfig"`
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
}

type RealityConfig struct {
	Dest         json.RawMessage `yaml:"Dest"`
	Xver         uint64          `yaml:"Xver"`
	ServerNames  []string        `yaml:"ServerNames"`
	PrivateKey   string          `yaml:"PrivateKey"`
	MinClientVer string          `yaml:"MinClientVer"`
	MaxClientVer string          `yaml:"MaxClientVer"`
	MaxTimeDiff  uint64          `yaml:"MaxTimeDiff"`
	ShortIds     []string        `yaml:"ShortIds"`
}
