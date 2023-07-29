package conf

type SingConfig struct {
	LogConfig    SingLogConfig `yaml:"LogConfig"`
	OriginalPath string        `yaml:"OriginalPath"`
}

type SingLogConfig struct {
	Disabled  bool   `yaml:"Disable"`
	Level     string `yaml:"Level"`
	Output    string `yaml:"Output"`
	Timestamp bool   `yaml:"Timestamp"`
}

func NewSingConfig() *SingConfig {
	return &SingConfig{
		LogConfig: SingLogConfig{
			Level:     "error",
			Timestamp: true,
		},
	}
}
