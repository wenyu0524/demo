// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package middleware

import (
	"demo/response"
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type OrdersLimitMiddleware struct {
	limiter *limit.TokenLimiter
}

func NewOrdersLimitMiddleware(rds *redis.Redis, rate, burst int) *OrdersLimitMiddleware {
	return &OrdersLimitMiddleware{
		limiter: limit.NewTokenLimiter(rate, burst, rds, "rate:orders"),
	}
}

func (m *OrdersLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !m.limiter.Allow() {
			logx.WithContext(r.Context()).Info("创建订单请求被限流：达到限额")
			response.ResponseStatus(w, http.StatusTooManyRequests, nil, errors.New("请求过于频繁，请稍后再试"))
			return
		}
		next(w, r)
	}
}
