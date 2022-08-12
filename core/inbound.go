package core

import (
	"context"
	"fmt"
	"github.com/Yuzuki616/V2bX/api"
	"github.com/Yuzuki616/V2bX/app/limiter"
	"github.com/Yuzuki616/V2bX/core/app/dispatcher"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/inbound"
	"github.com/xtls/xray-core/features/routing"
)

func (p *Core) RemoveInbound(tag string) error {
	err := p.ihm.RemoveHandler(context.Background(), tag)
	return err
}

func (p *Core) AddInbound(config *core.InboundHandlerConfig) error {
	rawHandler, err := core.CreateObject(p.Server, config)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(inbound.Handler)
	if !ok {
		return fmt.Errorf("not an InboundHandler: %s", err)
	}
	if err := p.ihm.AddHandler(context.Background(), handler); err != nil {
		return err
	}
	return nil
}

func (p *Core) AddInboundLimiter(tag string, nodeInfo *api.NodeInfo, userList []api.UserInfo) error {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	err := dispather.Limiter.AddInboundLimiter(tag, nodeInfo, userList)
	return err
}

func (p *Core) GetInboundLimiter(tag string) (*limiter.InboundInfo, error) {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	limit, ok := dispather.Limiter.InboundInfo.Load(tag)
	if ok {
		return limit.(*limiter.InboundInfo), nil
	}
	return nil, fmt.Errorf("not found limiter")
}

func (p *Core) UpdateInboundLimiter(tag string, nodeInfo *api.NodeInfo, updatedUserList []api.UserInfo) error {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	err := dispather.Limiter.UpdateInboundLimiter(tag, nodeInfo, updatedUserList)
	return err
}

func (p *Core) DeleteInboundLimiter(tag string) error {
	dispather := p.Server.GetFeature(routing.DispatcherType()).(*dispatcher.DefaultDispatcher)
	err := dispather.Limiter.DeleteInboundLimiter(tag)
	return err
}
