package iprecoder

import (
	"context"
	"fmt"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/Yuzuki616/V2bX/limiter"
	"github.com/go-redis/redis/v8"
	"strconv"
	"time"
)

type Redis struct {
	*conf.RedisConfig
	client *redis.Client
}

func NewRedis(c *conf.RedisConfig) *Redis {
	return &Redis{
		RedisConfig: c,
		client: redis.NewClient(&redis.Options{
			Addr:     c.Address,
			Password: c.Password,
			DB:       c.Db,
		}),
	}
}

func (r *Redis) SyncOnlineIp(Ips []limiter.UserIpList) ([]limiter.UserIpList, error) {
	ctx := context.Background()
	for i := range Ips {
		err := r.client.SAdd(ctx, "UserList", Ips[i].Uid).Err()
		if err != nil {
			return nil, fmt.Errorf("add user failed: %s", err)
		}
		r.client.Expire(ctx, "UserList", time.Second*time.Duration(r.Expiry))
		for _, ip := range Ips[i].IpList {
			err := r.client.SAdd(ctx, strconv.Itoa(Ips[i].Uid), ip).Err()
			if err != nil {
				return nil, fmt.Errorf("add ip failed: %s", err)
			}
			r.client.Expire(ctx, strconv.Itoa(Ips[i].Uid), time.Second*time.Duration(r.Expiry))
		}
	}
	c := r.client.SMembers(ctx, "UserList")
	if c.Err() != nil {
		return nil, fmt.Errorf("get user list failed: %s", c.Err())
	}
	Ips = make([]limiter.UserIpList, 0, len(c.Val()))
	for _, uid := range c.Val() {
		uidInt, err := strconv.Atoi(uid)
		if err != nil {
			return nil, fmt.Errorf("convert uid failed: %s", err)
		}
		ips := r.client.SMembers(ctx, uid)
		if ips.Err() != nil {
			return nil, fmt.Errorf("get ip list failed: %s", ips.Err())
		}
		Ips = append(Ips, limiter.UserIpList{
			Uid:    uidInt,
			IpList: ips.Val(),
		})
	}
	return Ips, nil
}
