package node

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core"
	"github.com/xtls/xray-core/common/task"
	"log"
	"runtime"
	"time"
)

type Node struct {
	server                    *core.Core
	config                    *conf.ControllerConfig
	clientInfo                panel.ClientInfo
	apiClient                 panel.Panel
	nodeInfo                  *panel.NodeInfo
	Tag                       string
	userList                  []panel.UserInfo
	nodeInfoMonitorPeriodic   *task.Periodic
	userReportPeriodic        *task.Periodic
	onlineIpReportPeriodic    *task.Periodic
	DynamicSpeedLimitPeriodic *task.Periodic
}

// New return a Node service with default parameters.
func New(server *core.Core, api panel.Panel, config *conf.ControllerConfig) *Node {
	controller := &Node{
		server:    server,
		config:    config,
		apiClient: api,
	}
	return controller
}

// Start implement the Start() function of the service interface
func (c *Node) Start() error {
	c.clientInfo = c.apiClient.Describe()
	// First fetch Node Info
	newNodeInfo, err := c.apiClient.GetNodeInfo()
	if err != nil {
		return err
	}
	c.nodeInfo = newNodeInfo
	c.Tag = c.buildNodeTag()
	// Add new tag
	err = c.addNewTag(newNodeInfo)
	if err != nil {
		log.Panic(err)
		return err
	}
	// Update user
	c.userList, err = c.apiClient.GetUserList()
	if err != nil {
		return err
	}
	err = c.addNewUser(c.userList, newNodeInfo)
	if err != nil {
		return err
	}
	if err := c.server.AddInboundLimiter(c.Tag, c.nodeInfo); err != nil {
		log.Print(err)
	}
	// Add Rule Manager
	if !c.config.DisableGetRule {
		if ruleList, err := c.apiClient.GetNodeRule(); err != nil {
			log.Printf("Get rule list filed: %s", err)
		} else if ruleList != nil {
			if err := c.server.UpdateRule(c.Tag, ruleList); err != nil {
				log.Print(err)
			}
		}
	}
	c.nodeInfoMonitorPeriodic = &task.Periodic{
		Interval: time.Duration(c.config.UpdatePeriodic) * time.Second,
		Execute:  c.nodeInfoMonitor,
	}
	c.userReportPeriodic = &task.Periodic{
		Interval: time.Duration(c.config.UpdatePeriodic) * time.Second,
		Execute:  c.userInfoMonitor,
	}
	log.Printf("[%s: %d] Start monitor node status", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
	// delay to start nodeInfoMonitor
	go func() {
		time.Sleep(time.Duration(c.config.UpdatePeriodic) * time.Second)
		_ = c.nodeInfoMonitorPeriodic.Start()
	}()

	log.Printf("[%s: %d] Start report node status", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
	// delay to start userReport
	go func() {
		time.Sleep(time.Duration(c.config.UpdatePeriodic) * time.Second)
		_ = c.userReportPeriodic.Start()
	}()
	if c.config.EnableIpRecorder {
		c.onlineIpReportPeriodic = &task.Periodic{
			Interval: time.Duration(c.config.IpRecorderConfig.Periodic) * time.Second,
			Execute:  c.onlineIpReport,
		}
		go func() {
			time.Sleep(time.Duration(c.config.IpRecorderConfig.Periodic) * time.Second)
			_ = c.onlineIpReportPeriodic.Start()
		}()
		log.Printf("[%s: %d] Start report online ip", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
	}
	if c.config.EnableDynamicSpeedLimit {
		c.DynamicSpeedLimitPeriodic = &task.Periodic{
			Interval: time.Duration(c.config.DynamicSpeedLimitConfig.Periodic) * time.Second,
			Execute:  c.DynamicSpeedLimit,
		}
		go func() {
			time.Sleep(time.Duration(c.config.DynamicSpeedLimitConfig.Periodic) * time.Second)
			_ = c.DynamicSpeedLimitPeriodic.Start()
		}()
		log.Printf("[%s: %d] Start dynamic speed limit", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
	}
	runtime.GC()
	return nil
}

// Close implement the Close() function of the service interface
func (c *Node) Close() error {
	if c.nodeInfoMonitorPeriodic != nil {
		err := c.nodeInfoMonitorPeriodic.Close()
		if err != nil {
			log.Panicf("node info periodic close failed: %s", err)
		}
	}

	if c.nodeInfoMonitorPeriodic != nil {
		err := c.userReportPeriodic.Close()
		if err != nil {
			log.Panicf("user report periodic close failed: %s", err)
		}
	}
	if c.onlineIpReportPeriodic != nil {
		err := c.onlineIpReportPeriodic.Close()
		if err != nil {
			log.Panicf("online ip report periodic close failed: %s", err)
		}
	}
	return nil
}

func (c *Node) buildNodeTag() string {
	return fmt.Sprintf("%s_%s_%d", c.nodeInfo.NodeType, c.config.ListenIP, c.nodeInfo.NodeId)
}
