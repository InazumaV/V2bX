package core

import (
	"context"
	"fmt"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/outbound"
)

func (p *Core) RemoveOutbound(tag string) error {
	err := p.ohm.RemoveHandler(context.Background(), tag)
	return err
}

func (p *Core) AddOutbound(config *core.OutboundHandlerConfig) error {
	rawHandler, err := core.CreateObject(p.Server, config)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(outbound.Handler)
	if !ok {
		return fmt.Errorf("not an InboundHandler: %s", err)
	}
	if err := p.ohm.AddHandler(context.Background(), handler); err != nil {
		return err
	}
	return nil
}
