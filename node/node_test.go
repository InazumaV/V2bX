package node_test

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core"
	_ "github.com/Yuzuki616/V2bX/core/distro/all"
	. "github.com/Yuzuki616/V2bX/node"
	xCore "github.com/xtls/xray-core/core"
	coreConf "github.com/xtls/xray-core/infra/conf"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"testing"
)

func TestController(t *testing.T) {
	serverConfig := &coreConf.Config{
		Stats:     &coreConf.StatsConfig{},
		LogConfig: &coreConf.LogConfig{LogLevel: "debug"},
	}
	policyConfig := &coreConf.PolicyConfig{}
	policyConfig.Levels = map[uint32]*coreConf.Policy{0: &coreConf.Policy{
		StatsUserUplink:   true,
		StatsUserDownlink: true,
	}}
	serverConfig.Policy = policyConfig
	config, _ := serverConfig.Build()

	// config := &core.Config{
	// 	App: []*serial.TypedMessage{
	// 		serial.ToTypedMessage(&dispatcher.Config{}),
	// 		serial.ToTypedMessage(&proxyman.InboundConfig{}),
	// 		serial.ToTypedMessage(&proxyman.OutboundConfig{}),
	// 		serial.ToTypedMessage(&stats.Config{}),
	// 	}}

	server, err := xCore.New(config)
	defer server.Close()
	if err != nil {
		t.Errorf("failed to create instance: %s", err)
	}
	if err = server.Start(); err != nil {
		t.Errorf("Failed to start instance: %s", err)
	}
	certConfig := &conf.CertConfig{
		CertMode:   "http",
		CertDomain: "test.ss.tk",
		Provider:   "alidns",
		Email:      "ss@ss.com",
	}
	controlerconfig := &conf.ControllerConfig{
		UpdatePeriodic: 5,
		CertConfig:     certConfig,
	}
	apiConfig := &conf.ApiConfig{
		APIHost:  "http://127.0.0.1:667",
		Key:      "123",
		NodeID:   41,
		NodeType: "V2ray",
	}
	apiclient := panel.New(apiConfig)
	c := &core.Core{Server: server}
	c.Start()
	node := New(c, apiclient, controlerconfig)
	fmt.Println("Sleep 1s")
	err = node.Start()
	if err != nil {
		t.Error(err)
	}
	//Explicitly triggering GC to remove garbage from config loading.
	runtime.GC()
	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
		<-osSignals
	}
}
