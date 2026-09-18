// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package svc

import (
	"demo/internal/config"
	"demo/internal/middleware"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type ServiceContext struct {
	Config                 config.Config
	OrderLimitMiddleware   rest.Middleware
	SendSmsLimitMiddleware rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	rds := redis.MustNewRedis(c.Redis)
	return &ServiceContext{
		Config:                 c,
		OrderLimitMiddleware:   middleware.NewOrdersLimitMiddleware(rds, c.RateLimit.OrdersRate, c.RateLimit.OrdersBurst).Handle,
		SendSmsLimitMiddleware: middleware.NewSendSmsLimitMiddleware(rds, c.RateLimit.SendSmsQuota).Handle,
	}
}
