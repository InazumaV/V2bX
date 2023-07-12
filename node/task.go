package node

import (
	"github.com/Yuzuki616/V2bX/common/task"
	vCore "github.com/Yuzuki616/V2bX/core"
	"github.com/Yuzuki616/V2bX/limiter"
	log "github.com/sirupsen/logrus"
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
	log.WithField("tag", c.Tag).Info("Start monitor node status")
	// delay to start nodeInfoMonitor
	_ = c.nodeInfoMonitorPeriodic.Start(false)
	log.WithField("tag", c.Tag).Info("Start report node status")
	_ = c.userReportPeriodic.Start(false)
	if c.nodeInfo.Tls {
		switch c.CertConfig.CertMode {
		case "reality", "none", "":
		default:
			c.renewCertPeriodic = &task.Task{
				Interval: time.Hour * 24,
				Execute:  c.reportUserTrafficTask,
			}
			log.WithField("tag", c.Tag).Info("Start renew cert")
			// delay to start renewCert
			_ = c.renewCertPeriodic.Start(true)
		}
	}
}

func (c *Controller) nodeInfoMonitor() (err error) {
	// get node info
	newNodeInfo, err := c.apiClient.GetNodeInfo()
	if err != nil {
		log.WithFields(log.Fields{
			"tag": c.Tag,
			"err": err,
		}).Error("Get node info failed")
		return nil
	}
	// get user info
	newUserInfo, err := c.apiClient.GetUserList()
	if err != nil {
		log.WithFields(log.Fields{
			"tag": c.Tag,
			"err": err,
		}).Error("Get user list failed")
		return nil
	}
	if newNodeInfo != nil {
		// nodeInfo changed
		// Remove old tag
		log.WithField("tag", c.Tag).Info("Node changed, reload")
		err = c.server.DelNode(c.Tag)
		if err != nil {
			log.WithFields(log.Fields{
				"tag": c.Tag,
				"err": err,
			}).Error("Delete node failed")
			return nil
		}
		// Remove Old limiter
		limiter.DeleteLimiter(c.Tag)
		// Add new Limiter
		c.Tag = c.buildNodeTag()
		l := limiter.AddLimiter(c.Tag, &c.LimitConfig, newUserInfo)
		// check cert
		if newNodeInfo.Tls || newNodeInfo.Type == "hysteria" {
			err = c.requestCert()
			if err != nil {
				log.WithFields(log.Fields{
					"tag": c.Tag,
					"err": err,
				}).Error("Request cert failed")
				return nil
			}
		}
		// add new node
		err = c.server.AddNode(c.Tag, newNodeInfo, c.ControllerConfig)
		if err != nil {
			log.WithFields(log.Fields{
				"tag": c.Tag,
				"err": err,
			}).Error("Add node failed")
			return nil
		}
		_, err = c.server.AddUsers(&vCore.AddUsersParams{
			Tag:      c.Tag,
			Config:   c.ControllerConfig,
			UserInfo: newUserInfo,
			NodeInfo: newNodeInfo,
		})
		if err != nil {
			log.WithFields(log.Fields{
				"tag": c.Tag,
				"err": err,
			}).Error("Add users failed")
			return nil
		}
		err = l.UpdateRule(newNodeInfo.Rules)
		if err != nil {
			log.WithFields(log.Fields{
				"tag": c.Tag,
				"err": err,
			}).Error("Update Rule failed")
			return nil
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
		c.nodeInfo = newNodeInfo
		c.userList = newUserInfo
		log.WithField("tag", c.Tag).Infof("Added %d new users", len(newUserInfo))
		// exit
		return nil
	}

	// node no changed, check users
	deleted, added := compareUserList(c.userList, newUserInfo)
	if len(deleted) > 0 {
		// have deleted users
		err = c.server.DelUsers(deleted, c.Tag)
		if err != nil {
			log.WithFields(log.Fields{
				"tag": c.Tag,
				"err": err,
			}).Error("Delete users failed")
			return nil
		}
	}
	if len(added) > 0 {
		// have added users
		_, err = c.server.AddUsers(&vCore.AddUsersParams{
			Tag:      c.Tag,
			Config:   c.ControllerConfig,
			UserInfo: added,
			NodeInfo: c.nodeInfo,
		})
		if err != nil {
			log.WithFields(log.Fields{
				"tag": c.Tag,
				"err": err,
			}).Error("Add users failed")
			return nil
		}
	}
	if len(added) > 0 || len(deleted) > 0 {
		// update Limiter
		err = limiter.UpdateLimiter(c.Tag, added, deleted)
		if err != nil {
			log.WithFields(log.Fields{
				"tag": c.Tag,
				"err": err,
			}).Error("limiter users failed")
			return nil
		}
	}
	c.userList = newUserInfo
	if len(added)+len(deleted) != 0 {
		log.WithField("tag", c.Tag).
			Infof("%d user deleted, %d user added", len(deleted), len(added))
	}
	return nil
}
