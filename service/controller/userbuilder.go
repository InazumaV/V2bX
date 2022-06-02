package controller

import (
	"fmt"
	"strings"

	"github.com/Yuzuki616/V2bX/api"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/common/serial"
	"github.com/xtls/xray-core/infra/conf"
	"github.com/xtls/xray-core/proxy/shadowsocks"
	"github.com/xtls/xray-core/proxy/trojan"
	"github.com/xtls/xray-core/proxy/vless"
)

func (c *Controller) buildVmessUsers(userInfo *[]api.UserInfo, serverAlterID uint16) (users []*protocol.User) {
	users = make([]*protocol.User, len(*userInfo))
	for i, user := range *userInfo {
		vmessAccount := &conf.VMessAccount{
			ID:       user.V2rayUser.Uuid,
			AlterIds: serverAlterID,
			Security: "auto",
		}
		users[i] = &protocol.User{
			Level:   0,
			Email:   c.buildUserTag(&user), // Email: InboundTag|email|uid
			Account: serial.ToTypedMessage(vmessAccount.Build()),
		}
	}
	return users
}

func (c *Controller) buildVmessUser(userInfo *api.UserInfo, serverAlterID uint16) (user *protocol.User) {
	vmessAccount := &conf.VMessAccount{
		ID:       userInfo.V2rayUser.Uuid,
		AlterIds: serverAlterID,
		Security: "auto",
	}
	user = &protocol.User{
		Level:   0,
		Email:   c.buildUserTag(userInfo), // Email: InboundTag|email|uid
		Account: serial.ToTypedMessage(vmessAccount.Build()),
	}
	return user
}

func (c *Controller) buildVlessUsers(userInfo *[]api.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(*userInfo))
	for i, user := range *userInfo {
		vlessAccount := &vless.Account{
			Id:   user.V2rayUser.Uuid,
			Flow: "xtls-rprx-direct",
		}
		users[i] = &protocol.User{
			Level:   0,
			Email:   c.buildUserTag(&user),
			Account: serial.ToTypedMessage(vlessAccount),
		}
	}
	return users
}

func (c *Controller) buildVlessUser(userInfo *api.UserInfo) (user *protocol.User) {
	vlessAccount := &vless.Account{
		Id:   userInfo.V2rayUser.Uuid,
		Flow: "xtls-rprx-direct",
	}
	user = &protocol.User{
		Level:   0,
		Email:   c.buildUserTag(userInfo),
		Account: serial.ToTypedMessage(vlessAccount),
	}
	return user
}

func (c *Controller) buildTrojanUsers(userInfo *[]api.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(*userInfo))
	for i, user := range *userInfo {
		trojanAccount := &trojan.Account{
			Password: user.V2rayUser.Uuid,
			Flow:     "xtls-rprx-direct",
		}
		users[i] = &protocol.User{
			Level:   0,
			Email:   c.buildUserTag(&user),
			Account: serial.ToTypedMessage(trojanAccount),
		}
	}
	return users
}

func (c *Controller) buildTrojanUser(userInfo *api.UserInfo) (user *protocol.User) {
	trojanAccount := &trojan.Account{
		Password: userInfo.V2rayUser.Uuid,
		Flow:     "xtls-rprx-direct",
	}
	user = &protocol.User{
		Level:   0,
		Email:   c.buildUserTag(userInfo),
		Account: serial.ToTypedMessage(trojanAccount),
	}
	return user
}

func cipherFromString(c string) shadowsocks.CipherType {
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

func (c *Controller) buildSSUsers(userInfo *[]api.UserInfo, method string) (users []*protocol.User) {
	users = make([]*protocol.User, 0)

	cypherMethod := cipherFromString(method)
	for _, user := range *userInfo {
		ssAccount := &shadowsocks.Account{
			Password:   user.Secret,
			CipherType: cypherMethod,
		}
		users = append(users, &protocol.User{
			Level:   0,
			Email:   c.buildUserTag(&user),
			Account: serial.ToTypedMessage(ssAccount),
		})
	}
	return users
}

func (c *Controller) buildSSUser(userInfo *api.UserInfo, method string) (user *protocol.User) {
	cypherMethod := cipherFromString(method)
	ssAccount := &shadowsocks.Account{
		Password:   userInfo.Secret,
		CipherType: cypherMethod,
	}
	user = &protocol.User{
		Level:   0,
		Email:   c.buildUserTag(userInfo),
		Account: serial.ToTypedMessage(ssAccount),
	}
	return user
}

func (c *Controller) buildUserTag(user *api.UserInfo) string {
	return fmt.Sprintf("%s|%s|%d", c.Tag, user.GetUserEmail(), user.UID)
}
