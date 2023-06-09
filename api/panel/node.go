package panel

import (
	"fmt"
	"github.com/goccy/go-json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CommonNodeRsp struct {
	Host       string     `json:"host"`
	ServerPort int        `json:"server_port"`
	ServerName string     `json:"server_name"`
	Routes     []Route    `json:"routes"`
	BaseConfig BaseConfig `json:"base_config"`
}

type Route struct {
	Id     int         `json:"id"`
	Match  interface{} `json:"match"`
	Action string      `json:"action"`
	//ActionValue interface{} `json:"action_value"`
}
type BaseConfig struct {
	PushInterval any `json:"push_interval"`
	PullInterval any `json:"pull_interval"`
}

type V2rayNodeRsp struct {
	Tls             int             `json:"tls"`
	Network         string          `json:"network"`
	NetworkSettings json.RawMessage `json:"networkSettings"`
	ServerName      string          `json:"server_name"`
}

type ShadowsocksNodeRsp struct {
	Cipher    string `json:"cipher"`
	ServerKey string `json:"server_key"`
}

type TrojanNodeRsp struct {
	Host       string `json:"host"`
	ServerName string `json:"server_name"`
}

type HysteriaNodeRsp struct {
	UpMbps   int    `json:"up_mbps"`
	DownMbps int    `json:"down_mbps"`
	Obfs     string `json:"obfs"`
}

type NodeInfo struct {
	Id              int
	Type            string
	Rules           []*regexp.Regexp
	Host            string
	Port            int
	Network         string
	NetworkSettings json.RawMessage
	Tls             bool
	ServerName      string
	UpMbps          int
	DownMbps        int
	ServerKey       string
	Cipher          string
	HyObfs          string
	PushInterval    time.Duration
	PullInterval    time.Duration
}

func (c *Client) GetNodeInfo() (node *NodeInfo, err error) {
	const path = "/api/v1/server/UniProxy/config"
	r, err := c.client.R().Get(path)
	if err = c.checkResponse(r, path, err); err != nil {
		return
	}
	if c.etag == r.Header().Get("ETag") { // node info not changed
		return nil, nil
	}
	// parse common params
	node = &NodeInfo{
		Id:   c.NodeId,
		Type: c.NodeType,
	}
	common := CommonNodeRsp{}
	err = json.Unmarshal(r.Body(), &common)
	if err != nil {
		return nil, fmt.Errorf("decode common params error: %s", err)
	}
	for i := range common.Routes { // parse rules from routes
		if common.Routes[i].Action == "block" {
			var matchs []string
			if _, ok := common.Routes[i].Match.(string); ok {
				matchs = strings.Split(common.Routes[i].Match.(string), ",")
			} else {
				matchs = common.Routes[i].Match.([]string)
			}
			for _, v := range matchs {
				node.Rules = append(node.Rules, regexp.MustCompile(v))
			}
		}
	}
	node.ServerName = common.ServerName
	node.Host = common.Host
	node.Port = common.ServerPort
	node.PullInterval = intervalToTime(common.BaseConfig.PullInterval)
	node.PushInterval = intervalToTime(common.BaseConfig.PushInterval)
	// parse protocol params
	switch c.NodeType {
	case "v2ray":
		rsp := V2rayNodeRsp{}
		err = json.Unmarshal(r.Body(), &rsp)
		if err != nil {
			return nil, fmt.Errorf("decode v2ray params error: %s", err)
		}
		node.Network = rsp.Network
		node.NetworkSettings = rsp.NetworkSettings
		node.ServerName = rsp.ServerName
		if rsp.Tls == 1 {
			node.Tls = true
		}
	case "shadowsocks":
		rsp := ShadowsocksNodeRsp{}
		err = json.Unmarshal(r.Body(), &rsp)
		if err != nil {
			return nil, fmt.Errorf("decode v2ray params error: %s", err)
		}
		node.ServerKey = rsp.ServerKey
		node.Cipher = rsp.Cipher
	case "trojan":
		rsp := TrojanNodeRsp{}
		err = json.Unmarshal(r.Body(), &rsp)
		if err != nil {
			return nil, fmt.Errorf("decode v2ray params error: %s", err)
		}
	case "hysteria":
		rsp := HysteriaNodeRsp{}
		err = json.Unmarshal(r.Body(), &rsp)
		if err != nil {
			return nil, fmt.Errorf("decode v2ray params error: %s", err)
		}
		node.DownMbps = rsp.DownMbps
		node.UpMbps = rsp.UpMbps
		node.HyObfs = rsp.Obfs
	}
	c.etag = r.Header().Get("Etag")
	return
}

func intervalToTime(i interface{}) time.Duration {
	switch reflect.TypeOf(i).Kind() {
	case reflect.Int:
		return time.Duration(i.(int)) * time.Second
	case reflect.String:
		i, _ := strconv.Atoi(i.(string))
		return time.Duration(i) * time.Second
	case reflect.Float64:
		return time.Duration(i.(float64)) * time.Second
	default:
		return time.Duration(reflect.ValueOf(i).Int()) * time.Second
	}
}
