package node

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/common/task"
	vCore "github.com/Yuzuki616/V2bX/core"
	"github.com/Yuzuki616/V2bX/limiter"
	"log"
	"time"
)

func (c *Controller) initTask() {
	// fetch node info task
	c.nodeInfoMonitorPeriodic = &task.Task{
		Interval: c.nodeInfo.PullInterval,
		Execute:  c.nodeInfoMonitor,
	}
	// fetch user list task
	c.userReportPeriodic = &task.Task{
		Interval: c.nodeInfo.PushInterval,
		Execute:  c.reportUserTrafficTask,
	}
	log.Printf("[%s] Start monitor node status", c.Tag)
	// delay to start nodeInfoMonitor
	_ = c.nodeInfoMonitorPeriodic.Start(false)
	log.Printf("[%s] Start report node status", c.Tag)
	_ = c.userReportPeriodic.Start(false)
	if c.nodeInfo.Tls {
		switch c.CertConfig.CertMode {
		case "reality", "none", "":
		default:
			c.renewCertPeriodic = &task.Task{
				Interval: time.Hour * 24,
				Execute:  c.reportUserTrafficTask,
			}
			log.Printf("[%s] Start renew cert", c.Tag)
			// delay to start renewCert
			_ = c.renewCertPeriodic.Start(true)
		}
	}
}

func (c *Controller) nodeInfoMonitor() (err error) {
	// First fetch Node Info
	newNodeInfo, err := c.apiClient.GetNodeInfo()
	if err != nil {
		log.Print(err)
		return nil
	}
	var nodeInfoChanged = false
	// If nodeInfo changed
	if newNodeInfo != nil {
		// Remove old tag
		oldTag := c.Tag
		err := c.server.DelNode(oldTag)
		if err != nil {
			log.Print(err)
			return nil
		}
		// Remove Old limiter
		limiter.DeleteLimiter(oldTag)
		// Add new tag
		c.nodeInfo = newNodeInfo
		c.Tag = c.buildNodeTag()
		err = c.server.AddNode(c.Tag, newNodeInfo, c.ControllerConfig)
		if err != nil {
			log.Print(err)
			return nil
		}
		if c.nodeInfo.Tls || c.nodeInfo.Type == "hysteria" {
			err = c.requestCert()
			if err != nil {
				return fmt.Errorf("request cert error: %s", err)
			}
		}
		nodeInfoChanged = true
	}
	// Update User
	newUserInfo, err := c.apiClient.GetUserList()
	if err != nil {
		log.Print(err)
		return nil
	}
	if nodeInfoChanged {
		c.userList = newUserInfo
		// Add new Limiter
		l := limiter.AddLimiter(c.Tag, &c.LimitConfig, newUserInfo)
		_, err = c.server.AddUsers(&vCore.AddUsersParams{
			Tag:      c.Tag,
			Config:   c.ControllerConfig,
			UserInfo: newUserInfo,
			NodeInfo: newNodeInfo,
		})
		if err != nil {
			log.Print(err)
			return nil
		}
		err = l.UpdateRule(newNodeInfo.Rules)
		if err != nil {
			log.Printf("Update Rule error: %s", err)
		}
		// Check interval
		if c.nodeInfoMonitorPeriodic.Interval != newNodeInfo.PullInterval &&
			newNodeInfo.PullInterval != 0 {
			c.nodeInfoMonitorPeriodic.Interval = newNodeInfo.PullInterval
			c.nodeInfoMonitorPeriodic.Close()
			_ = c.nodeInfoMonitorPeriodic.Start(false)
		}
		if c.userReportPeriodic.Interval != newNodeInfo.PushInterval &&
			newNodeInfo.PushInterval != 0 {
			c.userReportPeriodic.Interval = newNodeInfo.PullInterval
			c.userReportPeriodic.Close()
			_ = c.userReportPeriodic.Start(false)
		}
	} else {
		deleted, added := compareUserList(c.userList, newUserInfo)
		if len(deleted) > 0 {
			deletedEmail := make([]string, len(deleted))
			for i := range deleted {
				deletedEmail[i] = fmt.Sprintf("%s|%s|%d",
					c.Tag,
					(deleted)[i].Uuid,
					(deleted)[i].Id)
			}
			err := c.server.DelUsers(deletedEmail, c.Tag)
			if err != nil {
				log.Print(err)
			}
		}
		if len(added) > 0 {
			_, err := c.server.AddUsers(&vCore.AddUsersParams{
				Tag:      c.Tag,
				Config:   c.ControllerConfig,
				UserInfo: added,
				NodeInfo: c.nodeInfo,
			})
			if err != nil {
				log.Print(err)
			}
		}
		if len(added) > 0 || len(deleted) > 0 {
			// Update Limiter
			err = limiter.UpdateLimiter(c.Tag, added, deleted)
			if err != nil {
				log.Print("update limiter:", err)
			}
		}
		log.Printf("[%s: %d] %d user deleted, %d user added", c.nodeInfo.Type, c.nodeInfo.Id,
			len(deleted), len(added))
		c.userList = newUserInfo
	}
	return nil
}
