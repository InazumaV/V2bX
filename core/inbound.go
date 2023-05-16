package core

import (
	"context"
	"fmt"
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
