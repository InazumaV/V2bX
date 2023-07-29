package conf

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Conf struct {
	CoreConfig  CoreConfig    `yaml:"CoreConfig"`
	NodesConfig []*NodeConfig `yaml:"Nodes"`
}

func New() *Conf {
	return &Conf{
		CoreConfig: CoreConfig{
			Type:       "xray",
			XrayConfig: NewXrayConfig(),
			SingConfig: NewSingConfig(),
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
	return nil
}
