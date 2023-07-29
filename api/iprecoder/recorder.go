package iprecoder

import (
	"errors"
	"time"

	"github.com/InazumaV/V2bX/conf"
	"github.com/InazumaV/V2bX/limiter"
	"github.com/go-resty/resty/v2"
	"github.com/goccy/go-json"
)

type Recorder struct {
	client *resty.Client
	*conf.RecorderConfig
}

func NewRecorder(c *conf.RecorderConfig) *Recorder {
	return &Recorder{
		client:         resty.New().SetTimeout(time.Duration(c.Timeout) * time.Second),
		RecorderConfig: c,
	}
}

func (r *Recorder) SyncOnlineIp(ips []limiter.UserIpList) ([]limiter.UserIpList, error) {
	rsp, err := r.client.R().
		SetBody(ips).
		Post(r.Url + "/api/v1/SyncOnlineIp?token=" + r.Token)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode() != 200 {
		return nil, errors.New(rsp.String())
	}
	ips = []limiter.UserIpList{}
	err = json.Unmarshal(rsp.Body(), &ips)
	if err != nil {
		return nil, err
	}
	return ips, nil
}
