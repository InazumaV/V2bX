package iprecoder

import (
	"errors"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/core/app/dispatcher"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
	"time"
)

type Recorder struct {
	client *resty.Client
	*conf.RecorderConfig
}

func New(c *conf.RecorderConfig) *Recorder {
	return &Recorder{
		client:         resty.New().SetTimeout(time.Duration(c.Timeout) * time.Second),
		RecorderConfig: c,
	}
}

func (r *Recorder) SyncOnlineIp(ips []dispatcher.UserIpList) ([]dispatcher.UserIpList, error) {
	rsp, err := r.client.R().
		SetBody(ips).
		Post(r.Url + "/api/v1/SyncOnlineIp?token=" + r.Token)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode() != 200 {
		return nil, errors.New(rsp.String())
	}
	ips = []dispatcher.UserIpList{}
	err = json.Unmarshal(rsp.Body(), &ips)
	if err != nil {
		return nil, err
	}
	return ips, nil
}
