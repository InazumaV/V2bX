package conf

type CoreConfig struct {
	Type       string      `yaml:"Type"`
	XrayConfig *XrayConfig `yaml:"XrayConfig"`
	SingConfig *SingConfig `yaml:"SingConfig"`
}
