package conf

type Options struct {
	ListenIP    string      `yaml:"ListenIP"`
	SendIP      string      `yaml:"SendIP"`
	LimitConfig LimitConfig `yaml:"LimitConfig"`
	CertConfig  *CertConfig `yaml:"CertConfig"`
	XrayOptions XrayOptions `yaml:"XrayOptions"`
	HyOptions   HyOptions   `yaml:"HyOptions"`
}

type XrayOptions struct {
	EnableProxyProtocol bool             `yaml:"EnableProxyProtocol"`
	EnableDNS           bool             `yaml:"EnableDNS"`
	DNSType             string           `yaml:"DNSType"`
	EnableUot           bool             `yaml:"EnableUot"`
	EnableTFO           bool             `yaml:"EnableTFO"`
	DisableIVCheck      bool             `yaml:"DisableIVCheck"`
	DisableSniffing     bool             `yaml:"DisableSniffing"`
	EnableFallback      bool             `yaml:"EnableFallback"`
	FallBackConfigs     []FallBackConfig `yaml:"FallBackConfigs"`
}

type FallBackConfig struct {
	SNI              string `yaml:"SNI"`
	Alpn             string `yaml:"Alpn"`
	Path             string `yaml:"Path"`
	Dest             string `yaml:"Dest"`
	ProxyProtocolVer uint64 `yaml:"ProxyProtocolVer"`
}

type HyOptions struct {
	Resolver          string `yaml:"Resolver"`
	ResolvePreference string `yaml:"ResolvePreference"`
	SendDevice        string `yaml:"SendDevice"`
}
