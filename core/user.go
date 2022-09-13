package core

import (
	"context"
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/core/app/dispatcher"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/proxy"
)

func (p *Core) GetUserManager(tag string) (proxy.UserManager, error) {
	handler, err := p.ihm.GetHandler(context.Background(), tag)
	if err != nil {
		return nil, fmt.Errorf("no such inbound tag: %s", err)
	}
	inboundInstance, ok := handler.(proxy.GetInbound)
	if !ok {
		return nil, fmt.Errorf("handler %s is not implement proxy.GetInbound", tag)
	}
	userManager, ok := inboundInstance.GetInbound().(proxy.UserManager)
	if !ok {
		return nil, fmt.Errorf("handler %s is not implement proxy.UserManager", tag)
	}
	return userManager, nil
}

func (p *Core) AddUsers(users []*protocol.User, tag string) error {
	userManager, err := p.GetUserManager(tag)
	if err != nil {
		return fmt.Errorf("get user manager error: %s", err)
	}
	for _, item := range users {
		mUser, err := item.ToMemoryUser()
		if err != nil {
			return err
		}
		err = userManager.AddUser(context.Background(), mUser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Core) RemoveUsers(users []string, tag string) error {
	userManager, err := p.GetUserManager(tag)
	if err != nil {
		return fmt.Errorf("get user manager error: %s", err)
	}
	for _, email := range users {
		err = userManager.RemoveUser(context.Background(), email)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Core) GetUserTraffic(email string, reset bool) (up int64, down int64) {
	upName := "user>>>" + email + ">>>traffic>>>uplink"
	downName := "user>>>" + email + ">>>traffic>>>downlink"
	upCounter := p.shm.GetCounter(upName)
	downCounter := p.shm.GetCounter(downName)
	if reset {
		if upCounter != nil {
			up = upCounter.Set(0)
		}
		if downCounter != nil {
			down = downCounter.Set(0)
		}
	} else {
		if upCounter != nil {
			up = upCounter.Value()
		}
		if downCounter != nil {
			down = downCounter.Value()
		}
	}
	return up, down
}

func (p *Core) AddUserSpeedLimit(tag string, user *panel.UserInfo, speedLimit uint64, expire int64) error {
	return p.dispatcher.Limiter.AddUserSpeedLimit(tag, user, speedLimit, expire)
}

func (p *Core) ListOnlineIp(tag string) ([]dispatcher.UserIpList, error) {
	return p.dispatcher.Limiter.ListOnlineUserIp(tag)
}

func (p *Core) UpdateOnlineIp(tag string, ips []dispatcher.UserIpList) {
	p.dispatcher.Limiter.UpdateOnlineUserIP(tag, ips)
}

func (p *Core) ClearOnlineIp(tag string) {
	p.dispatcher.Limiter.ClearOnlineUserIpAndSpeedLimiter(tag)
}
