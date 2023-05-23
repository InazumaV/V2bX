package node

import (
	"errors"
	"fmt"
	"github.com/Yuzuki616/V2bX/api/iprecoder"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core"
	"github.com/Yuzuki616/V2bX/limiter"
	"github.com/xtls/xray-core/common/task"
	"log"
)

type Controller struct {
	server                    *core.Core
	apiClient                 *panel.Client
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

// NewController return a Node controller with default parameters.
func NewController(server *core.Core, api *panel.Client, config *conf.ControllerConfig) *Controller {
	controller := &Controller{
		server:           server,
		ControllerConfig: config,
		apiClient:        api,
	}
	return controller
}

// Start implement the Start() function of the service interface
func (c *Controller) Start() error {
	// First fetch Node Info
	var err error
	c.nodeInfo, err = c.apiClient.GetNodeInfo()
	if err != nil {
		return fmt.Errorf("get node info failed: %s", err)
	}
	// Update user
	c.userList, err = c.apiClient.GetUserList()
	if err != nil {
		return fmt.Errorf("get user list failed: %s", err)
	}
	if len(c.userList) == 0 {
		return errors.New("add users failed: not have any user")
	}
	c.Tag = c.buildNodeTag()

	// add limiter
	l := limiter.AddLimiter(c.Tag, &c.LimitConfig, c.userList)
	// add rule limiter
	if !c.DisableGetRule {
		if err = l.UpdateRule(c.nodeInfo.Rules); err != nil {
			log.Printf("Update rule filed: %s", err)
		}
	}
	// Add new tag
	err = c.server.AddNode(c.Tag, c.nodeInfo, c.ControllerConfig)
	if err != nil {
		return fmt.Errorf("add new tag failed: %s", err)
	}
	added, err := c.server.AddUsers(&core.AddUsersParams{
		Tag:      c.Tag,
		Config:   c.ControllerConfig,
		UserInfo: c.userList,
		NodeInfo: c.nodeInfo,
	})
	log.Printf("[%s: %d] Added %d new users", c.nodeInfo.NodeType, c.nodeInfo.NodeId, added)
	if err != nil {
		return err
	}
	c.initTask()
	return nil
}

// Close implement the Close() function of the service interface
func (c *Controller) Close() error {
	limiter.DeleteLimiter(c.Tag)
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

func (c *Controller) buildNodeTag() string {
	return fmt.Sprintf("%s_%s_%d", c.nodeInfo.NodeType, c.ListenIP, c.nodeInfo.NodeId)
}
