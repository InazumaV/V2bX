package controller

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/node/controller/legoCmd"
	"github.com/xtls/xray-core/common/protocol"
	"log"
	"reflect"
	"runtime"
	"time"
)

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
		if c.nodeInfo.SS == nil || !reflect.DeepEqual(c.nodeInfo.SS, newNodeInfo.SS) {
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
		}
	}

	// Check Rule
	if !c.config.DisableGetRule {
		if ruleList, err := c.apiClient.GetNodeRule(); err != nil {
			log.Printf("Get rule list filed: %s", err)
		} else if ruleList != nil {
			if err := c.server.UpdateRule(c.Tag, ruleList); err != nil {
				log.Print(err)
			}

		}
	}

	// Check Cert
	if c.nodeInfo.EnableTls && c.config.CertConfig.CertMode != "none" &&
		(c.config.CertConfig.CertMode == "dns" || c.config.CertConfig.CertMode == "http") {
		lego, err := legoCmd.New()
		if err != nil {
			log.Print(err)
		}
		// Core-core supports the OcspStapling certification hot renew
		_, _, err = lego.RenewCert(c.config.CertConfig.CertDomain, c.config.CertConfig.Email,
			c.config.CertConfig.CertMode, c.config.CertConfig.Provider, c.config.CertConfig.DNSEnv)
		if err != nil {
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
		newUserInfo = nil
		err = c.addNewUser(c.userList, newNodeInfo)
		if err != nil {
			log.Print(err)
			return nil
		}
		newNodeInfo = nil
		// Add Limiter
		if err := c.server.AddInboundLimiter(c.Tag, c.nodeInfo); err != nil {
			log.Print(err)
			return nil
		}
		runtime.GC()
	} else {
		deleted, added := compareUserList(c.userList, newUserInfo)
		if len(deleted) > 0 {
			deletedEmail := make([]string, len(deleted))
			for i := range deleted {
				deletedEmail[i] = fmt.Sprintf("%s|%s|%d", c.Tag,
					(deleted)[i].GetUserEmail(),
					(deleted)[i].UID)
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
			defer runtime.GC()
			// Update Limiter
			if err := c.server.UpdateInboundLimiter(c.Tag, deleted); err != nil {
				log.Print(err)
			}
		}
		log.Printf("[%s: %d] %d user deleted, %d user added", c.nodeInfo.NodeType, c.nodeInfo.NodeId,
			len(deleted), len(added))
		c.userList = newUserInfo
		newUserInfo = nil
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
	inboundConfig, err := buildInbound(c.config, newNodeInfo, c.Tag)
	if err != nil {
		return err
	}
	err = c.server.AddInbound(inboundConfig)
	if err != nil {

		return err
	}
	outBoundConfig, err := buildOutbound(c.config, newNodeInfo, c.Tag)
	if err != nil {

		return err
	}
	err = c.server.AddOutbound(outBoundConfig)
	if err != nil {

		return err
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
		users = c.buildSSUsers(userInfo, getCipherFromString(nodeInfo.SS.CypherMethod))
	} else {
		return fmt.Errorf("unsupported node type: %s", nodeInfo.NodeType)
	}
	err = c.server.AddUsers(users, c.Tag)
	if err != nil {
		return err
	}
	log.Printf("[%s: %d] Added %d new users", c.nodeInfo.NodeType, c.nodeInfo.NodeId, len(userInfo))
	return nil
}

func compareUserList(old, new []panel.UserInfo) (deleted, added []panel.UserInfo) {
	tmp := map[string]struct{}{}
	tmp2 := map[string]struct{}{}
	for i := range old {
		tmp[(old)[i].GetUserEmail()] = struct{}{}
	}
	l := len(tmp)
	for i := range new {
		e := (new)[i].GetUserEmail()
		tmp[e] = struct{}{}
		tmp2[e] = struct{}{}
		if l != len(tmp) {
			added = append(added, (new)[i])
			l++
		}
	}
	tmp = nil
	l = len(tmp2)
	for i := range old {
		tmp2[(old)[i].GetUserEmail()] = struct{}{}
		if l != len(tmp2) {
			deleted = append(deleted, (old)[i])
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
			if c.config.EnableDynamicSpeedLimit {
				c.userList[i].Traffic += up + down
			}
			userTraffic = append(userTraffic, panel.UserTraffic{
				UID:      (c.userList)[i].UID,
				Upload:   up,
				Download: down})
		}
	}
	if len(userTraffic) > 0 && !c.config.DisableUploadTraffic {
		err = c.apiClient.ReportUserTraffic(userTraffic)
		if err != nil {
			log.Printf("Report user traffic faild: %s", err)
		} else {
			log.Printf("[%s: %d] Report %d online users", c.nodeInfo.NodeType, c.nodeInfo.NodeId, len(userTraffic))
		}
	}
	userTraffic = nil
	if !c.config.EnableIpRecorder {
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
	if c.config.IpRecorderConfig.EnableIpSync {
		c.server.UpdateOnlineIp(c.Tag, onlineIp)
		log.Printf("[Node: %d] Updated %d online ip", c.nodeInfo.NodeId, len(onlineIp))
	}
	log.Printf("[Node: %d] Report %d online ip", c.nodeInfo.NodeId, len(onlineIp))
	return nil
}

func (c *Node) dynamicSpeedLimit() error {
	if c.config.EnableDynamicSpeedLimit {
		for i := range c.userList {
			up, down := c.server.GetUserTraffic(c.buildUserTag(&(c.userList)[i]), false)
			if c.userList[i].Traffic+down+up/1024/1024 > c.config.DynamicSpeedLimitConfig.Traffic {
				err := c.server.AddUserSpeedLimit(c.Tag,
					&c.userList[i],
					c.config.DynamicSpeedLimitConfig.SpeedLimit,
					time.Now().Add(time.Second*time.Duration(c.config.DynamicSpeedLimitConfig.ExpireTime)).Unix())
				if err != nil {
					log.Print(err)
				}
			}
			c.userList[i].Traffic = 0
		}
	}
	return nil
}
