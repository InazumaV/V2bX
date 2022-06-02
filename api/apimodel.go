package api

import (
	"github.com/xtls/xray-core/infra/conf"
	"regexp"
)

// API config
type Config struct {
	APIHost             string  `mapstructure:"ApiHost"`
	NodeID              int     `mapstructure:"NodeID"`
	Key                 string  `mapstructure:"ApiKey"`
	NodeType            string  `mapstructure:"NodeType"`
	EnableVless         bool    `mapstructure:"EnableVless"`
	EnableXTLS          bool    `mapstructure:"EnableXTLS"`
	EnableSS2022        bool    `mapstructure:"EnableSS2022"`
	Timeout             int     `mapstructure:"Timeout"`
	SpeedLimit          float64 `mapstructure:"SpeedLimit"`
	DeviceLimit         int     `mapstructure:"DeviceLimit"`
	RuleListPath        string  `mapstructure:"RuleListPath"`
	DisableCustomConfig bool    `mapstructure:"DisableCustomConfig"`
}

type OnlineUser struct {
	UID int
	IP  string
}

type UserTraffic struct {
	UID      int
	Email    string
	Upload   int64
	Download int64
}

type ClientInfo struct {
	APIHost  string
	NodeID   int
	Key      string
	NodeType string
}

type DetectRule struct {
	ID      int
	Pattern *regexp.Regexp
}

type DetectResult struct {
	UID    int
	RuleID int
}

type V2RayUserInfo struct {
	Uuid    string `json:"uuid"`
	Email   string `json:"email"`
	AlterId int    `json:"alter_id"`
}
type TrojanUserInfo struct {
	Password string `json:"password"`
}
type UserInfo struct {
	DeviceLimit int             `json:"device_limit"`
	SpeedLimit  uint64          `json:"speed_limit"`
	UID         int             `json:"id"`
	Port        int             `json:"port"`
	Cipher      string          `json:"cipher"`
	Secret      string          `json:"secret"`
	V2rayUser   *V2RayUserInfo  `json:"v2ray_user"`
	TrojanUser  *TrojanUserInfo `json:"trojan_user"`
}
type UserListBody struct {
	//Msg  string `json:"msg"`
	Data []UserInfo `json:"data"`
}

func (p *UserInfo) GetUserEmail() string {
	if p.V2rayUser != nil {
		return p.V2rayUser.Email
	} else if p.TrojanUser != nil {
		return p.TrojanUser.Password
	}
	return p.Cipher
}

type NodeInfo struct {
	RspMd5       string
	NodeType     string
	NodeId       int
	TLSType      string
	EnableVless  bool
	EnableTls    bool
	EnableSS2022 bool
	V2ray        *V2rayConfig
	Trojan       *TrojanConfig
	SS           *SSConfig
}

type SSConfig struct {
	Port              int    `json:"port"`
	TransportProtocol string `json:"transportProtocol"`
	CypherMethod      string `json:"cypher"`
}
type V2rayConfig struct {
	Inbounds []conf.InboundDetourConfig `json:"inbounds"`
	Routing  *struct {
		Rules []Rule `json:"rules"`
	} `json:"routing"`
}

type Rule struct {
	Type        string   `json:"type"`
	InboundTag  string   `json:"inboundTag,omitempty"`
	OutboundTag string   `json:"outboundTag"`
	Domain      []string `json:"domain,omitempty"`
	Protocol    []string `json:"protocol,omitempty"`
}

type TrojanConfig struct {
	LocalPort         int           `json:"local_port"`
	Password          []interface{} `json:"password"`
	TransportProtocol string
	Ssl               struct {
		Sni string `json:"sni"`
	} `json:"ssl"`
}
