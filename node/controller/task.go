package controller

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/api/iprecoder"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/node/controller/lego"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/task"
	"log"
	"runtime"
	"strconv"
	"time"
)

func (c *Node) initTask() {
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
	if c.nodeInfo.EnableTls && c.CertConfig.CertMode != "none" &&
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
	if c.EnableDynamicSpeedLimit {
		// Check dynamic speed limit task
		c.dynamicSpeedLimitPeriodic = &task.Periodic{
			Interval: time.Duration(c.DynamicSpeedLimitConfig.Periodic) * time.Second,
			Execute:  c.dynamicSpeedLimit,
		}
		go func() {
			time.Sleep(time.Duration(c.DynamicSpeedLimitConfig.Periodic) * time.Second)
			_ = c.dynamicSpeedLimitPeriodic.Start()
		}()
		log.Printf("[%s: %d] Start dynamic speed limit", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
	}
	if c.EnableIpRecorder {
		switch c.IpRecorderConfig.Type {
		case "Recorder":
			c.ipRecorder = iprecoder.NewRecorder(c.IpRecorderConfig.RecorderConfig)
		case "Redis":
			c.ipRecorder = iprecoder.NewRedis(c.IpRecorderConfig.RedisConfig)
		default:
			log.Printf("recorder type: %s is not vail, disable recorder", c.IpRecorderConfig.Type)
			return
		}
		// report and fetch online ip list task
		c.onlineIpReportPeriodic = &task.Periodic{
			Interval: time.Duration(c.IpRecorderConfig.Periodic) * time.Second,
			Execute:  c.reportOnlineIp,
		}
		go func() {
			time.Sleep(time.Duration(c.IpRecorderConfig.Periodic) * time.Second)
			_ = c.onlineIpReportPeriodic.Start()
		}()
		log.Printf("[%s: %d] Start report online ip", c.nodeInfo.NodeType, c.nodeInfo.NodeId)
	}
}

func (c *Node) nodeInfoMonitor() (err error) {
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
		err := c.removeOldTag(oldTag)
		if err != nil {
			log.Print(err)
			return nil
		}
		// Add new tag
		c.nodeInfo = newNodeInfo
		c.Tag = c.buildNodeTag()
		err = c.addNewTag(newNodeInfo)
		if err != nil {
			log.Print(err)
			return nil
		}
		nodeInfoChanged = true
		// Remove Old limiter
		if err = c.server.DeleteInboundLimiter(oldTag); err != nil {
			log.Print(err)
			return nil
		}
		if err := c.server.UpdateRule(c.Tag, newNodeInfo.Rules); err != nil {
			log.Print(err)
		}
	}
	// Update User
	newUserInfo, err := c.apiClient.GetUserList()
	if err != nil {
		log.Print(err)
		return nil
	}
	if nodeInfoChanged {
		c.userList = newUserInfo
		err = c.addNewUser(c.userList, newNodeInfo)
		if err != nil {
			log.Print(err)
			return nil
		}
		// Add Limiter
		if err := c.server.AddInboundLimiter(c.Tag, newNodeInfo, newUserInfo); err != nil {
			log.Print(err)
			return nil
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
			if err := c.server.UpdateInboundLimiter(c.Tag, added, deleted); err != nil {
				log.Print(err)
			}
		}
		log.Printf("[%s: %d] %d user deleted, %d user added", c.nodeInfo.NodeType, c.nodeInfo.NodeId,
			len(deleted), len(added))
		c.userList = newUserInfo
	}
	return nil
}

func (c *Node) removeOldTag(oldTag string) (err error) {
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

func (c *Node) addNewTag(newNodeInfo *panel.NodeInfo) (err error) {
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

func (c *Node) addNewUser(userInfo []panel.UserInfo, nodeInfo *panel.NodeInfo) (err error) {
	users := make([]*protocol.User, 0)
	if nodeInfo.NodeType == "V2ray" {
		if nodeInfo.EnableVless {
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

func (c *Node) reportUserTraffic() (err error) {
	// Get User traffic
	userTraffic := make([]panel.UserTraffic, 0)
	for i := range c.userList {
		up, down := c.server.GetUserTraffic(c.buildUserTag(&(c.userList)[i]), true)
		if up > 0 || down > 0 {
			if c.EnableDynamicSpeedLimit {
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
	if !c.EnableIpRecorder {
		c.server.ClearOnlineIp(c.Tag)
	}
	runtime.GC()
	return nil
}

func (c *Node) reportOnlineIp() (err error) {
	onlineIp, err := c.server.ListOnlineIp(c.Tag)
	if err != nil {
		log.Print(err)
		return nil
	}
	onlineIp, err = c.ipRecorder.SyncOnlineIp(onlineIp)
	if err != nil {
		log.Print("Report online ip error: ", err)
		c.server.ClearOnlineIp(c.Tag)
	}
	if c.IpRecorderConfig.EnableIpSync {
		c.server.UpdateOnlineIp(c.Tag, onlineIp)
		log.Printf("[Node: %d] Updated %d online ip", c.nodeInfo.NodeId, len(onlineIp))
	}
	log.Printf("[Node: %d] Report %d online ip", c.nodeInfo.NodeId, len(onlineIp))
	return nil
}

func (c *Node) dynamicSpeedLimit() error {
	if c.EnableDynamicSpeedLimit {
		for i := range c.userList {
			up, down := c.server.GetUserTraffic(c.buildUserTag(&(c.userList)[i]), false)
			if c.userList[i].Traffic+down+up/1024/1024 > c.DynamicSpeedLimitConfig.Traffic {
				err := c.server.AddUserSpeedLimit(c.Tag,
					&c.userList[i],
					c.DynamicSpeedLimitConfig.SpeedLimit,
					time.Now().Add(time.Second*time.Duration(c.DynamicSpeedLimitConfig.ExpireTime)).Unix())
				if err != nil {
					log.Print(err)
				}
			}
			c.userList[i].Traffic = 0
		}
	}
	return nil
}

func (c *Node) RenewCert() {
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
