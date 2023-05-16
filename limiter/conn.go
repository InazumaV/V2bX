package limiter

import (
	"sync"
)

type ConnLimiter struct {
	ipLimit   int
	connLimit int
	count     sync.Map //map[string]int
	ip        sync.Map //map[string]map[string]*sync.Map
}

func NewConnLimiter(conn int, ip int) *ConnLimiter {
	return &ConnLimiter{
		connLimit: conn,
		ipLimit:   ip,
		count:     sync.Map{},
		ip:        sync.Map{},
	}
}

func (c *ConnLimiter) AddConnCount(user string, ip string) (limit bool) {
	if c.connLimit != 0 {
		if v, ok := c.count.Load(user); ok {
			if v.(int) >= c.connLimit {
				return true
			} else {
				c.count.Store(user, v.(int)+1)
			}
		} else {
			c.count.Store(user, 1)
		}
	}
	if c.ipLimit == 0 {
		return false
	}
	ipMap := new(sync.Map)
	ipMap.Store(ip, 1)
	if v, ok := c.ip.LoadOrStore(user, ipMap); ok {
		// have online ip
		ips := v.(*sync.Map)
		cn := 0
		if online, ok := ips.Load(ip); !ok {
			ips.Range(func(key, value interface{}) bool {
				cn++
				if cn >= c.ipLimit {
					limit = true
					return false
				}
				return true
			})
			if limit {
				return
			}
			ips.Store(ip, 1)
		} else {
			// have this ip
			ips.Store(ip, online.(int)+1)
		}
	}
	return false
}

func (c *ConnLimiter) DelConnCount(user string, ip string) {
	if c.connLimit != 0 {
		if v, ok := c.count.Load(user); ok {
			if v.(int) == 1 {
				c.count.Delete(user)
			} else {
				c.count.Store(user, v.(int)-1)
			}
		}
	}
	if c.ipLimit == 0 {
		return
	}
	if i, ok := c.ip.Load(user); ok {
		is := i.(*sync.Map)
		if i, ok := is.Load(ip); ok {
			if i.(int) == 1 {
				is.Delete(ip)
			} else {
				is.Store(user, i.(int)-1)
			}
			notDel := false
			c.ip.Range(func(_, _ any) bool {
				notDel = true
				return true
			})
			if !notDel {
				c.ip.Delete(user)
			}
		}
	}
}
