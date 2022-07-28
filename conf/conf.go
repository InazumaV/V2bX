package conf

type Conf struct {
	LogConfig          *LogConfig       `mapstructure:"Log"`
	DnsConfigPath      string           `mapstructure:"DnsConfigPath"`
	InboundConfigPath  string           `mapstructure:"InboundConfigPath"`
	OutboundConfigPath string           `mapstructure:"OutboundConfigPath"`
	RouteConfigPath    string           `mapstructure:"RouteConfigPath"`
	ConnectionConfig   *ConnetionConfig `mapstructure:"ConnectionConfig"`
	NodesConfig        []*NodeConfig    `mapstructure:"Nodes"`
}

func New() *Conf {
	return &Conf{
		LogConfig:          NewLogConfig(),
		DnsConfigPath:      "",
		InboundConfigPath:  "",
		OutboundConfigPath: "",
		RouteConfigPath:    "",
		ConnectionConfig:   NewConnetionConfig(),
		NodesConfig:        []*NodeConfig{},
	}
}
