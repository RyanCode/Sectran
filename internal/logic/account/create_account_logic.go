package account

import (
	"context"

	"sectran_admin/ent"
	"sectran_admin/internal/svc"
	"sectran_admin/internal/types"
	"sectran_admin/internal/utils/dberrorhandler"

	"github.com/suyuan32/simple-admin-common/i18n"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAccountLogic {
	return &CreateAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateAccountLogic) CreateAccount(req *types.AccountInfo) (*types.BaseMsgResp, error) {
	if err := ModifyCheckout(l.svcCtx, l.ctx, req); err != nil {
		return nil, err
	}

	device, err := l.svcCtx.DB.Device.Get(l.ctx, *req.DeviceId)
	if err != nil {
		if _, ok := err.(*ent.NotFoundError); ok {
			return nil, types.CustomError("设备不存在")
		}
		return nil, types.ErrInternalError
	}

	_, err = l.svcCtx.DB.Account.Create().
		SetNotNilUsername(req.Username).
		SetNotNilPort(req.Port).
		SetNotNilProtocol(req.Protocol).
		SetNotNilPassword(req.Password).
		SetNotNilPrivateKey(req.PrivateKey).
		SetNotNilPrivateKeyPassword(req.PrivateKeyPassword).
		SetNotNilDeviceID(req.DeviceId).
		SetDepartmentID(device.DepartmentID).
		Save(l.ctx)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, req)
	}

	return &types.BaseMsgResp{Msg: l.svcCtx.Trans.Trans(l.ctx, i18n.CreateSuccess)}, nil
}
