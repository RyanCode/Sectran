package department

import (
	"context"
	"fmt"

	"sectran_admin/ent/department"
	"sectran_admin/ent/predicate"
	"sectran_admin/internal/svc"
	"sectran_admin/internal/types"
	"sectran_admin/internal/utils/dberrorhandler"

	"entgo.io/ent/dialect/sql"
	"github.com/suyuan32/simple-admin-common/i18n"

	"github.com/suyuan32/simple-admin-common/utils/pointy"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetChildrenDepartmentByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewChildrenDepartmentByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChildrenDepartmentByIdLogic {
	return &GetChildrenDepartmentByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetChildrenDepartmentByIdLogic) GetChildrenDepartmentById(req *types.ChildrenReq) (*types.DepartmentListResp, error) {
	var prefix string
	dept, err := l.svcCtx.DB.Department.Get(l.ctx, req.Id)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, req)
	}

	var predicates []predicate.Department

	//部门名称条件查询
	if req.Name != nil {
		predicates = append(predicates, department.NameContains(*req.Name))
	}

	//部门地区条件查询
	if req.Area != nil {
		predicates = append(predicates, department.AreaContains(*req.Area))
	}

	//部门描述条件查询
	if req.Description != nil {
		predicates = append(predicates, department.DescriptionContains(*req.Description))
	}

	//部门🌲深度，如果是1只查询一级子部门、否则查询所有子部门
	if req.Deep == 1 {
		predicates = append(predicates, department.ParentDepartmentID(req.Id))
	}

	//拼接部门层级
	if len(dept.ParentDepartments) > 0 {
		prefix = fmt.Sprintf("%s,%d", dept.ParentDepartments, dept.ID)
	} else {
		prefix = fmt.Sprint(dept.ID)
	}

	//查询当前部门的子部门
	predicates = append(predicates, department.ParentDepartmentsHasPrefix(prefix))

	deptQuery := l.svcCtx.DB.Department.Query().Where(predicates...)
	data, err := deptQuery.
		Order(department.ByParentDepartments()).
		Order(department.ByID(sql.OrderAsc())).
		Page(l.ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, dberrorhandler.DefaultEntError(l.Logger, err, req)
	}

	resp := &types.DepartmentListResp{}
	resp.Msg = l.svcCtx.Trans.Trans(l.ctx, i18n.Success)
	resp.Data.Total = data.PageDetails.Total

	HasChildren := func(id uint64) bool {
		c, err := l.svcCtx.DB.Department.Query().Where(department.ParentDepartmentID(id)).Count(l.ctx)
		return (err == nil) && c > 0
	}

	for _, v := range data.List {
		resp.Data.Data = append(resp.Data.Data,
			types.DepartmentInfo{
				BaseIDInfo: types.BaseIDInfo{
					Id:        &v.ID,
					CreatedAt: pointy.GetPointer(v.CreatedAt.UnixMilli()),
					UpdatedAt: pointy.GetPointer(v.UpdatedAt.UnixMilli()),
				},
				Name:               &v.Name,
				Area:               &v.Area,
				Description:        &v.Description,
				ParentDepartments:  &v.ParentDepartments,
				ParentDepartmentId: &v.ParentDepartmentID,
				HasChildren:        HasChildren(v.ID),
			})
	}

	return resp, nil
}
