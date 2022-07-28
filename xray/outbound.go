package xray

import (
	"context"
	"fmt"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/outbound"
)

func (p *Xray) RemoveOutbound(tag string) error {
	outboundManager := p.Server.GetFeature(outbound.ManagerType()).(outbound.Manager)
	err := outboundManager.RemoveHandler(context.Background(), tag)
	return err
}

func (p *Xray) AddOutbound(config *core.OutboundHandlerConfig) error {
	outboundManager := p.Server.GetFeature(outbound.ManagerType()).(outbound.Manager)
	rawHandler, err := core.CreateObject(p.Server, config)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(outbound.Handler)
	if !ok {
		return fmt.Errorf("not an InboundHandler: %s", err)
	}
	if err := outboundManager.AddHandler(context.Background(), handler); err != nil {
		return err
	}
	return nil
}
