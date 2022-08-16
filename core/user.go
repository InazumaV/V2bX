package core

import (
	"context"
	"fmt"
	"github.com/Yuzuki616/V2bX/core/app/dispatcher"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/features/stats"
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

func (p *Core) GetUserTraffic(email string) (up int64, down int64) {
	upName := "user>>>" + email + ">>>traffic>>>uplink"
	downName := "user>>>" + email + ">>>traffic>>>downlink"
	statsManager := p.Server.GetFeature(stats.ManagerType()).(stats.Manager)
	upCounter := statsManager.GetCounter(upName)
	downCounter := statsManager.GetCounter(downName)
	if upCounter != nil {
		up = upCounter.Value()
		upCounter.Set(0)
	}
	if downCounter != nil {
		down = downCounter.Value()
		downCounter.Set(0)
	}
	return up, down
}

func (p *Core) GetOnlineIps(tag string) ([]dispatcher.UserIp, error) {
	return p.dispatcher.Limiter.GetOnlineUserIp(tag)
}

func (p *Core) UpdateOnlineIps(tag string, ips []dispatcher.UserIp) {
	p.dispatcher.Limiter.UpdateOnlineUserIP(tag, ips)
}

func (p *Core) ClearOnlineIps(tag string) {
	p.dispatcher.Limiter.ClearOnlineUserIP(tag)
}
