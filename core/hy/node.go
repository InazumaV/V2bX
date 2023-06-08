package hy

import (
	"errors"
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/apernet/hysteria/core/cs"
)

func (h *Hy) AddNode(tag string, info *panel.NodeInfo, c *conf.ControllerConfig) error {
	if info.Type != "hysteria" {
		return errors.New("the core not support " + info.Type)
	}
	s := NewServer(tag)
	err := s.runServer(info, c)
	if err != nil {
		return fmt.Errorf("run hy server error: %s", err)
	}
	h.servers.Store(tag, s)
	return nil
}

func (h *Hy) DelNode(tag string) error {
	if s, e := h.servers.Load(tag); e {
		err := s.(*cs.Server).Close()
		if err != nil {
			return err
		}
		h.servers.Delete(tag)
		return nil
	}
	return errors.New("the node is not have")
}
