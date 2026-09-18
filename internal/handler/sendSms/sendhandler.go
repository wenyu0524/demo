package handler

import (
	"demo/internal/logic/sendSms"
	"demo/internal/svc"
	"demo/internal/types"
	"demo/response"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SendHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendSmsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := sendSms.NewSendLogic(r.Context(), svcCtx)
		resp, err := l.Send(&req)
		response.Response(w, resp, err)

	}
}
