// Package limiter is to control the links that go into the dispather
package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/Yuzuki616/V2bX/api"
	"github.com/juju/ratelimit"
)

type UserInfo struct {
	UID         int
	SpeedLimit  uint64
	DeviceLimit int
}

type InboundInfo struct {
	Tag            string
	NodeSpeedLimit uint64
	UserInfo       *sync.Map // Key: Email value: UserInfo
	BucketHub      *sync.Map // key: Email, value: *ratelimit.Bucket
	UserOnlineIP   *sync.Map // Key: Email Value: *sync.Map: Key: IP, Value: UID
}

type Limiter struct {
	InboundInfo *sync.Map // Key: Tag, Value: *InboundInfo
}

func New() *Limiter {
	return &Limiter{
		InboundInfo: new(sync.Map),
	}
}

func (l *Limiter) AddInboundLimiter(tag string, nodeInfo *api.NodeInfo, userList *[]api.UserInfo) error {
	inboundInfo := &InboundInfo{
		Tag:            tag,
		NodeSpeedLimit: nodeInfo.SpeedLimit,
		BucketHub:      new(sync.Map),
		UserOnlineIP:   new(sync.Map),
	}
	userMap := new(sync.Map)
	for i := range *userList {
		if (*userList)[i].SpeedLimit == 0 {
			(*userList)[i].SpeedLimit = nodeInfo.SpeedLimit
		}
		if (*userList)[i].DeviceLimit == 0 {
			(*userList)[i].DeviceLimit = nodeInfo.DeviceLimit
		}
		userMap.Store(fmt.Sprintf("%s|%d|%d", tag, (*userList)[i].Port, (*userList)[i].UID), UserInfo{
			UID:         (*userList)[i].UID,
			SpeedLimit:  (*userList)[i].SpeedLimit,
			DeviceLimit: (*userList)[i].DeviceLimit,
		})
	}
	inboundInfo.UserInfo = userMap
	l.InboundInfo.Store(tag, inboundInfo) // Replace the old inbound info
	return nil
}

func (l *Limiter) UpdateInboundLimiter(tag string, nodeInfo *api.NodeInfo, updatedUserList *[]api.UserInfo, usersIndex *[]int) error {

	if value, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := value.(*InboundInfo)
		// Update User info
		for _, u := range *usersIndex {
			if (*updatedUserList)[u].SpeedLimit == 0 {
				(*updatedUserList)[u].SpeedLimit = nodeInfo.SpeedLimit
			}
			if (*updatedUserList)[u].DeviceLimit == 0 {
				(*updatedUserList)[u].DeviceLimit = nodeInfo.DeviceLimit
			}
			inboundInfo.UserInfo.Store(fmt.Sprintf("%s|%s|%d", tag, (*updatedUserList)[u].GetUserEmail(), (*updatedUserList)[u].UID), UserInfo{
				UID:         (*updatedUserList)[u].UID,
				SpeedLimit:  (*updatedUserList)[u].SpeedLimit,
				DeviceLimit: (*updatedUserList)[u].DeviceLimit,
			})
			inboundInfo.BucketHub.Delete(fmt.Sprintf("%s|%s|%d", tag, (*updatedUserList)[u].GetUserEmail(), (*updatedUserList)[u].UID)) // Delete old limiter bucket
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

func (l *Limiter) GetOnlineDevice(tag string) (*[]api.OnlineUser, error) {
	onlineUser := make([]api.OnlineUser, 0)
	if value, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := value.(*InboundInfo)
		// Clear Speed Limiter bucket for users who are not online
		inboundInfo.BucketHub.Range(func(key, value interface{}) bool {
			email := key.(string)
			if _, exists := inboundInfo.UserOnlineIP.Load(email); !exists {
				inboundInfo.BucketHub.Delete(email)
			}
			return true
		})
		inboundInfo.UserOnlineIP.Range(func(key, value interface{}) bool {
			ipMap := value.(*sync.Map)
			ipMap.Range(func(key, value interface{}) bool {
				ip := key.(string)
				uid := value.(int)
				onlineUser = append(onlineUser, api.OnlineUser{UID: uid, IP: ip})
				return true
			})
			email := key.(string)
			inboundInfo.UserOnlineIP.Delete(email) // Reset online device
			return true
		})
	} else {
		return nil, fmt.Errorf("no such inbound in limiter: %s", tag)
	}
	return &onlineUser, nil
}

func (l *Limiter) GetUserBucket(tag string, email string, ip string) (limiter *ratelimit.Bucket, SpeedLimit bool, Reject bool) {
	if value, ok := l.InboundInfo.Load(tag); ok {
		inboundInfo := value.(*InboundInfo)
		nodeLimit := inboundInfo.NodeSpeedLimit
		var userLimit uint64 = 0
		var deviceLimit = 0
		var uid = 0
		if v, ok := inboundInfo.UserInfo.Load(email); ok {
			u := v.(UserInfo)
			uid = u.UID
			userLimit = u.SpeedLimit
			deviceLimit = u.DeviceLimit
		}
		// Report online device
		ipMap := new(sync.Map)
		ipMap.Store(ip, uid)
		// If any device is online
		if v, ok := inboundInfo.UserOnlineIP.LoadOrStore(email, ipMap); ok {
			ipMap := v.(*sync.Map)
			// If this ip is a new device
			if _, ok := ipMap.LoadOrStore(ip, uid); !ok {
				counter := 0
				ipMap.Range(func(key, value interface{}) bool {
					counter++
					return true
				})
				if counter > deviceLimit && deviceLimit > 0 {
					ipMap.Delete(ip)
					return nil, false, true
				}
			}
		}
		limit := determineRate(nodeLimit, userLimit) // If need the Speed limit
		if limit > 0 {
			limiter := ratelimit.NewBucketWithQuantum(time.Second, int64(limit), int64(limit)) // Byte/s
			if v, ok := inboundInfo.BucketHub.LoadOrStore(email, limiter); ok {
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
