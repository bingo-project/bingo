package system

import (
	"context"
	"regexp"

	"github.com/jinzhu/copier"

	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	"bingo/internal/pkg/model"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1"
)

type MenuBiz interface {
	List(ctx context.Context, req *v1.ListMenuRequest) (*v1.ListMenuResponse, error)
	Create(ctx context.Context, req *v1.CreateMenuRequest) (*v1.MenuInfo, error)
	Get(ctx context.Context, ID uint) (*v1.MenuInfo, error)
	Update(ctx context.Context, ID uint, req *v1.UpdateMenuRequest) (*v1.MenuInfo, error)
	Delete(ctx context.Context, ID uint) error

	All(ctx context.Context) ([]*v1.MenuInfo, error)
	Tree(ctx context.Context) ([]*v1.MenuInfo, error)
	ToggleHidden(ctx context.Context, ID uint) (*v1.MenuInfo, error)
}

type menuBiz struct {
	ds store.IStore
}

var _ MenuBiz = (*menuBiz)(nil)

func NewMenu(ds store.IStore) *menuBiz {
	return &menuBiz{ds: ds}
}

func (b *menuBiz) List(ctx context.Context, req *v1.ListMenuRequest) (*v1.ListMenuResponse, error) {
	count, list, err := b.ds.SysMenu().ListWithRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list menus", "err", err)

		return nil, err
	}

	data := make([]v1.MenuInfo, 0, len(list))
	for _, item := range list {
		var menu v1.MenuInfo
		_ = copier.Copy(&menu, item)

		data = append(data, menu)
	}

	return &v1.ListMenuResponse{Total: count, Data: data}, nil
}

func (b *menuBiz) Create(ctx context.Context, req *v1.CreateMenuRequest) (*v1.MenuInfo, error) {
	var menuM model.MenuM
	_ = copier.Copy(&menuM, req)

	err := b.ds.SysMenu().Create(ctx, &menuM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.MenuInfo
	_ = copier.Copy(&resp, menuM)

	return &resp, nil
}

func (b *menuBiz) Get(ctx context.Context, ID uint) (*v1.MenuInfo, error) {
	menu, err := b.ds.SysMenu().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	var resp v1.MenuInfo
	_ = copier.Copy(&resp, menu)

	return &resp, nil
}

func (b *menuBiz) Update(ctx context.Context, ID uint, req *v1.UpdateMenuRequest) (*v1.MenuInfo, error) {
	menuM, err := b.ds.SysMenu().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	if req.ParentID != nil {
		menuM.ParentID = *req.ParentID
	}

	if req.Title != nil {
		menuM.Title = *req.Title
	}

	if req.Name != nil {
		menuM.Name = *req.Name
	}

	if req.Path != nil {
		menuM.Path = *req.Path
	}

	if req.Hidden != nil {
		menuM.Hidden = *req.Hidden
	}

	if req.Sort != nil {
		menuM.Sort = *req.Sort
	}

	if req.Icon != nil {
		menuM.Icon = *req.Icon
	}

	if req.Component != nil {
		menuM.Component = *req.Component
	}

	if req.Redirect != nil {
		menuM.Redirect = *req.Redirect
	}

	if err := b.ds.SysMenu().Update(ctx, menuM); err != nil {
		return nil, err
	}

	var resp v1.MenuInfo
	_ = copier.Copy(&resp, menuM)

	return &resp, nil
}

func (b *menuBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.SysMenu().DeleteByID(ctx, ID)
}

func (b *menuBiz) All(ctx context.Context) ([]*v1.MenuInfo, error) {
	list, err := b.ds.SysMenu().All(ctx)
	if err != nil {
		log.C(ctx).Errorw("Failed to list menus", "err", err)

		return nil, err
	}

	data := make([]*v1.MenuInfo, 0, len(list))
	for _, item := range list {
		var menu v1.MenuInfo
		_ = copier.Copy(&menu, item)

		data = append(data, &menu)
	}

	return data, nil
}

func (b *menuBiz) Tree(ctx context.Context) (ret []*v1.MenuInfo, err error) {
	all, err := b.ds.SysMenu().All(ctx)
	if err != nil {
		return nil, err
	}

	tree, err := b.ds.SysMenu().Tree(ctx, all)
	if err != nil {
		return
	}

	data := make([]*v1.MenuInfo, 0, len(tree))
	for _, item := range tree {
		var menu v1.MenuInfo
		_ = copier.Copy(&menu, item)

		data = append(data, &menu)
	}

	return data, nil
}

func (b *menuBiz) ToggleHidden(ctx context.Context, ID uint) (*v1.MenuInfo, error) {
	menuM, err := b.ds.SysMenu().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	menuM.Hidden = !menuM.Hidden
	if err := b.ds.SysMenu().Update(ctx, menuM, "hidden"); err != nil {
		return nil, err
	}

	var resp v1.MenuInfo
	_ = copier.Copy(&resp, menuM)

	return &resp, nil
}
