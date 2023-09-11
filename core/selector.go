package core

import (
	"errors"
	"fmt"
	"sync"

	"github.com/InazumaV/V2bX/api/panel"
	"github.com/InazumaV/V2bX/conf"
	"github.com/hashicorp/go-multierror"
)

type Selector struct {
	cores []Core
	nodes sync.Map
}

func (s *Selector) Start() error {
	for i := range s.cores {
		err := s.cores[i].Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Selector) Close() error {
	var errs error
	for i := range s.cores {
		errs = multierror.Append(errs, s.cores[i].Close())
	}
	return errs
}

func isSupported(protocol string, protocols []string) bool {
	for i := range protocols {
		if protocol == protocols[i] {
			return true
		}
	}
	return false
}

func (s *Selector) AddNode(tag string, info *panel.NodeInfo, option *conf.Options) error {
	for i, c := range s.cores {
		if len(option.Core) == 0 {
			if !isSupported(info.Type, c.Protocols()) {
				continue
			}
			option.Core = c.Type()
			err := option.UnmarshalJSON(option.RawOptions)
			if err != nil {
				return fmt.Errorf("unmarshal option error: %s", err)
			}
		} else if option.Core != c.Type() {
			continue
		}
		err := c.AddNode(tag, info, option)
		if err != nil {
			return err
		}
		s.nodes.Store(tag, i)
		return nil
	}
	return errors.New("the node type is not support")
}

func (s *Selector) DelNode(tag string) error {
	if t, e := s.nodes.Load(tag); e {
		err := s.cores[t.(int)].DelNode(tag)
		if err != nil {
			return err
		}
		s.nodes.Delete(tag)
		return nil
	}
	return errors.New("the node is not have")
}

func (s *Selector) AddUsers(p *AddUsersParams) (added int, err error) {
	t, e := s.nodes.Load(p.Tag)
	if !e {
		return 0, errors.New("the node is not have")
	}
	return s.cores[t.(int)].AddUsers(p)
}

func (s *Selector) GetUserTraffic(tag, uuid string, reset bool) (up int64, down int64) {
	t, e := s.nodes.Load(tag)
	if !e {
		return 0, 0
	}
	return s.cores[t.(int)].GetUserTraffic(tag, uuid, reset)
}

func (s *Selector) DelUsers(users []panel.UserInfo, tag string) error {
	t, e := s.nodes.Load(tag)
	if !e {
		return errors.New("the node is not have")
	}
	return s.cores[t.(int)].DelUsers(users, tag)
}

func (s *Selector) Protocols() []string {
	protocols := make([]string, 0)
	for i := range s.cores {
		protocols = append(protocols, s.cores[i].Protocols()...)
	}
	return protocols
}

func (s *Selector) Type() string {
	t := "Selector("
	for i := range s.cores {
		if i != 0 {
			t += " "
		}
		t += s.cores[i].Type()
	}
	t += ")"
	return t
}
