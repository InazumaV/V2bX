package hy

import (
	"fmt"
	"sync"

	"github.com/Yuzuki616/V2bX/conf"
	vCore "github.com/Yuzuki616/V2bX/core"
	"github.com/hashicorp/go-multierror"
)

func init() {
	vCore.RegisterCore("hy", NewHy)
}

type Hy struct {
	servers sync.Map
}

func NewHy(_ *conf.CoreConfig) (vCore.Core, error) {
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

func (h *Hy) Protocols() []string {
	return []string{
		"hysteria",
	}
}
