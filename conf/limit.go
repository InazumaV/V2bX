package conf

type LimitConfig struct {
	EnableRealtime          bool                     `yaml:"EnableRealtime"`
	SpeedLimit              int                      `yaml:"SpeedLimit"`
	IPLimit                 int                      `yaml:"DeviceLimit"`
	ConnLimit               int                      `yaml:"ConnLimit"`
	EnableIpRecorder        bool                     `yaml:"EnableIpRecorder"`
	IpRecorderConfig        *IpReportConfig          `yaml:"IpRecorderConfig"`
	EnableDynamicSpeedLimit bool                     `yaml:"EnableDynamicSpeedLimit"`
	DynamicSpeedLimitConfig *DynamicSpeedLimitConfig `yaml:"DynamicSpeedLimitConfig"`
}

type RecorderConfig struct {
	Url     string `yaml:"Url"`
	Token   string `yaml:"Token"`
	Timeout int    `yaml:"Timeout"`
}

type RedisConfig struct {
	Address  string `yaml:"Address"`
	Password string `yaml:"Password"`
	Db       int    `yaml:"Db"`
	Expiry   int    `json:"Expiry"`
}

type IpReportConfig struct {
	Periodic       int             `yaml:"Periodic"`
	Type           string          `yaml:"Type"`
	RecorderConfig *RecorderConfig `yaml:"RecorderConfig"`
	RedisConfig    *RedisConfig    `yaml:"RedisConfig"`
	EnableIpSync   bool            `yaml:"EnableIpSync"`
}

type DynamicSpeedLimitConfig struct {
	Periodic   int   `yaml:"Periodic"`
	Traffic    int64 `yaml:"Traffic"`
	SpeedLimit int   `yaml:"SpeedLimit"`
	ExpireTime int   `yaml:"ExpireTime"`
}
