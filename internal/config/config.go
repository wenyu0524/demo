// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	RateLimit struct {
		OrdersRate   int
		OrdersBurst  int
		SendSmsQuota int
	}

	Redis redis.RedisConf
}
