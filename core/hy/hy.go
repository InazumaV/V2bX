package hy

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/hashicorp/go-multierror"
	"sync"
)

type Hy struct {
	servers sync.Map
}

func New(_ *conf.CoreConfig) (*Hy, error) {
	return &Hy{
		servers: sync.Map{},
	}, nil
}

func (h *Hy) Start() error {
	return nil
}

func (h *Hy) Close() error {
	var errs error
	h.servers.Range(func(tag, s any) bool {
		err := s.(*Server).Close()
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("close %s error: %s", tag, err))
		}
		return true
	})
	if errs != nil {
		return errs
	}
	return nil
}
