package hy

import (
	"encoding/base64"
	"log"
	"testing"
	"time"

	"github.com/Yuzuki616/V2bX/api/panel"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/limiter"
	"github.com/sirupsen/logrus"
)

func TestServer(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	limiter.Init()
	l := limiter.AddLimiter("test", &conf.LimitConfig{}, nil)
	s := NewServer("test", l)
	t.Log(s.runServer(&panel.NodeInfo{
		Port:     1145,
		UpMbps:   100,
		DownMbps: 100,
		HyObfs:   "atresssdaaaadd",
	}, &conf.ControllerConfig{
		ListenIP:  "127.0.0.1",
		HyOptions: conf.HyOptions{},
		CertConfig: &conf.CertConfig{
			CertFile: "../../test_data/1.pem",
			KeyFile:  "../../test_data/1.key",
		},
	}))
	s.users.Store("test1111", struct{}{})
	go func() {
		for {
			time.Sleep(10 * time.Second)
			auth := base64.StdEncoding.EncodeToString([]byte("test1111"))
			log.Println(auth)
			log.Println(s.counter.getCounters(auth).UpCounter.Load())
		}
	}()
	select {}
}
