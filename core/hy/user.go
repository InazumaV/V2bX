package hy

import (
	"errors"
	"github.com/Yuzuki616/V2bX/core"
)

func (h *Hy) AddUsers(p *core.AddUsersParams) (int, error) {
	s, ok := h.servers.Load(p.Tag)
	if !ok {
		return 0, errors.New("the node not have")
	}
	u := &s.(*Server).users
	for i := range p.UserInfo {
		u.Store(p.UserInfo[i].Uuid, struct{}{})
	}
	return len(p.UserInfo), nil
}

func (h *Hy) GetUserTraffic(tag, uuid string, reset bool) (up int64, down int64) {
	s, _ := h.servers.Load(tag)
	c := &s.(*Server).counter
	up = c.getCounters(uuid).UpCounter.Load()
	down = c.getCounters(uuid).DownCounter.Load()
	if reset {
		c.Reset(uuid)
	}
	return
}

func (h *Hy) DelUsers(users []string, tag string) error {
	v, e := h.servers.Load(tag)
	if !e {
		return errors.New("the node is not have")
	}
	s := v.(*Server)
	for i := range users {
		s.users.Delete(users[i])
		s.counter.Delete(users[i])
	}
	return nil
}
