// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package sendSms

import (
	"context"
	"demo/utils"

	"demo/internal/svc"
	"demo/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendLogic {
	return &SendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendLogic) Send(req *types.SendSmsRequest) (resp *types.SendSmsResponse, err error) {
	// 模拟生成6位验证码
	code := utils.RandomNumberString(6)

	l.Logger.Infof("发送验证码成功：phone=%s", req.Phone)

	return &types.SendSmsResponse{
		Code: code,
	}, nil
}
