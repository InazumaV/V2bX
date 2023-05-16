package iprecoder

import (
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/limiter"
	"log"
	"testing"
)

func TestRedis_SyncOnlineIp(t *testing.T) {
	r := NewRedis(&conf.RedisConfig{
		Address:  "127.0.0.1:6379",
		Password: "",
		Db:       0,
	})
	users, err := r.SyncOnlineIp([]limiter.UserIpList{
		{1, []string{"5.5.5.5", "4.4.4.4"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	log.Println(users)
}
