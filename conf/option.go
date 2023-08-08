package conf

type Options struct {
	ListenIP    string      `yaml:"ListenIP"`
	SendIP      string      `yaml:"SendIP"`
	LimitConfig LimitConfig `yaml:"LimitConfig"`
	CertConfig  *CertConfig `yaml:"CertConfig"`
	XrayOptions XrayOptions `yaml:"XrayOptions"`
	SingOptions SingOptions `yaml:"SingOptions"`
}

type XrayOptions struct {
	EnableProxyProtocol bool                    `yaml:"EnableProxyProtocol"`
	EnableDNS           bool                    `yaml:"EnableDNS"`
	DNSType             string                  `yaml:"DNSType"`
	EnableUot           bool                    `yaml:"EnableUot"`
	EnableTFO           bool                    `yaml:"EnableTFO"`
	DisableIVCheck      bool                    `yaml:"DisableIVCheck"`
	DisableSniffing     bool                    `yaml:"DisableSniffing"`
	EnableFallback      bool                    `yaml:"EnableFallback"`
	FallBackConfigs     []FallBackConfigForXray `yaml:"FallBackConfigs"`
}

type SingOptions struct {
	EnableProxyProtocol      bool                   `yaml:"EnableProxyProtocol"`
	TCPFastOpen              bool                   `yaml:"EnableTFO"`
	SniffEnabled             bool                   `yaml:"EnableSniff"`
	SniffOverrideDestination bool                   `yaml:"SniffOverrideDestination"`
	FallBackConfigs          *FallBackConfigForSing `yaml:"FallBackConfigs"`
}

type FallBackConfigForXray struct {
	SNI              string `yaml:"SNI"`
	Alpn             string `yaml:"Alpn"`
	Path             string `yaml:"Path"`
	Dest             string `yaml:"Dest"`
	ProxyProtocolVer uint64 `yaml:"ProxyProtocolVer"`
}

type FallBackConfigForSing struct {
	// sing-box
	FallBack        FallBack            `yaml:"FallBack"`
	FallBackForALPN map[string]FallBack `yaml:"FallBackForALPN"`
}
type FallBack struct {
	Server     string `yaml:"Server"`
	ServerPort string `yaml:"ServerPort"`
}
