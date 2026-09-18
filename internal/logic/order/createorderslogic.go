// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package order

import (
	"context"
	"demo/internal/svc"
	"demo/internal/types"
	"demo/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateOrderLogic) CreateOrder(req *types.CreateOrderRequest) (resp *types.CreateOrderResponse, err error) {
	// 模拟生成12位订单号
	orderId := utils.RandomNumberString(12)

	l.Logger.Infof("创建订单成功：userId=%s, productId=%s", req.UserId, req.ProductId)

	return &types.CreateOrderResponse{
		OrderId: orderId,
		Message: "支付成功",
	}, nil
}
