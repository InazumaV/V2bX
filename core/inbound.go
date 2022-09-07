package core

import (
	"context"
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/core/app/dispatcher"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/inbound"
)

func (p *Core) RemoveInbound(tag string) error {
	return p.ihm.RemoveHandler(context.Background(), tag)
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

func (p *Core) AddInboundLimiter(tag string, nodeInfo *panel.NodeInfo, userList []panel.UserInfo) error {
	return p.dispatcher.Limiter.AddInboundLimiter(tag, nodeInfo, userList)
}

func (p *Core) GetInboundLimiter(tag string) (*dispatcher.InboundInfo, error) {
	limit, ok := p.dispatcher.Limiter.InboundInfo.Load(tag)
	if ok {
		return limit.(*dispatcher.InboundInfo), nil
	}
	return nil, fmt.Errorf("not found limiter")
}

func (p *Core) UpdateInboundLimiter(tag string, nodeInfo *panel.NodeInfo, added, deleted []panel.UserInfo) error {
	return p.dispatcher.Limiter.UpdateInboundLimiter(tag, nodeInfo, added, deleted)
}

func (p *Core) DeleteInboundLimiter(tag string) error {
	return p.dispatcher.Limiter.DeleteInboundLimiter(tag)
}
