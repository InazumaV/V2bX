package xray

import (
	"context"
	"fmt"
	"github.com/Yuzuki616/V2bX/api"
	"github.com/Yuzuki616/V2bX/app/dispatcher"
	"github.com/Yuzuki616/V2bX/common/limiter"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/inbound"
	"github.com/xtls/xray-core/features/routing"
)

func (p *Xray) RemoveInbound(tag string) error {
	inboundManager := p.Server.GetFeature(inbound.ManagerType()).(inbound.Manager)
	err := inboundManager.RemoveHandler(context.Background(), tag)
	return err
}

func (p *Xray) AddInbound(config *core.InboundHandlerConfig) error {
	inboundManager := p.Server.GetFeature(inbound.ManagerType()).(inbound.Manager)
	rawHandler, err := core.CreateObject(p.Server, config)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(inbound.Handler)
	if !ok {
		return fmt.Errorf("not an InboundHandler: %s", err)
	}
	if err := inboundManager.AddHandler(context.Background(), handler); err != nil {
		return err
	}
	return nil
}

func (p *Xray) AddInboundLimiter(tag string, nodeInfo *api.NodeInfo, userList *[]api.UserInfo) error {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	err := dispather.Limiter.AddInboundLimiter(tag, nodeInfo, userList)
	return err
}

func (p *Xray) GetInboundLimiter(tag string) (*limiter.InboundInfo, error) {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	limit, ok := dispather.Limiter.InboundInfo.Load(tag)
	if ok {
		return limit.(*limiter.InboundInfo), nil
	}
	return nil, fmt.Errorf("not found limiter")
}

func (p *Xray) UpdateInboundLimiter(tag string, nodeInfo *api.NodeInfo, updatedUserList *[]api.UserInfo) error {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	err := dispather.Limiter.UpdateInboundLimiter(tag, nodeInfo, updatedUserList)
	return err
}

func (p *Xray) DeleteInboundLimiter(tag string) error {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	err := dispather.Limiter.DeleteInboundLimiter(tag)
	return err
}
