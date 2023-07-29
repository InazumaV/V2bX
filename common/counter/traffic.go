package counter

import (
	"sync"
	"sync/atomic"
)

type TrafficCounter struct {
	counters map[string]*TrafficStorage
	lock     sync.RWMutex
}

type TrafficStorage struct {
	UpCounter   atomic.Int64
	DownCounter atomic.Int64
}

func NewTrafficCounter() *TrafficCounter {
	return &TrafficCounter{
		counters: map[string]*TrafficStorage{},
	}
}

func (c *TrafficCounter) GetCounter(id string) *TrafficStorage {
	c.lock.RLock()
	cts, ok := c.counters[id]
	c.lock.RUnlock()
	if !ok {
		cts = &TrafficStorage{}
		c.counters[id] = cts
	}
	return cts
}

func (c *TrafficCounter) GetUpCount(id string) int64 {
	c.lock.RLock()
	cts, ok := c.counters[id]
	c.lock.RUnlock()
	if ok {
		return cts.UpCounter.Load()
	}
	return 0
}

func (c *TrafficCounter) GetDownCount(id string) int64 {
	c.lock.RLock()
	cts, ok := c.counters[id]
	c.lock.RUnlock()
	if ok {
		return cts.DownCounter.Load()
	}
	return 0
}

func (c *TrafficCounter) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return len(c.counters)
}

func (c *TrafficCounter) Reset(id string) {
	c.lock.RLock()
	cts := c.GetCounter(id)
	c.lock.RUnlock()
	cts.UpCounter.Store(0)
	cts.DownCounter.Store(0)
}

func (c *TrafficCounter) Delete(id string) {
	c.lock.Lock()
	delete(c.counters, id)
	c.lock.Unlock()
}

func (c *TrafficCounter) Rx(id string, n int) {
	cts := c.GetCounter(id)
	cts.DownCounter.Add(int64(n))
}

func (c *TrafficCounter) Tx(id string, n int) {
	cts := c.GetCounter(id)
	cts.UpCounter.Add(int64(n))
}

func (c *TrafficCounter) IncConn(auth string) {
	return
}

func (c *TrafficCounter) DecConn(auth string) {
	return
}
