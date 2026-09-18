package middleware

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func RequestLog(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := logx.WithFields(r.Context(),
			logx.Field("method", r.Method),
			logx.Field("path", r.URL.Path),
		)

		start := time.Now()
		logx.WithContext(ctx).Info("请求开始")
		next(w, r.WithContext(ctx))
		logx.WithContext(ctx).Infof("请求完成耗时= %d ms", time.Since(start).Milliseconds())
	}
}
