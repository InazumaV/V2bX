package node

import (
	"encoding/base64"
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/proxy/shadowsocks"
	"github.com/xtls/xray-core/proxy/shadowsocks_2022"
	"github.com/xtls/xray-core/proxy/trojan"
	"github.com/xtls/xray-core/proxy/vless"
	"log"
	"strings"
)

func (c *Controller) addNewUser(userInfo []panel.UserInfo, nodeInfo *panel.NodeInfo) (err error) {
	users := make([]*protocol.User, 0, len(userInfo))
	switch nodeInfo.NodeType {
	case "v2ray":
		if c.EnableVless {
			users = c.buildVlessUsers(userInfo)
		} else {
			users = c.buildVmessUsers(userInfo)
		}
	case "Trojan":
		users = c.buildTrojanUsers(userInfo)
	case "Shadowsocks":
		users = c.buildSSUsers(userInfo, getCipherFromString(nodeInfo.Cipher))
	default:
		return fmt.Errorf("unsupported node type: %s", nodeInfo.NodeType)
	}
	err = c.server.AddUsers(users, c.Tag)
	if err != nil {
		return fmt.Errorf("add users error: %s", err)
	}
	log.Printf("[%s: %d] Added %d new users", c.nodeInfo.NodeType, c.nodeInfo.NodeId, len(userInfo))
	return nil
}

func (c *Controller) buildVmessUsers(userInfo []panel.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i, user := range userInfo {
		users[i] = c.buildVmessUser(&user, 0)
	}
	return users
}

func (c *Controller) buildVmessUser(userInfo *panel.UserInfo, serverAlterID uint16) (user *protocol.User) {
	vmessAccount := &conf.VMessAccount{
		ID:       userInfo.Uuid,
		AlterIds: serverAlterID,
		Security: "auto",
	}
	return &protocol.User{
		Level:   0,
		Email:   c.buildUserTag(userInfo), // Uid: InboundTag|email|uid
		Account: serial.ToTypedMessage(vmessAccount.Build()),
	}
}

func (c *Controller) buildVlessUsers(userInfo []panel.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i := range userInfo {
		users[i] = c.buildVlessUser(&(userInfo)[i])
	}
	return users
}

func (c *Controller) buildVlessUser(userInfo *panel.UserInfo) (user *protocol.User) {
	vlessAccount := &vless.Account{
		Id:   userInfo.Uuid,
		Flow: "xtls-rprx-direct",
	}
	return &protocol.User{
		Level:   0,
		Email:   c.buildUserTag(userInfo),
		Account: serial.ToTypedMessage(vlessAccount),
	}
}

func (c *Controller) buildTrojanUsers(userInfo []panel.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i := range userInfo {
		users[i] = c.buildTrojanUser(&(userInfo)[i])
	}
	return users
}

func (c *Controller) buildTrojanUser(userInfo *panel.UserInfo) (user *protocol.User) {
	trojanAccount := &trojan.Account{
		Password: userInfo.Uuid,
		Flow:     "xtls-rprx-direct",
	}
	return &protocol.User{
		Level:   0,
		Email:   c.buildUserTag(userInfo),
		Account: serial.ToTypedMessage(trojanAccount),
	}
}

func getCipherFromString(c string) shadowsocks.CipherType {
	switch strings.ToLower(c) {
	case "aes-128-gcm", "aead_aes_128_gcm":
		return shadowsocks.CipherType_AES_128_GCM
	case "aes-256-gcm", "aead_aes_256_gcm":
		return shadowsocks.CipherType_AES_256_GCM
	case "chacha20-poly1305", "aead_chacha20_poly1305", "chacha20-ietf-poly1305":
		return shadowsocks.CipherType_CHACHA20_POLY1305
	case "none", "plain":
		return shadowsocks.CipherType_NONE
	default:
		return shadowsocks.CipherType_UNKNOWN
	}
}

func (c *Controller) buildSSUsers(userInfo []panel.UserInfo, cypher shadowsocks.CipherType) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i := range userInfo {
		users[i] = c.buildSSUser(&(userInfo)[i], cypher)
	}
	return users
}

func (c *Controller) buildSSUser(userInfo *panel.UserInfo, cypher shadowsocks.CipherType) (user *protocol.User) {
	if c.nodeInfo.ServerKey == "" {
		ssAccount := &shadowsocks.Account{
			Password:   userInfo.Uuid,
			CipherType: cypher,
		}
		return &protocol.User{
			Level:   0,
			Email:   c.buildUserTag(userInfo),
			Account: serial.ToTypedMessage(ssAccount),
		}
	} else {
		ssAccount := &shadowsocks_2022.User{
			Key: base64.StdEncoding.EncodeToString([]byte(userInfo.Uuid[:32])),
		}
		return &protocol.User{
			Level:   0,
			Email:   c.buildUserTag(userInfo),
			Account: serial.ToTypedMessage(ssAccount),
		}
	}
}

func (c *Controller) buildUserTag(user *panel.UserInfo) string {
	return fmt.Sprintf("%s|%s|%d", c.Tag, user.Uuid, user.Id)
}
