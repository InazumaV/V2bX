package limiter

import (
	"errors"
	"fmt"
	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/juju/ratelimit"
	"github.com/xtls/xray-core/common/task"
	"log"
	"sync"
	"time"
)

var limitLock sync.RWMutex
var limiter map[string]*Limiter

func Init() {
	limiter = map[string]*Limiter{}
	c := task.Periodic{
		Interval: time.Minute * 2,
		Execute:  ClearPacketOnlineIP,
	}
	go func() {
		log.Println("Limiter: ClearPacketOnlineIP started")
		time.Sleep(time.Minute * 2)
		c.Start()
	}()
}

type Limiter struct {
	Rules         []panel.DestinationRule
	ProtocolRules []string
	SpeedLimit    int
	UserLimitInfo *sync.Map    // Key: Uid value: UserLimitInfo
	ConnLimiter   *ConnLimiter // Key: Uid value: ConnLimiter
	SpeedLimiter  *sync.Map    // key: Uid, value: *ratelimit.Bucket
}

type UserLimitInfo struct {
	UID               int
	SpeedLimit        int
	DynamicSpeedLimit int
	ExpireTime        int64
}

type LimitConfig struct {
	SpeedLimit int
	IpLimit    int
	ConnLimit  int
}

func AddLimiter(tag string, l *LimitConfig, users []panel.UserInfo) *Limiter {
	info := &Limiter{
		SpeedLimit:    l.SpeedLimit,
		UserLimitInfo: new(sync.Map),
		ConnLimiter:   NewConnLimiter(l.ConnLimit, l.IpLimit),
		SpeedLimiter:  new(sync.Map),
	}
	for i := range users {
		if users[i].SpeedLimit != 0 {
			userLimit := &UserLimitInfo{
				UID:        users[i].Id,
				SpeedLimit: users[i].SpeedLimit,
				ExpireTime: 0,
			}
			info.UserLimitInfo.Store(fmt.Sprintf("%s|%s|%d", tag, users[i].Uuid, users[i].Id), userLimit)
		}
	}
	limitLock.Lock()
	limiter[tag] = info
	limitLock.Unlock()
	return info
}

func GetLimiter(tag string) (info *Limiter, err error) {
	limitLock.RLock()
	info, ok := limiter[tag]
	limitLock.RUnlock()
	if !ok {
		return nil, errors.New("not found")
	}
	return
}

func UpdateLimiter(tag string, added []panel.UserInfo, deleted []panel.UserInfo) error {
	l, err := GetLimiter(tag)
	if err != nil {
		return fmt.Errorf("get limit error: %s", err)
	}
	for i := range deleted {
		l.UserLimitInfo.Delete(fmt.Sprintf("%s|%s|%d",
			tag,
			deleted[i].Uuid,
			deleted[i].Id))
	}
	for i := range added {
		if added[i].SpeedLimit != 0 {
			userLimit := &UserLimitInfo{
				UID:        added[i].Id,
				SpeedLimit: added[i].SpeedLimit,
				ExpireTime: 0,
			}
			l.UserLimitInfo.Store(fmt.Sprintf("%s|%s|%d",
				tag,
				added[i].Uuid,
				added[i].Id), userLimit)
		}
	}
	return nil
}

func DeleteLimiter(tag string) {
	limitLock.Lock()
	delete(limiter, tag)
	limitLock.Unlock()
}

func (l *Limiter) CheckLimit(email string, ip string, isTcp bool) (Bucket *ratelimit.Bucket, Reject bool) {
	// ip and conn limiter
	if l.ConnLimiter.AddConnCount(email, ip, isTcp) {
		return nil, true
	}
	// check and gen speed limit Bucket
	nodeLimit := l.SpeedLimit
	userLimit := 0
	if v, ok := l.UserLimitInfo.Load(email); ok {
		u := v.(*UserLimitInfo)
		if u.ExpireTime < time.Now().Unix() && u.ExpireTime != 0 {
			if u.SpeedLimit != 0 {
				userLimit = u.SpeedLimit
				u.DynamicSpeedLimit = 0
				u.ExpireTime = 0
			} else {
				l.UserLimitInfo.Delete(email)
			}
		} else {
			userLimit = determineSpeedLimit(u.SpeedLimit, u.DynamicSpeedLimit)
		}
	}
	limit := int64(determineSpeedLimit(nodeLimit, userLimit)) * 1000000 / 8 // If you need the Speed limit
	if limit > 0 {
		Bucket = ratelimit.NewBucketWithQuantum(time.Second, limit, limit) // Byte/s
		if v, ok := l.SpeedLimiter.LoadOrStore(email, Bucket); ok {
			return v.(*ratelimit.Bucket), false
		} else {
			l.SpeedLimiter.Store(email, Bucket)
			return Bucket, false
		}
	} else {
		return nil, false
	}
}

type UserIpList struct {
	Uid    int      `json:"Uid"`
	IpList []string `json:"Ips"`
}

func determineDeviceLimit(nodeLimit, userLimit int) (limit int) {
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
