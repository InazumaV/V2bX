package node_test

import (
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	. "github.com/Yuzuki616/V2bX/node"
	"testing"
)

func TestBuildV2ray(t *testing.T) {
	nodeInfo := &panel.NodeInfo{
		NodeType:        "v2ray",
		NodeId:          1,
		ServerPort:      1145,
		Network:         "ws",
		NetworkSettings: nil,
		Host:            "test.test.tk",
		ServerName:      "test.test.tk",
	}
	certConfig := &conf.CertConfig{
		CertMode:   "none",
		CertDomain: "test.test.tk",
		Provider:   "alidns",
		Email:      "test@gmail.com",
	}
	config := &conf.ControllerConfig{
		ListenIP:   "0.0.0.0",
		CertConfig: certConfig,
	}
	_, err := BuildInbound(config, nodeInfo, "11")
	if err != nil {
		t.Error(err)
	}
}
