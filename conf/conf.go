package conf

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

type Conf struct {
	CoreConfig  CoreConfig    `yaml:"CoreConfig"`
	NodesConfig []*NodeConfig `yaml:"Nodes"`
}

func New() *Conf {
	return &Conf{
		CoreConfig: CoreConfig{
			Type: "xray",
			XrayConfig: &XrayConfig{
				LogConfig:          NewLogConfig(),
				AssetPath:          "/etc/V2bX/",
				DnsConfigPath:      "",
				InboundConfigPath:  "",
				OutboundConfigPath: "",
				RouteConfigPath:    "",
				ConnectionConfig:   NewConnectionConfig(),
			},
		},
		NodesConfig: []*NodeConfig{},
	}
}

func (p *Conf) LoadFromPath(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open config file error: %s", err)
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read file error: %s", err)
	}
	err = yaml.Unmarshal(content, p)
	if err != nil {
		return fmt.Errorf("decode config error: %s", err)
	}
	old := &OldConfig{}
	err = yaml.Unmarshal(content, old)
	if err == nil {
		migrateOldConfig(p, old)
	}
	return nil
}
