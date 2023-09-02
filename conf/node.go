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
			APIHost: "http://127.0.0.1",
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
			ListenIP: "0.0.0.0",
			SendIP:   "0.0.0.0",
		}
		err = json.Unmarshal(data, &n.Options)
		if err != nil {
			return
		}
	}
	return
}

type Options struct {
	Core        string          `json:"Core"`
	ListenIP    string          `json:"ListenIP"`
	SendIP      string          `json:"SendIP"`
	LimitConfig LimitConfig     `json:"LimitConfig"`
	RawOptions  json.RawMessage `json:"RawOptions"`
	XrayOptions *XrayOptions    `json:"XrayOptions"`
	SingOptions *SingOptions    `json:"SingOptions"`
	CertConfig  *CertConfig     `json:"CertConfig"`
}

func (o *Options) UnmarshalJSON(data []byte) error {
	type opt Options
	err := json.Unmarshal(data, (*opt)(o))
	if err != nil {
		return err
	}
	switch o.Core {
	case "xray":
		o.XrayOptions = NewXrayOptions()
		return json.Unmarshal(data, o.XrayOptions)
	case "sing":
		o.SingOptions = NewSingOptions()
		return json.Unmarshal(data, o.SingOptions)
	default:
		o.RawOptions = data
	}
	return nil
}
