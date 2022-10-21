package iprecoder

import "github.com/Yuzuki616/V2bX/core/app/dispatcher"

type IpRecorder interface {
	SyncOnlineIp(Ips []dispatcher.UserIpList) ([]dispatcher.UserIpList, error)
}
