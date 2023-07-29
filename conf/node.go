package conf

type NodeConfig struct {
	ApiConfig *ApiConfig `yaml:"ApiConfig"`
	Options   *Options   `yaml:"Options"`
}

type ApiConfig struct {
	APIHost      string `yaml:"ApiHost"`
	NodeID       int    `yaml:"NodeID"`
	Key          string `yaml:"ApiKey"`
	NodeType     string `yaml:"NodeType"`
	Timeout      int    `yaml:"Timeout"`
	RuleListPath string `yaml:"RuleListPath"`
}
