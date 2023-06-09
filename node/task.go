package node

import (
	"fmt"
	"log"
	"runtime"
	"time"

	vCore "github.com/Yuzuki616/V2bX/core"

	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/limiter"
	"github.com/Yuzuki616/V2bX/node/lego"
	"github.com/xtls/xray-core/common/task"
)

func (c *Controller) initTask() {
	// fetch node info task
	c.nodeInfoMonitorPeriodic = &task.Periodic{
		Interval: c.nodeInfo.PullInterval,
		Execute:  c.nodeInfoMonitor,
	}
	// fetch user list task
	c.userReportPeriodic = &task.Periodic{
		Interval: c.nodeInfo.PushInterval,
		Execute:  c.reportUserTraffic,
	}
	log.Printf("[%s: %d] Start monitor node status", c.nodeInfo.Type, c.nodeInfo.Id)
	// delay to start nodeInfoMonitor
	go func() {
		time.Sleep(c.nodeInfo.PullInterval)
		_ = c.nodeInfoMonitorPeriodic.Start()
	}()
	log.Printf("[%s: %d] Start report node status", c.nodeInfo.Type, c.nodeInfo.Id)
	// delay to start userReport
	go func() {
		time.Sleep(c.nodeInfo.PullInterval)
		_ = c.userReportPeriodic.Start()
	}()
	if c.nodeInfo.Tls && c.CertConfig.CertMode != "none" &&
		(c.CertConfig.CertMode == "dns" || c.CertConfig.CertMode == "http") {
		c.renewCertPeriodic = &task.Periodic{
			Interval: time.Hour * 24,
			Execute:  c.reportUserTraffic,
		}
		log.Printf("[%s: %d] Start renew cert", c.nodeInfo.Type, c.nodeInfo.Id)
		// delay to start renewCert
		go func() {
			_ = c.renewCertPeriodic.Start()
		}()
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
		if newNodeInfo.Tls {
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
			_ = c.nodeInfoMonitorPeriodic.Close()
			go func() {
				time.Sleep(c.nodeInfoMonitorPeriodic.Interval)
				_ = c.nodeInfoMonitorPeriodic.Start()
			}()
		}
		if c.userReportPeriodic.Interval != newNodeInfo.PushInterval &&
			newNodeInfo.PushInterval != 0 {
			c.userReportPeriodic.Interval = newNodeInfo.PullInterval
			_ = c.userReportPeriodic.Close()
			go func() {
				time.Sleep(c.userReportPeriodic.Interval)
				_ = c.userReportPeriodic.Start()
			}()
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

func (c *Controller) reportUserTraffic() (err error) {
	// Get User traffic
	userTraffic := make([]panel.UserTraffic, 0)
	for i := range c.userList {
		up, down := c.server.GetUserTraffic(c.Tag, c.userList[i].Uuid, true)
		if up > 0 || down > 0 {
			if c.LimitConfig.EnableDynamicSpeedLimit {
				c.userList[i].Traffic += up + down
			}
			userTraffic = append(userTraffic, panel.UserTraffic{
				UID:      (c.userList)[i].Id,
				Upload:   up,
				Download: down})
		}
	}
	if len(userTraffic) > 0 && !c.DisableUploadTraffic {
		err = c.apiClient.ReportUserTraffic(userTraffic)
		if err != nil {
			log.Printf("Report user traffic faild: %s", err)
		} else {
			log.Printf("[%s: %d] Report %d online users", c.nodeInfo.Type, c.nodeInfo.Id, len(userTraffic))
		}
	}
	userTraffic = nil
	runtime.GC()
	return nil
}

func (c *Controller) RenewCert() {
	l, err := lego.New(c.CertConfig)
	if err != nil {
		log.Print("new lego error: ", err)
		return
	}
	err = l.RenewCert()
	if err != nil {
		log.Print("renew cert error: ", err)
		return
	}
}
