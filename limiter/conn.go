package limiter

import (
	"sync"
	"time"
)

type ConnLimiter struct {
	realtime  bool
	ipLimit   int
	connLimit int
	count     sync.Map // map[string]int
	ip        sync.Map // map[string]map[string]int
}

func NewConnLimiter(conn int, ip int, realtime bool) *ConnLimiter {
	return &ConnLimiter{
		realtime:  realtime,
		connLimit: conn,
		ipLimit:   ip,
		count:     sync.Map{},
		ip:        sync.Map{},
	}
}

func (c *ConnLimiter) AddConnCount(user string, ip string, isTcp bool) (limit bool) {
	if c.connLimit != 0 {
		if v, ok := c.count.Load(user); ok {
			if v.(int) >= c.connLimit {
				// over connection limit
				return true
			} else if isTcp {
				// tcp protocol
				// connection count add
				c.count.Store(user, v.(int)+1)
			}
		} else if isTcp {
			// tcp protocol
			// store connection count
			c.count.Store(user, 1)
		}
	}
	if c.ipLimit == 0 {
		return false
	}
	// first user map
	ipMap := new(sync.Map)
	if c.realtime {
		if isTcp {
			ipMap.Store(ip, 2)
		} else {
			ipMap.Store(ip, 1)
		}
	} else {
		ipMap.Store(ip, time.Now())
	}
	// check user online ip
	if v, ok := c.ip.LoadOrStore(user, ipMap); ok {
		// have user
		ips := v.(*sync.Map)
		cn := 0
		if online, ok := ips.Load(ip); ok {
			// online ip
			if c.realtime {
				if isTcp {
					// tcp count add
					ips.Store(ip, online.(int)+2)
				}
			} else {
				// update connect time for not realtime
				ips.Store(ip, time.Now())
			}
		} else {
			// not online ip
			ips.Range(func(_, _ interface{}) bool {
				cn++
				if cn >= c.ipLimit {
					limit = true
					return false
				}
				return true
			})
			if limit {
				// over ip limit
				return
			}
			if c.realtime {
				if isTcp {
					ips.Store(ip, 2)
				} else {
					ips.Store(ip, 1)
				}
			} else {
				ips.Store(ip, time.Now())
			}
		}
	}
	return
}

// DelConnCount Delete tcp connection count, no tcp do not use
func (c *ConnLimiter) DelConnCount(user string, ip string) {
	if !c.realtime {
		return
	}
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
			if i.(int) == 2 {
				is.Delete(ip)
			} else {
				is.Store(user, i.(int)-2)
			}
			notDel := false
			c.ip.Range(func(_, _ any) bool {
				notDel = true
				return false
			})
			if !notDel {
				c.ip.Delete(user)
			}
		}
	}
}

// ClearOnlineIP Clear udp,icmp and other packet protocol online ip
func (c *ConnLimiter) ClearOnlineIP() {
	c.ip.Range(func(u, v any) bool {
		userIp := v.(*sync.Map)
		notDel := false
		userIp.Range(func(ip, v any) bool {
			notDel = true
			if _, ok := v.(int); ok {
				if v.(int) == 1 {
					// clear packet ip for realtime
					userIp.Delete(ip)
				}
				return true
			} else {
				// clear ip for not realtime
				if v.(time.Time).Before(time.Now().Add(time.Minute)) {
					// 1 minute no active
					userIp.Delete(ip)
				}
			}
			return true
		})
		if !notDel {
			c.ip.Delete(u)
		}
		return true
	})
}
