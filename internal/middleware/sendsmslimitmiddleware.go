// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package middleware

import (
	"bytes"
	"demo/response"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type SendSmsLimitMiddleware struct {
	limiter *limit.PeriodLimit
}

func NewSendSmsLimitMiddleware(rds *redis.Redis, quota int) *SendSmsLimitMiddleware {
	return &SendSmsLimitMiddleware{
		limiter: limit.NewPeriodLimit(60, quota, rds, "rate:sendsms"),
	}
}

func (m *SendSmsLimitMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logx.WithContext(r.Context()).Errorf("发送验证码请求失败：读取请求体出错：%v", err)
			response.ResponseStatus(w, http.StatusInternalServerError, nil, errors.New("服务器内部错误"))
			return
		}
		// 将已经读取完的r.Body重新设置成一个可读的Body
		r.Body = io.NopCloser(bytes.NewReader(body))

		// 解析JSON
		var req struct {
			Phone string `json:"phone"`
		}
		err = json.Unmarshal(body, &req)
		if err != nil {
			logx.WithContext(r.Context()).Infof("发送验证码请求失败：请求体JSON格式错误：%v", err)
			response.ResponseStatus(w, http.StatusBadRequest, nil, errors.New("请求格式错误"))
			return
		}

		key := req.Phone
		if key == "" {
			logx.WithContext(r.Context()).Info("发送验证码请求失败：手机号为空")
			response.ResponseStatus(w, http.StatusBadRequest, nil, errors.New("手机号不能为空"))
			return
		}

		// 根据手机号限流
		result, err := m.limiter.TakeCtx(r.Context(), key)
		if err != nil {
			logx.WithContext(r.Context()).Errorf("发送验证码请求失败：限流器执行异常：%v", err)
			response.ResponseStatus(w, http.StatusInternalServerError, nil, errors.New("服务暂时不可用"))
			return
		}

		switch result {
		case limit.Allowed, limit.HitQuota:
			next(w, r)
		case limit.OverQuota:
			logx.WithContext(r.Context()).Info("发送验证码请求被限流：达到限额")
			response.ResponseStatus(w, http.StatusTooManyRequests, nil, errors.New("请求过于频繁，请稍后再试"))
			return
		default:
			logx.WithContext(r.Context()).Errorf("发送验证码请求失败：未知的限流结果：%v", result)
			response.ResponseStatus(w, http.StatusInternalServerError, nil, errors.New("服务器内部错误"))
			return
		}
	}
}
