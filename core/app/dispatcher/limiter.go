package dispatcher

import (
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/juju/ratelimit"
	"github.com/xtls/xray-core/common"
	"github.com/xtls/xray-core/common/buf"
	"io"
	"sync"
	"time"
)

type UserInfo struct {
	UID         int
	SpeedLimit  uint64
	ExpireTime  int64
	DeviceLimit int
}

type InboundInfo struct {
	Tag            string
	NodeSpeedLimit uint64
	UserInfo       *sync.Map // Key: Uid value: UserInfo
	SpeedLimiter   *sync.Map // key: Uid, value: *ratelimit.Bucket
	UserOnlineIP   *sync.Map // Key: Uid Value: *sync.Map: Key: IP, Value: bool
}

type Limiter struct {
	InboundInfo *sync.Map // Key: Tag, Value: *InboundInfo
}

func NewLimiter() *Limiter {
	return &Limiter{
		InboundInfo: new(sync.Map),
	}
}

func (l *Limiter) AddInboundLimiter(tag string, nodeInfo *panel.NodeInfo, userList []panel.UserInfo) error {
	inboundInfo := &InboundInfo{
		Tag:            tag,
		NodeSpeedLimit: nodeInfo.SpeedLimit,
		SpeedLimiter:   new(sync.Map),
		UserOnlineIP:   new(sync.Map),
	}
	userMap := new(sync.Map)
	for i := range userList {
		/*if (*userList)[i].SpeedLimit == 0 {
			(*userList)[i].SpeedLimit = nodeInfo.SpeedLimit
		}
		if (*userList)[i].DeviceLimit == 0 {
			(*userList)[i].DeviceLimit = nodeInfo.DeviceLimit
		}*/
		userMap.Store(fmt.Sprintf("%s|%s|%d", tag, (userList)[i].V2rayUser.Email, (userList)[i].UID),
			&UserInfo{
				UID:         (userList)[i].UID,
				SpeedLimit:  nodeInfo.SpeedLimit,
				DeviceLimit: nodeInfo.DeviceLimit,
			})
	}
	inboundInfo.UserInfo = userMap
	l.InboundInfo.Store(tag, inboundInfo) // Replace the old inbound info
	return nil
}

func (l *Limiter) UpdateInboundLimiter(tag string, nodeInfo *panel.NodeInfo, added, deleted []panel.UserInfo) error {
	if value, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := value.(*InboundInfo)
		// Update User info
		for i := range added {
			inboundInfo.UserInfo.Store(fmt.Sprintf("%s|%s|%d", tag,
				(added)[i].V2rayUser.Email, (added)[i].UID), &UserInfo{
				UID:         (added)[i].UID,
				SpeedLimit:  nodeInfo.SpeedLimit,
				DeviceLimit: nodeInfo.DeviceLimit,
			})
		}
		for i := range deleted {
			inboundInfo.UserInfo.Delete(fmt.Sprintf("%s|%s|%d", tag,
				(deleted)[i].V2rayUser.Email, (deleted)[i].UID))
			inboundInfo.SpeedLimiter.Delete(fmt.Sprintf("%s|%s|%d", tag,
				(deleted)[i].V2rayUser.Email, (deleted)[i].UID)) // Delete limiter bucket
		}
	} else {
		return fmt.Errorf("no such inbound in limiter: %s", tag)
	}
	return nil
}

func (l *Limiter) DeleteInboundLimiter(tag string) error {
	l.InboundInfo.Delete(tag)
	return nil
}

func (l *Limiter) UpdateUserSpeedLimit(tag string, userInfo *panel.UserInfo, limit uint64, expire int64) error {
	if value, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := value.(*InboundInfo)
		if user, ok := inboundInfo.UserInfo.Load(fmt.Sprintf("%s|%s|%d", tag, userInfo.GetUserEmail(), userInfo.UID)); ok {
			user.(*UserInfo).SpeedLimit = limit
			user.(*UserInfo).ExpireTime = time.Now().Add(time.Duration(expire) * time.Second).Unix()
			inboundInfo.SpeedLimiter.Delete(fmt.Sprintf("%s|%s|%d", tag, userInfo.GetUserEmail(), userInfo.UID))
		} else {
			return fmt.Errorf("no such user in limiter: %s", userInfo.GetUserEmail())
		}
		return nil
	} else {
		return fmt.Errorf("no such inbound in limiter: %s", tag)
	}
}

type UserIpList struct {
	Uid    int      `json:"Uid"`
	IpList []string `json:"Ips"`
}

func (l *Limiter) ListOnlineUserIp(tag string) ([]UserIpList, error) {
	if value, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := value.(*InboundInfo)
		onlineUser := make([]UserIpList, 0)
		var ipMap *sync.Map
		inboundInfo.UserOnlineIP.Range(func(key, value interface{}) bool {
			ipMap = value.(*sync.Map)
			var ip []string
			ipMap.Range(func(key, v interface{}) bool {
				if v.(bool) {
					ip = append(ip, key.(string))
				}
				return true
			})
			if len(ip) > 0 {
				if u, ok := inboundInfo.UserInfo.Load(key.(string)); ok {
					onlineUser = append(onlineUser, UserIpList{
						Uid:    u.(*UserInfo).UID,
						IpList: ip,
					})
				}
			}
			return true
		})
		if len(onlineUser) == 0 {
			return nil, nil
		}
		return onlineUser, nil
	} else {
		return nil, fmt.Errorf("no such inbound in limiter: %s", tag)
	}
}

func (l *Limiter) UpdateOnlineUserIP(tag string, userIpList []UserIpList) {
	if v, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := v.(*InboundInfo)
		//Clear old IP
		inboundInfo.UserOnlineIP.Range(func(key, value interface{}) bool {
			inboundInfo.UserOnlineIP.Delete(key)
			return true
		})
		// Update User Online IP
		for i := range userIpList {
			ipMap := new(sync.Map)
			for _, userIp := range (userIpList)[i].IpList {
				ipMap.Store(userIp, false)
			}
			inboundInfo.UserOnlineIP.Store((userIpList)[i].Uid, ipMap)
		}
		inboundInfo.SpeedLimiter.Range(func(key, value interface{}) bool {
			if _, exists := inboundInfo.UserOnlineIP.Load(key.(string)); !exists {
				inboundInfo.SpeedLimiter.Delete(key.(string))
			}
			return true
		})
	}
}

func (l *Limiter) ClearOnlineUserIpAndSpeedLimiter(tag string) {
	if v, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := v.(*InboundInfo)
		inboundInfo.SpeedLimiter.Range(func(key, value interface{}) bool {
			if _, exists := inboundInfo.UserOnlineIP.Load(key.(string)); !exists {
				inboundInfo.SpeedLimiter.Delete(key.(string))
			}
			return true
		})
		inboundInfo.UserOnlineIP.Range(func(key, value interface{}) bool {
			inboundInfo.UserOnlineIP.Delete(key)
			return true
		})
	}
}

func (l *Limiter) CheckSpeedAndDeviceLimit(tag string, email string, ip string) (speedLimiter *ratelimit.Bucket, SpeedLimit bool, Reject bool) {
	if value, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := value.(*InboundInfo)
		nodeLimit := inboundInfo.NodeSpeedLimit
		var userLimit uint64 = 0
		var deviceLimit = 0
		expired := false
		if v, ok := inboundInfo.UserInfo.Load(email); ok {
			u := v.(*UserInfo)
			if u.ExpireTime < time.Now().Unix() && u.ExpireTime != 0 {
				userLimit = 0
				expired = true
			} else {
				userLimit = u.SpeedLimit
			}
			deviceLimit = u.DeviceLimit
		}
		ipMap := new(sync.Map)
		ipMap.Store(ip, true)
		// If any device is online
		if v, ok := inboundInfo.UserOnlineIP.LoadOrStore(email, ipMap); ok {
			ipMap := v.(*sync.Map)
			// If this ip is a new device
			if online, ok := ipMap.LoadOrStore(ip, true); !ok {
				counter := 0
				ipMap.Range(func(key, value interface{}) bool {
					counter++
					return true
				})
				if counter > deviceLimit && deviceLimit > 0 {
					ipMap.Delete(ip)
					return nil, false, true
				}
			} else {
				if !online.(bool) {
					ipMap.Store(ip, true)
				}
			}
		}
		limit := determineRate(nodeLimit, userLimit) // If need the Speed limit
		if limit > 0 {
			limiter := ratelimit.NewBucketWithQuantum(time.Second, int64(limit), int64(limit)) // Byte/s
			if v, ok := inboundInfo.SpeedLimiter.LoadOrStore(email, limiter); ok {
				if expired {
					inboundInfo.SpeedLimiter.Store(email, limiter)
					return limiter, true, false
				}
				bucket := v.(*ratelimit.Bucket)
				return bucket, true, false
			} else {
				return limiter, true, false
			}
		} else {
			return nil, false, false
		}
	} else {
		newError("Get Inbound Limiter information failed").AtDebug().WriteToLog()
		return nil, false, false
	}
}

type Writer struct {
	writer  buf.Writer
	limiter *ratelimit.Bucket
	w       io.Writer
}

func (l *Limiter) RateWriter(writer buf.Writer, limiter *ratelimit.Bucket) buf.Writer {
	return &Writer{
		writer:  writer,
		limiter: limiter,
	}
}

func (w *Writer) Close() error {
	return common.Close(w.writer)
}

func (w *Writer) WriteMultiBuffer(mb buf.MultiBuffer) error {
	w.limiter.Wait(int64(mb.Len()))
	return w.writer.WriteMultiBuffer(mb)
}

// determineRate returns the minimum non-zero rate
func determineRate(nodeLimit, userLimit uint64) (limit uint64) {
	if nodeLimit == 0 || userLimit == 0 {
		if nodeLimit > userLimit {
			return nodeLimit
		} else if nodeLimit < userLimit {
			return userLimit
		} else {
			return 0
		}
	} else {
		if nodeLimit > userLimit {
			return userLimit
		} else if nodeLimit < userLimit {
			return nodeLimit
		} else {
			return nodeLimit
		}
	}
}
