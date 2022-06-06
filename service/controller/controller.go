package controller

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"runtime"
	"time"

	"github.com/Yuzuki616/V2bX/api"
	"github.com/Yuzuki616/V2bX/common/legoCmd"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/task"
	"github.com/xtls/xray-core/core"
)

type Controller struct {
	server                  *core.Instance
	config                  *Config
	clientInfo              api.ClientInfo
	apiClient               api.API
	nodeInfo                *api.NodeInfo
	Tag                     string
	userList                *[]api.UserInfo
	nodeInfoMonitorPeriodic *task.Periodic
	userReportPeriodic      *task.Periodic
	panelType               string
}

// New return a Controller service with default parameters.
func New(server *core.Instance, api api.API, config *Config) *Controller {
	controller := &Controller{
		server:    server,
		config:    config,
		apiClient: api,
	}
	return controller
}

// Start implement the Start() function of the service interface
func (c *Controller) Start() error {
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
	if err := c.AddInboundLimiter(c.Tag, userInfo); err != nil {
		log.Print(err)
	}
	// Add Rule Manager
	if !c.config.DisableGetRule {
		if ruleList, err := c.apiClient.GetNodeRule(); err != nil {
			log.Printf("Get rule list filed: %s", err)
		} else if len(*ruleList) > 0 {
			if err := c.UpdateRule(c.Tag, *ruleList); err != nil {
				log.Print(err)
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
	runtime.GC()
	return nil
}

// Close implement the Close() function of the service interface
func (c *Controller) Close() error {
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
	return nil
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
		if !reflect.DeepEqual(c.nodeInfo, newNodeInfo) {
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
			if err = c.DeleteInboundLimiter(oldtag); err != nil {
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
			if err := c.UpdateRule(c.Tag, *ruleList); err != nil {
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
		// Xray-core supports the OcspStapling certification hot renew
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
		if newUserInfo != nil {
			c.userList = newUserInfo
		}
		newUserInfo = nil
		err = c.addNewUser(c.userList, newNodeInfo)
		if err != nil {
			log.Print(err)
			return nil
		}
		newNodeInfo = nil
		// Add Limiter
		if err := c.AddInboundLimiter(c.Tag, c.userList); err != nil {
			log.Print(err)
			return nil
		}
		runtime.GC()
	} else if newUserInfo != nil {
		deleted, added := compareUserList(c.userList, newUserInfo)
		if len(deleted) > 0 {
			deletedEmail := make([]string, len(deleted))
			for i := range deleted {
				deletedEmail[i] = fmt.Sprintf("%s|%s|%d", c.Tag,
					(*c.userList)[deleted[i]].GetUserEmail(),
					(*c.userList)[deleted[i]].UID)
			}
			err := c.removeUsers(deletedEmail, c.Tag)
			if err != nil {
				log.Print(err)
			}
		}
		if len(added) > 0 {
			err = c.addNewUserFromIndex(newUserInfo, &added, c.nodeInfo)
			if err != nil {
				log.Print(err)
			}
			// Update Limiter
			if err := c.UpdateInboundLimiter(c.Tag, newUserInfo, &added); err != nil {
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

func (c *Controller) removeOldTag(oldtag string) (err error) {
	err = c.removeInbound(oldtag)
	if err != nil {
		return err
	}
	err = c.removeOutbound(oldtag)
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) addNewTag(newNodeInfo *api.NodeInfo) (err error) {
	inboundConfig, err := InboundBuilder(c.config, newNodeInfo, c.Tag)
	if err != nil {
		return err
	}
	err = c.addInbound(inboundConfig)
	if err != nil {

		return err
	}
	outBoundConfig, err := OutboundBuilder(c.config, newNodeInfo, c.Tag)
	if err != nil {

		return err
	}
	err = c.addOutbound(outBoundConfig)
	if err != nil {

		return err
	}
	return nil
}

func (c *Controller) addNewUser(userInfo *[]api.UserInfo, nodeInfo *api.NodeInfo) (err error) {
	users := make([]*protocol.User, 0)
	if nodeInfo.NodeType == "V2ray" {
		if nodeInfo.EnableVless {
			users = c.buildVlessUsers(userInfo)
		} else {
			alterID := 0
			alterID = (*userInfo)[0].V2rayUser.AlterId
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
		users = c.buildSSUsers(userInfo, nodeInfo.SS.CypherMethod)
	} else {
		return fmt.Errorf("unsupported node type: %s", nodeInfo.NodeType)
	}
	err = c.addUsers(users, c.Tag)
	if err != nil {
		return err
	}
	log.Printf("[%s: %d] Added %d new users", c.nodeInfo.NodeType, c.nodeInfo.NodeId, len(*userInfo))
	return nil
}

func (c *Controller) addNewUserFromIndex(userInfo *[]api.UserInfo, userIndex *[]int, nodeInfo *api.NodeInfo) (err error) {
	users := make([]*protocol.User, 0, len(*userIndex))
	for _, v := range *userIndex {
		if nodeInfo.NodeType == "V2ray" {
			if nodeInfo.EnableVless {
				users = append(users, c.buildVlessUser(&(*userInfo)[v]))
			} else {
				alterID := 0
				alterID = (*userInfo)[0].V2rayUser.AlterId
				if alterID >= 0 && alterID < math.MaxUint16 {
					users = append(users, c.buildVmessUser(&(*userInfo)[v], uint16(alterID)))
				} else {
					users = append(users, c.buildVmessUser(&(*userInfo)[v], 0))
					return fmt.Errorf("AlterID should between 0 to 1<<16 - 1, set it to 0 for now")
				}
			}
		} else if nodeInfo.NodeType == "Trojan" {
			users = append(users, c.buildTrojanUser(&(*userInfo)[v]))
		} else if nodeInfo.NodeType == "Shadowsocks" {
			users = append(users, c.buildSSUser(&(*userInfo)[v], nodeInfo.SS.CypherMethod))
		} else {
			return fmt.Errorf("unsupported node type: %s", nodeInfo.NodeType)
		}
	}
	err = c.addUsers(users, c.Tag)
	if err != nil {
		return err
	}
	log.Printf("[%s: %d] Added %d new users", c.nodeInfo.NodeType, c.nodeInfo.NodeId, len(*userIndex))
	return nil
}

func compareUserList(old, new *[]api.UserInfo) (deleted, added []int) {
	tmp := map[int]struct{}{}
	tmp2 := map[int]struct{}{}
	for i := range *old {
		tmp[(*old)[i].UID] = struct{}{}
	}
	l := len(tmp)
	for i := range *new {
		tmp[(*new)[i].UID] = struct{}{}
		tmp2[(*new)[i].UID] = struct{}{}
		if l != len(tmp) {
			added = append(added, i)
			l++
		}
	}
	tmp = nil
	l = len(tmp2)
	for i := range *old {
		tmp2[(*old)[i].UID] = struct{}{}
		if l != len(tmp2) {
			deleted = append(deleted, i)
			l++
		}
	}
	return deleted, added
}

func (c *Controller) userInfoMonitor() (err error) {
	// Get User traffic
	userTraffic := make([]api.UserTraffic, 0)
	for i := range *c.userList {
		up, down := c.getTraffic(c.buildUserTag(&(*c.userList)[i]))
		if up > 0 || down > 0 {
			userTraffic = append(userTraffic, api.UserTraffic{
				UID:      (*c.userList)[i].UID,
				Upload:   up,
				Download: down})
		}
	}
	if len(userTraffic) > 0 && !c.config.DisableUploadTraffic {
		err = c.apiClient.ReportUserTraffic(&userTraffic)
		if err != nil {
			log.Print(err)
		}
	}

	// Report Online info
	if onlineDevice, err := c.GetOnlineDevice(c.Tag); err != nil {
		log.Print(err)
	} else {
		log.Printf("[%s: %d] Report %d online users", c.nodeInfo.NodeType, c.nodeInfo.NodeId, len(*onlineDevice))
	}
	// Report Illegal user
	if detectResult, err := c.GetDetectResult(c.Tag); err != nil {
		log.Print(err)
	} else {
		log.Printf("[%s: %d] Report %d illegal behaviors", c.nodeInfo.NodeType, c.nodeInfo.NodeId, len(*detectResult))
	}
	userTraffic = nil
	runtime.GC()
	return nil
}

func (c *Controller) buildNodeTag() string {
	return fmt.Sprintf("%s_%s_%d", c.nodeInfo.NodeType, c.config.ListenIP, c.nodeInfo.NodeId)
}
