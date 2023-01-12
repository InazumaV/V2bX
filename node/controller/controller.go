package controller

import (
	"errors"
	"fmt"
	"github.com/Yuzuki616/V2bX/api/iprecoder"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core"
	"github.com/xtls/xray-core/common/task"
	"log"
)

type Node struct {
	server                    *core.Core
	clientInfo                panel.ClientInfo
	apiClient                 panel.Panel
	nodeInfo                  *panel.NodeInfo
	Tag                       string
	userList                  []panel.UserInfo
	ipRecorder                iprecoder.IpRecorder
	nodeInfoMonitorPeriodic   *task.Periodic
	userReportPeriodic        *task.Periodic
	renewCertPeriodic         *task.Periodic
	dynamicSpeedLimitPeriodic *task.Periodic
	onlineIpReportPeriodic    *task.Periodic
	*conf.ControllerConfig
}

// New return a Node service with default parameters.
func New(server *core.Core, api panel.Panel, config *conf.ControllerConfig) *Node {
	controller := &Node{
		server:           server,
		ControllerConfig: config,
		apiClient:        api,
	}
	return controller
}

// Start implement the Start() function of the service interface
func (c *Node) Start() error {
	c.clientInfo = c.apiClient.Describe()
	// First fetch Node Info
	var err error
	c.nodeInfo, err = c.apiClient.GetNodeInfo()
	if err != nil {
		return fmt.Errorf("get node info failed: %s", err)
	}
	c.Tag = c.buildNodeTag()
	// Add new tag
	err = c.addNewTag(c.nodeInfo)
	if err != nil {
		return fmt.Errorf("add new tag failed: %s", err)
	}
	// Update user
	c.userList, err = c.apiClient.GetUserList()
	if err != nil {
		return fmt.Errorf("get user list failed: %s", err)
	}
	if len(c.userList) == 0 {
		return errors.New("add users failed: not have any user")
	}
	err = c.addNewUser(c.userList, c.nodeInfo)
	if err != nil {
		return err
	}
	if err := c.server.AddInboundLimiter(c.Tag, c.nodeInfo, c.userList); err != nil {
		return fmt.Errorf("add inbound limiter failed: %s", err)
	}
	// Add Rule Manager
	if !c.DisableGetRule {
		if err := c.server.UpdateRule(c.Tag, c.nodeInfo.Rules); err != nil {
			log.Printf("Update rule filed: %s", err)
		}
	}
	c.initTask()
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
	if c.renewCertPeriodic != nil {
		err := c.renewCertPeriodic.Close()
		if err != nil {
			log.Panicf("renew cert periodic close failed: %s", err)
		}
	}
	if c.dynamicSpeedLimitPeriodic != nil {
		err := c.dynamicSpeedLimitPeriodic.Close()
		if err != nil {
			log.Panicf("dynamic speed limit periodic close failed: %s", err)
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
	return fmt.Sprintf("%s_%s_%d", c.nodeInfo.NodeType, c.ListenIP, c.nodeInfo.NodeId)
}
