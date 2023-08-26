package conf

import (
	"github.com/goccy/go-json"
)

type NodeConfig struct {
	ApiConfig ApiConfig `json:"-"`
	Options   Options   `json:"-"`
}

type rawNodeConfig struct {
	ApiRaw *json.RawMessage `json:"ApiConfig"`
	OptRaw *json.RawMessage `json:"Options"`
}

type ApiConfig struct {
	APIHost      string `json:"ApiHost"`
	NodeID       int    `json:"NodeID"`
	Key          string `json:"ApiKey"`
	NodeType     string `json:"NodeType"`
	Timeout      int    `json:"Timeout"`
	RuleListPath string `json:"RuleListPath"`
}

type Options struct {
	Core        string       `json:"Core"`
	ListenIP    string       `json:"ListenIP"`
	SendIP      string       `json:"SendIP"`
	LimitConfig LimitConfig  `json:"LimitConfig"`
	XrayOptions *XrayOptions `json:"XrayOptions"`
	SingOptions *SingOptions `json:"SingOptions"`
	CertConfig  *CertConfig  `json:"CertConfig"`
}

func (n *NodeConfig) UnmarshalJSON(data []byte) (err error) {
	r := rawNodeConfig{}
	err = json.Unmarshal(data, &r)
	if err != nil {
		return err
	}
	if r.ApiRaw != nil {
		err = json.Unmarshal(*r.ApiRaw, &n.ApiConfig)
		if err != nil {
			return
		}
	} else {
		n.ApiConfig = ApiConfig{
			Timeout: 30,
		}
		err = json.Unmarshal(data, &n.ApiConfig)
		if err != nil {
			return
		}
	}
	if r.OptRaw != nil {
		data = *r.OptRaw
		err = json.Unmarshal(data, &n.Options)
		if err != nil {
			return
		}
	} else {
		n.Options = Options{
			Core:     "xray",
			ListenIP: "0.0.0.0",
			SendIP:   "0.0.0.0",
		}
		err = json.Unmarshal(data, &n.Options)
		if err != nil {
			return
		}
	}
	switch n.Options.Core {
	case "xray":
		n.Options.XrayOptions = NewXrayOptions()
		return json.Unmarshal(data, n.Options.XrayOptions)
	case "sing":
		n.Options.SingOptions = NewSingOptions()
		return json.Unmarshal(data, n.Options.SingOptions)
	}
	return
}
