package core

import (
	"encoding/json"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core/app/dispatcher"
	_ "github.com/Yuzuki616/V2bX/core/distro/all"
	"github.com/xtls/xray-core/app/proxyman"
	"github.com/xtls/xray-core/app/stats"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/inbound"
	"github.com/xtls/xray-core/features/outbound"
	"github.com/xtls/xray-core/features/routing"
	coreConf "github.com/xtls/xray-core/infra/conf"
	io "io/ioutil"
	"log"
	"sync"
)

// Core Structure
type Core struct {
	access     sync.Mutex
	Server     *core.Instance
	ihm        inbound.Manager
	ohm        outbound.Manager
	dispatcher *dispatcher.DefaultDispatcher
}

func New(c *conf.Conf) *Core {
	return &Core{Server: getCore(c)}
}

func parseConnectionConfig(c *conf.ConnetionConfig) (policy *coreConf.Policy) {
	policy = &coreConf.Policy{
		StatsUserUplink:   true,
		StatsUserDownlink: true,
		Handshake:         &c.Handshake,
		ConnectionIdle:    &c.ConnIdle,
		UplinkOnly:        &c.UplinkOnly,
		DownlinkOnly:      &c.DownlinkOnly,
		BufferSize:        &c.BufferSize,
	}
	return
}

func getCore(v2bXConfig *conf.Conf) *core.Instance {
	// Log Config
	coreLogConfig := &coreConf.LogConfig{}
	coreLogConfig.LogLevel = v2bXConfig.LogConfig.Level
	coreLogConfig.AccessLog = v2bXConfig.LogConfig.AccessPath
	coreLogConfig.ErrorLog = v2bXConfig.LogConfig.ErrorPath
	// DNS config
	coreDnsConfig := &coreConf.DNSConfig{}
	if v2bXConfig.DnsConfigPath != "" {
		if data, err := io.ReadFile(v2bXConfig.DnsConfigPath); err != nil {
			log.Panicf("Failed to read DNS config file at: %s", v2bXConfig.DnsConfigPath)
		} else {
			if err = json.Unmarshal(data, coreDnsConfig); err != nil {
				log.Panicf("Failed to unmarshal DNS config: %s", v2bXConfig.DnsConfigPath)
			}
		}
	}
	dnsConfig, err := coreDnsConfig.Build()
	if err != nil {
		log.Panicf("Failed to understand DNS config, Please check: https://xtls.github.io/config/dns.html for help: %s", err)
	}
	// Routing config
	coreRouterConfig := &coreConf.RouterConfig{}
	if v2bXConfig.RouteConfigPath != "" {
		if data, err := io.ReadFile(v2bXConfig.RouteConfigPath); err != nil {
			log.Panicf("Failed to read Routing config file at: %s", v2bXConfig.RouteConfigPath)
		} else {
			if err = json.Unmarshal(data, coreRouterConfig); err != nil {
				log.Panicf("Failed to unmarshal Routing config: %s", v2bXConfig.RouteConfigPath)
			}
		}
	}
	routeConfig, err := coreRouterConfig.Build()
	if err != nil {
		log.Panicf("Failed to understand Routing config  Please check: https://xtls.github.io/config/routing.html for help: %s", err)
	}
	// Custom Inbound config
	var coreCustomInboundConfig []coreConf.InboundDetourConfig
	if v2bXConfig.InboundConfigPath != "" {
		if data, err := io.ReadFile(v2bXConfig.InboundConfigPath); err != nil {
			log.Panicf("Failed to read Custom Inbound config file at: %s", v2bXConfig.OutboundConfigPath)
		} else {
			if err = json.Unmarshal(data, &coreCustomInboundConfig); err != nil {
				log.Panicf("Failed to unmarshal Custom Inbound config: %s", v2bXConfig.OutboundConfigPath)
			}
		}
	}
	var inBoundConfig []*core.InboundHandlerConfig
	for _, config := range coreCustomInboundConfig {
		oc, err := config.Build()
		if err != nil {
			log.Panicf("Failed to understand Inbound config, Please check: https://xtls.github.io/config/inbound.html for help: %s", err)
		}
		inBoundConfig = append(inBoundConfig, oc)
	}
	// Custom Outbound config
	var coreCustomOutboundConfig []coreConf.OutboundDetourConfig
	if v2bXConfig.OutboundConfigPath != "" {
		if data, err := io.ReadFile(v2bXConfig.OutboundConfigPath); err != nil {
			log.Panicf("Failed to read Custom Outbound config file at: %s", v2bXConfig.OutboundConfigPath)
		} else {
			if err = json.Unmarshal(data, &coreCustomOutboundConfig); err != nil {
				log.Panicf("Failed to unmarshal Custom Outbound config: %s", v2bXConfig.OutboundConfigPath)
			}
		}
	}
	var outBoundConfig []*core.OutboundHandlerConfig
	for _, config := range coreCustomOutboundConfig {
		oc, err := config.Build()
		if err != nil {
			log.Panicf("Failed to understand Outbound config, Please check: https://xtls.github.io/config/outbound.html for help: %s", err)
		}
		outBoundConfig = append(outBoundConfig, oc)
	}
	// Policy config
	levelPolicyConfig := parseConnectionConfig(v2bXConfig.ConnectionConfig)
	corePolicyConfig := &coreConf.PolicyConfig{}
	corePolicyConfig.Levels = map[uint32]*coreConf.Policy{0: levelPolicyConfig}
	policyConfig, _ := corePolicyConfig.Build()
	// Build Core conf
	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(coreLogConfig.Build()),
			serial.ToTypedMessage(&dispatcher.Config{}),
			serial.ToTypedMessage(&stats.Config{}),
			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
			serial.ToTypedMessage(policyConfig),
			serial.ToTypedMessage(dnsConfig),
			serial.ToTypedMessage(routeConfig),
		},
		Inbound:  inBoundConfig,
		Outbound: outBoundConfig,
	}
	server, err := core.New(config)
	if err != nil {
		log.Panicf("failed to create instance: %s", err)
	}
	log.Printf("Core Version: %s", core.Version())

	return server
}

// Start the Core
func (p *Core) Start() {
	p.access.Lock()
	defer p.access.Unlock()
	log.Print("Start the panel..")
	if err := p.Server.Start(); err != nil {
		log.Panicf("Failed to start instance: %s", err)
	}
	p.ihm = p.Server.GetFeature(inbound.ManagerType()).(inbound.Manager)
	p.ohm = p.Server.GetFeature(outbound.ManagerType()).(outbound.Manager)
	p.dispatcher = p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	return
}

// Close  the core
func (p *Core) Close() {
	p.access.Lock()
	defer p.access.Unlock()
	p.Server.Close()
	return
}
