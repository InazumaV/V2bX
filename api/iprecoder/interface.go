package iprecoder

import (
	"github.com/InazumaV/V2bX/limiter"
)

type IpRecorder interface {
	SyncOnlineIp(Ips []limiter.UserIpList) ([]limiter.UserIpList, error)
}
