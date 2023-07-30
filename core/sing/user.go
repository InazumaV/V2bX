package sing

import (
	"errors"

	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/common/counter"
	"github.com/InazumaV/V2bX/core"
	"github.com/inazumav/sing-box/inbound"
	"github.com/inazumav/sing-box/option"
)

func (b *Box) AddUsers(p *core.AddUsersParams) (added int, err error) {
	switch p.NodeInfo.Type {
	case "v2ray":
		if p.NodeInfo.ExtraConfig.EnableVless == "true" {
			us := make([]option.VLESSUser, len(p.UserInfo))
			for i := range p.UserInfo {
				us[i] = option.VLESSUser{
					Name: p.UserInfo[i].Uuid,
					Flow: p.NodeInfo.ExtraConfig.VlessFlow,
					UUID: p.UserInfo[i].Uuid,
				}
			}
			err = b.inbounds[p.Tag].(*inbound.VLESS).AddUsers(us)
		} else {
			us := make([]option.VMessUser, len(p.UserInfo))
			for i := range p.UserInfo {
				us[i] = option.VMessUser{
					Name: p.UserInfo[i].Uuid,
					UUID: p.UserInfo[i].Uuid,
				}
			}
			err = b.inbounds[p.Tag].(*inbound.VMess).AddUsers(us)
		}
	case "shadowsocks":
		us := make([]option.ShadowsocksUser, len(p.UserInfo))
		for i := range p.UserInfo {
			us[i] = option.ShadowsocksUser{
				Name:     p.UserInfo[i].Uuid,
				Password: p.UserInfo[i].Uuid,
			}
		}
	}
	if err != nil {
		return 0, err
	}
	return len(p.UserInfo), err
}

func (b *Box) GetUserTraffic(tag, uuid string, reset bool) (up int64, down int64) {
	if v, ok := b.hookServer.Hooker().counter.Load(tag); ok {
		c := v.(*counter.TrafficCounter)
		up = c.GetUpCount(uuid)
		down = c.GetDownCount(uuid)
		if reset {
			c.Reset(uuid)
		}
		return
	}
	return 0, 0
}

type UserDeleter interface {
	DelUsers(uuid []string) error
}

func (b *Box) DelUsers(users []panel.UserInfo, tag string) error {
	var del UserDeleter
	if i, ok := b.inbounds[tag]; ok {
		switch i.Type() {
		case "vmess":
			del = i.(*inbound.VMess)
		case "shadowsocks":
			del = i.(*inbound.ShadowsocksMulti)
		}
	} else {
		return errors.New("the inbound not found")
	}
	uuids := make([]string, len(users))
	for i := range users {
		uuids[i] = users[i].Uuid
	}
	err := del.DelUsers(uuids)
	if err != nil {
		return err
	}
	return nil
}
