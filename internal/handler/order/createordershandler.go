package handler

import (
	"demo/internal/logic/order"
	"demo/internal/svc"
	"demo/internal/types"
	"demo/response"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreateOrdersHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CreateOrderRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}

		l := order.NewCreateOrdersLogic(r.Context(), svcCtx)
		resp, err := l.CreateOrder(&req)
		response.Response(w, resp, err)

	}
}
