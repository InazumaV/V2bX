package iprecoder

import (
	"github.com/Yuzuki616/V2bX/limiter"
)

type IpRecorder interface {
	SyncOnlineIp(Ips []limiter.UserIpList) ([]limiter.UserIpList, error)
}
