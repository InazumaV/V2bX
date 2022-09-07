package node

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core"
	"github.com/Yuzuki616/V2bX/core/app/dispatcher"
	"github.com/Yuzuki616/V2bX/node/legoCmd"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/task"
	"log"
	"math"
	"reflect"
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
	userInfo, err := c.apiClient.GetUserList()
	if err != nil {
		return err
	}
	err = c.addNewUser(userInfo, newNodeInfo)
	if err != nil {
		return err
	}
	//sync controller userList
	c.userList = userInfo
	if err := c.server.AddInboundLimiter(c.Tag, c.nodeInfo, userInfo); err != nil {
		log.Print(err)
	}
	// Add Rule Manager
	if !c.config.DisableGetRule {
		if ruleList, protocolRule, err := c.apiClient.GetNodeRule(); err != nil {
			log.Printf("Get rule list filed: %s", err)
		} else if len(ruleList) > 0 {
			if err := c.server.UpdateRule(c.Tag, ruleList); err != nil {
				log.Print(err)
			}
			if len(protocolRule) > 0 {
				if err := c.server.UpdateProtocolRule(c.Tag, protocolRule); err != nil {
					log.Print(err)
				}
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
			oldtag := c.Tag
			err := c.removeOldTag(oldtag)
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
			if err = c.server.DeleteInboundLimiter(oldtag); err != nil {
				log.Print(err)
				return nil
			}
		}
	}

	// Check Rule
	if !c.config.DisableGetRule {
		if ruleList, protocolRule, err := c.apiClient.GetNodeRule(); err != nil {
			log.Printf("Get rule list filed: %s", err)
		} else if len(ruleList) > 0 {
			if err := c.server.UpdateRule(c.Tag, ruleList); err != nil {
				log.Print(err)
			}
			if len(protocolRule) > 0 {
				if err := c.server.UpdateProtocolRule(c.Tag, protocolRule); err != nil {
					log.Print(err)
				}
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
		if err := c.server.AddInboundLimiter(c.Tag, c.nodeInfo, c.userList); err != nil {
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
			// Update Limiter
			if err := c.server.UpdateInboundLimiter(c.Tag, c.nodeInfo, added, deleted); err != nil {
				log.Print(err)
			}
		}
		log.Printf("[%s: %d] %d user deleted, %d user added", c.nodeInfo.NodeType, c.nodeInfo.NodeId,
			len(deleted), len(added))
		c.userList = newUserInfo
		newUserInfo = nil
		runtime.GC()
	}
	return nil
}

func (c *Node) removeOldTag(oldtag string) (err error) {
	err = c.server.RemoveInbound(oldtag)
	if err != nil {
		return err
	}
	err = c.server.RemoveOutbound(oldtag)
	if err != nil {
		return err
	}
	return nil
}

func (c *Node) addNewTag(newNodeInfo *panel.NodeInfo) (err error) {
	inboundConfig, err := InboundBuilder(c.config, newNodeInfo, c.Tag)
	if err != nil {
		return err
	}
	err = c.server.AddInbound(inboundConfig)
	if err != nil {

		return err
	}
	outBoundConfig, err := OutboundBuilder(c.config, newNodeInfo, c.Tag)
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
			alterID := 0
			alterID = (userInfo)[0].V2rayUser.AlterId
			if alterID >= 0 && alterID < math.MaxUint16 {
				users = c.buildVmessUsers(userInfo, uint16(alterID))
			} else {
				users = c.buildVmessUsers(userInfo, 0)
				return fmt.Errorf("AlterID should between 0 to 1<<16 - 1, set it to 0 for now")
			}
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

func (c *Node) userInfoMonitor() (err error) {
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
			log.Print(err)
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

func (c *Node) onlineIpReport() (err error) {
	onlineIp, err := c.server.ListOnlineIp(c.Tag)
	if err != nil {
		log.Print(err)
		return nil
	}
	rsp, err := resty.New().SetTimeout(time.Duration(c.config.IpRecorderConfig.Timeout) * time.Second).
		R().
		SetBody(onlineIp).
		Post(c.config.IpRecorderConfig.Url +
			"/api/v1/SyncOnlineIp?token=" +
			c.config.IpRecorderConfig.Token)
	if err != nil {
		log.Print(err)
		c.server.ClearOnlineIp(c.Tag)
		return nil
	}
	log.Printf("[Node: %d] Report %d online ip", c.nodeInfo.NodeId, len(onlineIp))
	if rsp.StatusCode() == 200 {
		onlineIp = []dispatcher.UserIpList{}
		err := json.Unmarshal(rsp.Body(), &onlineIp)
		if err != nil {
			log.Print(err)
			c.server.ClearOnlineIp(c.Tag)
			return nil
		}
		if c.config.IpRecorderConfig.EnableIpSync {
			c.server.UpdateOnlineIp(c.Tag, onlineIp)
			log.Printf("[Node: %d] Updated %d online ip", c.nodeInfo.NodeId, len(onlineIp))
		}
	} else {
		c.server.ClearOnlineIp(c.Tag)
	}
	return nil
}

func (c *Node) DynamicSpeedLimit() error {
	if c.config.EnableDynamicSpeedLimit {
		for i := range c.userList {
			up, down := c.server.GetUserTraffic(c.buildUserTag(&(c.userList)[i]), false)
			if c.userList[i].Traffic+down+up/1024/1024 > c.config.DynamicSpeedLimitConfig.Traffic {
				err := c.server.UpdateUserSpeedLimit(c.Tag,
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

func (c *Node) buildNodeTag() string {
	return fmt.Sprintf("%s_%s_%d", c.nodeInfo.NodeType, c.config.ListenIP, c.nodeInfo.NodeId)
}
