package conf

type XrayConfig struct {
	LogConfig          *XrayLogConfig        `yaml:"Log"`
	AssetPath          string                `yaml:"AssetPath"`
	DnsConfigPath      string                `yaml:"DnsConfigPath"`
	RouteConfigPath    string                `yaml:"RouteConfigPath"`
	ConnectionConfig   *XrayConnectionConfig `yaml:"XrayConnectionConfig"`
	InboundConfigPath  string                `yaml:"InboundConfigPath"`
	OutboundConfigPath string                `yaml:"OutboundConfigPath"`
}

type XrayLogConfig struct {
	Level      string `yaml:"Level"`
	AccessPath string `yaml:"AccessPath"`
	ErrorPath  string `yaml:"ErrorPath"`
}

type XrayConnectionConfig struct {
	Handshake    uint32 `yaml:"handshake"`
	ConnIdle     uint32 `yaml:"connIdle"`
	UplinkOnly   uint32 `yaml:"uplinkOnly"`
	DownlinkOnly uint32 `yaml:"downlinkOnly"`
	BufferSize   int32  `yaml:"bufferSize"`
}

func NewXrayConfig() *XrayConfig {
	return &XrayConfig{
		LogConfig: &XrayLogConfig{
			Level:      "warning",
			AccessPath: "",
			ErrorPath:  "",
		},
		AssetPath:          "/etc/V2bX/",
		DnsConfigPath:      "",
		InboundConfigPath:  "",
		OutboundConfigPath: "",
		RouteConfigPath:    "",
		ConnectionConfig: &XrayConnectionConfig{
			Handshake:    4,
			ConnIdle:     30,
			UplinkOnly:   2,
			DownlinkOnly: 4,
			BufferSize:   64,
		},
	}
}
