package conf

type CertConfig struct {
	CertMode         string            `yaml:"CertMode"` // none, file, http, dns
	RejectUnknownSni bool              `yaml:"RejectUnknownSni"`
	CertDomain       string            `yaml:"CertDomain"`
	CertFile         string            `yaml:"CertFile"`
	KeyFile          string            `yaml:"KeyFile"`
	Provider         string            `yaml:"Provider"` // alidns, cloudflare, gandi, godaddy....
	Email            string            `yaml:"Email"`
	DNSEnv           map[string]string `yaml:"DNSEnv"`
	RealityConfig    *RealityConfig    `yaml:"RealityConfig"`
}

type RealityConfig struct {
	Dest         string   `yaml:"Dest" json:"Dest"`
	Xver         uint64   `yaml:"Xver" json:"Xver"`
	ServerNames  []string `yaml:"ServerNames" json:"ServerNames"`
	PrivateKey   string   `yaml:"PrivateKey" json:"PrivateKey"`
	MinClientVer string   `yaml:"MinClientVer" json:"MinClientVer"`
	MaxClientVer string   `yaml:"MaxClientVer" json:"MaxClientVer"`
	MaxTimeDiff  uint64   `yaml:"MaxTimeDiff" json:"MaxTimeDiff"`
	ShortIds     []string `yaml:"ShortIds" json:"ShortIds"`
}
