package node

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/limiter"
	"github.com/Yuzuki616/V2bX/node/lego"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/task"
	"log"
	"runtime"
	"strconv"
	"time"
)

func (c *Controller) initTask() {
	// fetch node info task
	c.nodeInfoMonitorPeriodic = &task.Periodic{
		Interval: time.Duration(c.nodeInfo.BaseConfig.PullInterval.(int)) * time.Second,
		Execute:  c.nodeInfoMonitor,
	}
	// fetch user list task
	c.userReportPeriodic = &task.Periodic{
		Interval: time.Duration(c.nodeInfo.BaseConfig.PushInterval.(int)) * time.Second,
		Execute:  c.reportUserTraffic,
	}
	log.Printf("[%s: %d] Start monitor node status", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
	// delay to start nodeInfoMonitor
	go func() {
		time.Sleep(time.Duration(c.nodeInfo.BaseConfig.PullInterval.(int)) * time.Second)
		_ = c.nodeInfoMonitorPeriodic.Start()
	}()
	log.Printf("[%s: %d] Start report node status", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
	// delay to start userReport
	go func() {
		time.Sleep(time.Duration(c.nodeInfo.BaseConfig.PushInterval.(int)) * time.Second)
		_ = c.userReportPeriodic.Start()
	}()
	if c.EnableTls && c.CertConfig.CertMode != "none" &&
		(c.CertConfig.CertMode == "dns" || c.CertConfig.CertMode == "http") {
		c.renewCertPeriodic = &task.Periodic{
			Interval: time.Hour * 24,
			Execute:  c.reportUserTraffic,
		}
		log.Printf("[%s: %d] Start renew cert", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
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
		err := c.removeOldNode(oldTag)
		if err != nil {
			log.Print(err)
			return nil
		}
		// Remove Old limiter
		limiter.DeleteLimiter(oldTag)
		// Add new tag
		c.nodeInfo = newNodeInfo
		c.Tag = c.buildNodeTag()
		err = c.addNewNode(newNodeInfo)
		if err != nil {
			log.Print(err)
			return nil
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
		err = c.addNewUser(newUserInfo, newNodeInfo)
		if err != nil {
			log.Print(err)
			return nil
		}
		err = l.UpdateRule(newNodeInfo.Rules)
		if err != nil {
			log.Printf("Update Rule error: %s", err)
		}
		// Check interval
		if c.nodeInfoMonitorPeriodic.Interval != time.Duration(newNodeInfo.BaseConfig.PullInterval.(int))*time.Second {
			c.nodeInfoMonitorPeriodic.Interval = time.Duration(newNodeInfo.BaseConfig.PullInterval.(int)) * time.Second
			_ = c.nodeInfoMonitorPeriodic.Close()
			go func() {
				time.Sleep(c.nodeInfoMonitorPeriodic.Interval)
				_ = c.nodeInfoMonitorPeriodic.Start()
			}()
		}
		if c.userReportPeriodic.Interval != time.Duration(newNodeInfo.BaseConfig.PushInterval.(int))*time.Second {
			c.userReportPeriodic.Interval = time.Duration(newNodeInfo.BaseConfig.PushInterval.(int)) * time.Second
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
			err := c.server.RemoveUsers(deletedEmail, c.Tag)
			if err != nil {
				log.Print(err)
			}
		}
		if len(added) > 0 {
			err = c.addNewUser(added, c.nodeInfo)
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
		log.Printf("[%s: %d] %d user deleted, %d user added", c.nodeInfo.NodeType, c.nodeInfo.NodeId,
			len(deleted), len(added))
		c.userList = newUserInfo
	}
	return nil
}

func (c *Controller) removeOldNode(oldTag string) (err error) {
	err = c.server.RemoveInbound(oldTag)
	if err != nil {
		return err
	}
	err = c.server.RemoveOutbound(oldTag)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) addNewNode(newNodeInfo *panel.NodeInfo) (err error) {
	inboundConfig, err := buildInbound(c.ControllerConfig, newNodeInfo, c.Tag)
	if err != nil {
		return fmt.Errorf("build inbound error: %s", err)
	}
	err = c.server.AddInbound(inboundConfig)
	if err != nil {
		return fmt.Errorf("add inbound error: %s", err)
	}
	outBoundConfig, err := buildOutbound(c.ControllerConfig, newNodeInfo, c.Tag)
	if err != nil {
		return fmt.Errorf("build outbound error: %s", err)
	}
	err = c.server.AddOutbound(outBoundConfig)
	if err != nil {
		return fmt.Errorf("add outbound error: %s", err)
	}
	return nil
}

func (c *Controller) addNewUser(userInfo []panel.UserInfo, nodeInfo *panel.NodeInfo) (err error) {
	users := make([]*protocol.User, 0)
	if nodeInfo.NodeType == "V2ray" {
		if c.EnableVless {
			users = c.buildVlessUsers(userInfo)
		} else {
			users = c.buildVmessUsers(userInfo)
		}
	} else if nodeInfo.NodeType == "Trojan" {
		users = c.buildTrojanUsers(userInfo)
	} else if nodeInfo.NodeType == "Shadowsocks" {
		users = c.buildSSUsers(userInfo, getCipherFromString(nodeInfo.Cipher))
	} else {
		return fmt.Errorf("unsupported node type: %s", nodeInfo.NodeType)
	}
	err = c.server.AddUsers(users, c.Tag)
	if err != nil {
		return fmt.Errorf("add users error: %s", err)
	}
	log.Printf("[%s: %d] Added %d new users", c.nodeInfo.NodeType, c.nodeInfo.NodeId, len(userInfo))
	return nil
}

func compareUserList(old, new []panel.UserInfo) (deleted, added []panel.UserInfo) {
	tmp := map[string]struct{}{}
	tmp2 := map[string]struct{}{}
	for i := range old {
		tmp[old[i].Uuid+strconv.Itoa(old[i].SpeedLimit)] = struct{}{}
	}
	l := len(tmp)
	for i := range new {
		e := new[i].Uuid + strconv.Itoa(new[i].SpeedLimit)
		tmp[e] = struct{}{}
		tmp2[e] = struct{}{}
		if l != len(tmp) {
			added = append(added, new[i])
			l++
		}
	}
	tmp = nil
	l = len(tmp2)
	for i := range old {
		tmp2[old[i].Uuid+strconv.Itoa(old[i].SpeedLimit)] = struct{}{}
		if l != len(tmp2) {
			deleted = append(deleted, old[i])
			l++
		}
	}
	return deleted, added
}

func (c *Controller) reportUserTraffic() (err error) {
	// Get User traffic
	userTraffic := make([]panel.UserTraffic, 0)
	for i := range c.userList {
		up, down := c.server.GetUserTraffic(c.buildUserTag(&(c.userList)[i]), true)
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
			log.Printf("[%s: %d] Report %d online users", c.nodeInfo.NodeType, c.nodeInfo.NodeId, len(userTraffic))
		}
	}
	userTraffic = nil
	runtime.GC()
	return nil
}

func (c *Controller) RenewCert() {
	l, err := lego.New(c.CertConfig)
	if err != nil {
		log.Print(err)
		return
	}
	err = l.RenewCert()
	if err != nil {
		log.Print(err)
		return
	}
}
