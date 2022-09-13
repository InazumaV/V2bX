package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

type Conf struct {
	LogConfig          *LogConfig       `yaml:"Log"`
	DnsConfigPath      string           `yaml:"DnsConfigPath"`
	InboundConfigPath  string           `yaml:"InboundConfigPath"`
	OutboundConfigPath string           `yaml:"OutboundConfigPath"`
	RouteConfigPath    string           `yaml:"RouteConfigPath"`
	ConnectionConfig   *ConnetionConfig `yaml:"ConnectionConfig"`
	NodesConfig        []*NodeConfig    `yaml:"Nodes"`
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

func (p *Conf) LoadFromPath(filePath string) error {
	confPath := path.Dir(filePath)
	os.Setenv("XRAY_LOCATION_ASSET", confPath)
	os.Setenv("XRAY_LOCATION_CONFIG", confPath)
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open config file error: %v", err)
	}
	err = yaml.NewDecoder(f).Decode(p)
	if err != nil {
		return fmt.Errorf("decode config error: %v", err)
	}
	return nil
}
