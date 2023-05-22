package builder

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
	"strings"
)

const xtlsFLow = "xtls-rprx-vision"

func BuildVmessUsers(tag string, userInfo []panel.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i, user := range userInfo {
		users[i] = BuildVmessUser(tag, &user)
	}
	return users
}

func BuildVmessUser(tag string, userInfo *panel.UserInfo) (user *protocol.User) {
	vmessAccount := &conf.VMessAccount{
		ID:       userInfo.Uuid,
		AlterIds: 0,
		Security: "auto",
	}
	return &protocol.User{
		Level:   0,
		Email:   BuildUserTag(tag, userInfo), // Uid: InboundTag|email|uid
		Account: serial.ToTypedMessage(vmessAccount.Build()),
	}
}

func BuildVlessUsers(tag string, userInfo []panel.UserInfo, xtls bool) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i := range userInfo {
		users[i] = BuildVlessUser(tag, &(userInfo)[i], xtls)
	}
	return users
}

func BuildVlessUser(tag string, userInfo *panel.UserInfo, xtls bool) (user *protocol.User) {
	vlessAccount := &vless.Account{
		Id: userInfo.Uuid,
	}
	if xtls {
		vlessAccount.Flow = xtlsFLow
	}
	return &protocol.User{
		Level:   0,
		Email:   BuildUserTag(tag, userInfo),
		Account: serial.ToTypedMessage(vlessAccount),
	}
}

func BuildTrojanUsers(tag string, userInfo []panel.UserInfo) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i := range userInfo {
		users[i] = BuildTrojanUser(tag, &(userInfo)[i])
	}
	return users
}

func BuildTrojanUser(tag string, userInfo *panel.UserInfo) (user *protocol.User) {
	trojanAccount := &trojan.Account{
		Password: userInfo.Uuid,
	}
	return &protocol.User{
		Level:   0,
		Email:   BuildUserTag(tag, userInfo),
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

func BuildSSUsers(tag string, userInfo []panel.UserInfo, cypher shadowsocks.CipherType, serverKey string) (users []*protocol.User) {
	users = make([]*protocol.User, len(userInfo))
	for i := range userInfo {
		users[i] = BuildSSUser(tag, &userInfo[i], cypher, serverKey)
	}
	return users
}

func BuildSSUser(tag string, userInfo *panel.UserInfo, cypher shadowsocks.CipherType, serverKey string) (user *protocol.User) {
	if serverKey == "" {
		ssAccount := &shadowsocks.Account{
			Password:   userInfo.Uuid,
			CipherType: cypher,
		}
		return &protocol.User{
			Level:   0,
			Email:   tag,
			Account: serial.ToTypedMessage(ssAccount),
		}
	} else {
		ssAccount := &shadowsocks_2022.User{
			Key: base64.StdEncoding.EncodeToString([]byte(userInfo.Uuid[:32])),
		}
		return &protocol.User{
			Level:   0,
			Email:   tag,
			Account: serial.ToTypedMessage(ssAccount),
		}
	}
}

func BuildUserTag(tag string, user *panel.UserInfo) string {
	return fmt.Sprintf("%s|%s|%d", tag, user.Uuid, user.Id)
}
