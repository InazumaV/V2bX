package hy

import (
	"github.com/apernet/hysteria/core/cs"
	"sync"
	"sync/atomic"
)

type UserTrafficCounter struct {
	counters map[string]*counters
	lock     sync.RWMutex
}

type counters struct {
	UpCounter   atomic.Int64
	DownCounter atomic.Int64
	//ConnGauge   atomic.Int64
}

func NewUserTrafficCounter() cs.TrafficCounter {
	return new(UserTrafficCounter)
}

func (c *UserTrafficCounter) getCounters(auth string) *counters {
	c.lock.RLock()
	cts, ok := c.counters[auth]
	c.lock.RUnlock()
	if !ok {
		cts = &counters{}
		c.counters[auth] = cts
	}
	return cts
}

func (c *UserTrafficCounter) Rx(auth string, n int) {
	cts := c.getCounters(auth)
	cts.DownCounter.Add(int64(n))
}

func (c *UserTrafficCounter) Tx(auth string, n int) {
	cts := c.getCounters(auth)
	cts.UpCounter.Add(int64(n))
}

func (c *UserTrafficCounter) IncConn(_ string) {
	/*cts := c.getCounters(auth)
	cts.ConnGauge.Add(1)*/
	return
}

func (c *UserTrafficCounter) DecConn(_ string) {
	/*cts := c.getCounters(auth)
	cts.ConnGauge.Add(1)*/
	return
}

func (c *UserTrafficCounter) Reset(auth string) {
	cts := c.getCounters(auth)
	cts.UpCounter.Store(0)
	cts.DownCounter.Store(0)
}

func (c *UserTrafficCounter) Delete(auth string) {
	c.lock.Lock()
	delete(c.counters, auth)
	c.lock.Unlock()
}
