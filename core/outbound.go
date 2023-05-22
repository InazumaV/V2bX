package core

import (
	"context"
	"fmt"
	"github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/features/outbound"
)

func (c *Core) RemoveOutbound(tag string) error {
	err := c.ohm.RemoveHandler(context.Background(), tag)
	return err
}

func (c *Core) AddOutbound(config *core.OutboundHandlerConfig) error {
	rawHandler, err := core.CreateObject(c.Server, config)
	if err != nil {
		return err
	}
	handler, ok := rawHandler.(outbound.Handler)
	if !ok {
		return fmt.Errorf("not an InboundHandler: %s", err)
	}
	if err := c.ohm.AddHandler(context.Background(), handler); err != nil {
		return err
	}
	return nil
}
