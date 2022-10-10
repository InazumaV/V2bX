package controller

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"strings"

	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/proxy/shadowsocks"
	"github.com/xtls/xray-core/proxy/trojan"
	"github.com/xtls/xray-core/proxy/vless"
)

func (c *Node) buildVmessUsers(userInfo []panel.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i, user := range userInfo {
		users[i] = c.buildVmessUser(&user, 0)
	}
	return users
}

func (c *Node) buildVmessUser(userInfo *panel.UserInfo, serverAlterID uint16) (user *protocol.User) {
	vmessAccount := &conf.VMessAccount{
		ID:       userInfo.V2rayUser.Uuid,
		AlterIds: serverAlterID,
		Security: "auto",
	}
	return &protocol.User{
		Level:   0,
		Email:   c.buildUserTag(userInfo), // Uid: InboundTag|email|uid
		Account: serial.ToTypedMessage(vmessAccount.Build()),
	}
}

func (c *Node) buildVlessUsers(userInfo []panel.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i := range userInfo {
		users[i] = c.buildVlessUser(&(userInfo)[i])
	}
	return users
}

func (c *Node) buildVlessUser(userInfo *panel.UserInfo) (user *protocol.User) {
	vlessAccount := &vless.Account{
		Id:   userInfo.V2rayUser.Uuid,
		Flow: "xtls-rprx-direct",
	}
	return &protocol.User{
		Level:   0,
		Email:   c.buildUserTag(userInfo),
		Account: serial.ToTypedMessage(vlessAccount),
	}
}

func (c *Node) buildTrojanUsers(userInfo []panel.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i := range userInfo {
		users[i] = c.buildTrojanUser(&(userInfo)[i])
	}
	return users
}

func (c *Node) buildTrojanUser(userInfo *panel.UserInfo) (user *protocol.User) {
	trojanAccount := &trojan.Account{
		Password: userInfo.TrojanUser.Password,
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

func (c *Node) buildSSUsers(userInfo []panel.UserInfo, cypher shadowsocks.CipherType) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i := range userInfo {
		c.buildSSUser(&(userInfo)[i], cypher)
	}
	return users
}

func (c *Node) buildSSUser(userInfo *panel.UserInfo, cypher shadowsocks.CipherType) (user *protocol.User) {
	ssAccount := &shadowsocks.Account{
		Password:   userInfo.Secret,
		CipherType: cypher,
	}
	return &protocol.User{
		Level:   0,
		Email:   c.buildUserTag(userInfo),
		Account: serial.ToTypedMessage(ssAccount),
	}
}

func (c *Node) buildUserTag(user *panel.UserInfo) string {
	return fmt.Sprintf("%s|%s|%d", c.Tag, user.GetUserEmail(), user.UID)
}
