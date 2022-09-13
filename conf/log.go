package conf

type LogConfig struct {
	Level      string `yaml:"Level"`
	AccessPath string `yaml:"AccessPath"`
	ErrorPath  string `yaml:"ErrorPath"`
}

func NewLogConfig() *LogConfig {
	return &LogConfig{
		Level:      "none",
		AccessPath: "",
		ErrorPath:  "",
	}
}
