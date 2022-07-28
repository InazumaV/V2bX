package xray

import (
	"context"
	"fmt"
	"github.com/Yuzuki616/V2bX/app/dispatcher"
	"github.com/Yuzuki616/V2bX/common/limiter"
	"github.com/xtls/xray-core/common/protocol"
	"github.com/xtls/xray-core/features/inbound"
	"github.com/xtls/xray-core/features/routing"
	"github.com/xtls/xray-core/features/stats"
	"github.com/xtls/xray-core/proxy"
)

func (p *Xray) GetUserManager(tag string) (proxy.UserManager, error) {
	inboundManager := p.Server.GetFeature(inbound.ManagerType()).(inbound.Manager)
	handler, err := inboundManager.GetHandler(context.Background(), tag)
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

func (p *Xray) AddUsers(users []*protocol.User, tag string) error {
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

func (p *Xray) RemoveUsers(users []string, tag string) error {
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

func (p *Xray) GetUserTraffic(email string) (up int64, down int64) {
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

func (p *Xray) GetOnlineIps(tag string) (*[]limiter.UserIp, error) {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	return dispather.Limiter.GetOnlineUserIp(tag)
}

func (p *Xray) UpdateOnlineIps(tag string, ips *[]limiter.UserIp) {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	dispather.Limiter.UpdateOnlineUserIP(tag, ips)
}

func (p *Xray) ClearOnlineIps(tag string) {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	dispather.Limiter.ClearOnlineUserIP(tag)
}
