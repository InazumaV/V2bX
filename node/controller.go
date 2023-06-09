package node

import (
	"errors"
	"fmt"
	"log"

	"github.com/Yuzuki616/V2bX/api/iprecoder"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	vCore "github.com/Yuzuki616/V2bX/core"
	"github.com/Yuzuki616/V2bX/limiter"
	"github.com/xtls/xray-core/common/task"
)

type Controller struct {
	server                    vCore.Core
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
func NewController(server vCore.Core, api *panel.Client, config *conf.ControllerConfig) *Controller {
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
		return fmt.Errorf("get node info error: %s", err)
	}
	// Update user
	c.userList, err = c.apiClient.GetUserList()
	if err != nil {
		return fmt.Errorf("get user list error: %s", err)
	}
	if len(c.userList) == 0 {
		return errors.New("add users error: not have any user")
	}
	c.Tag = c.buildNodeTag()

	// add limiter
	l := limiter.AddLimiter(c.Tag, &c.LimitConfig, c.userList)
	// add rule limiter
	if !c.DisableGetRule {
		if err = l.UpdateRule(c.nodeInfo.Rules); err != nil {
			return fmt.Errorf("update rule error: %s", err)
		}
	}
	if c.nodeInfo.Tls {
		err = c.requestCert()
		if err != nil {
			return fmt.Errorf("request cert error: %s", err)
		}
	}
	// Add new tag
	err = c.server.AddNode(c.Tag, c.nodeInfo, c.ControllerConfig)
	if err != nil {
		return fmt.Errorf("add new node error: %s", err)
	}
	added, err := c.server.AddUsers(&vCore.AddUsersParams{
		Tag:      c.Tag,
		Config:   c.ControllerConfig,
		UserInfo: c.userList,
		NodeInfo: c.nodeInfo,
	})
	if err != nil {
		return fmt.Errorf("add users error: %s", err)
	}
	log.Printf("[%s: %d] Added %d new users", c.nodeInfo.Type, c.nodeInfo.Id, added)
	c.initTask()
	return nil
}

// Close implement the Close() function of the service interface
func (c *Controller) Close() error {
	limiter.DeleteLimiter(c.Tag)
	if c.nodeInfoMonitorPeriodic != nil {
		err := c.nodeInfoMonitorPeriodic.Close()
		if err != nil {
			return fmt.Errorf("node info periodic close error: %s", err)
		}
	}
	if c.nodeInfoMonitorPeriodic != nil {
		err := c.userReportPeriodic.Close()
		if err != nil {
			return fmt.Errorf("user report periodic close error: %s", err)
		}
	}
	if c.renewCertPeriodic != nil {
		err := c.renewCertPeriodic.Close()
		if err != nil {
			return fmt.Errorf("renew cert periodic close error: %s", err)
		}
	}
	if c.dynamicSpeedLimitPeriodic != nil {
		err := c.dynamicSpeedLimitPeriodic.Close()
		if err != nil {
			return fmt.Errorf("dynamic speed limit periodic close error: %s", err)
		}
	}
	if c.onlineIpReportPeriodic != nil {
		err := c.onlineIpReportPeriodic.Close()
		if err != nil {
			return fmt.Errorf("online ip report periodic close error: %s", err)
		}
	}
	return nil
}

func (c *Controller) buildNodeTag() string {
	return fmt.Sprintf("%s_%s_%d", c.nodeInfo.Type, c.ListenIP, c.nodeInfo.Id)
}
