package conf

type ConnectionConfig struct {
	Handshake    uint32 `yaml:"handshake"`
	ConnIdle     uint32 `yaml:"connIdle"`
	UplinkOnly   uint32 `yaml:"uplinkOnly"`
	DownlinkOnly uint32 `yaml:"downlinkOnly"`
	BufferSize   int32  `yaml:"bufferSize"`
}

func NewConnectionConfig() *ConnectionConfig {
	return &ConnectionConfig{
		Handshake:    4,
		ConnIdle:     30,
		UplinkOnly:   2,
		DownlinkOnly: 4,
		BufferSize:   64,
	}
}
