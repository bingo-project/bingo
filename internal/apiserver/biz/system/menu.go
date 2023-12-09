package system

import (
	"context"
	"regexp"

	"github.com/bingo-project/component-base/log"
	"github.com/jinzhu/copier"

	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/model"
	v1 "bingo/pkg/api/bingo/v1"
)

type MenuBiz interface {
	List(ctx context.Context, req *v1.ListMenuRequest) (*v1.ListResponse, error)
	Create(ctx context.Context, req *v1.CreateMenuRequest) (*v1.MenuInfo, error)
	Get(ctx context.Context, ID uint) (*v1.MenuInfo, error)
	Update(ctx context.Context, ID uint, req *v1.UpdateMenuRequest) (*v1.MenuInfo, error)
	Delete(ctx context.Context, ID uint) error

	All(ctx context.Context) ([]*v1.MenuInfo, error)
}

type menuBiz struct {
	ds store.IStore
}

var _ MenuBiz = (*menuBiz)(nil)

func NewMenu(ds store.IStore) *menuBiz {
	return &menuBiz{ds: ds}
}

func (b *menuBiz) List(ctx context.Context, req *v1.ListMenuRequest) (*v1.ListResponse, error) {
	count, list, err := b.ds.Menus().List(ctx, req)
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

	return &v1.ListResponse{Total: count, Data: data}, nil
}

func (b *menuBiz) Create(ctx context.Context, req *v1.CreateMenuRequest) (*v1.MenuInfo, error) {
	var menuM model.MenuM
	_ = copier.Copy(&menuM, req)

	err := b.ds.Menus().Create(ctx, &menuM)
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
	menu, err := b.ds.Menus().Get(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	var resp v1.MenuInfo
	_ = copier.Copy(&resp, menu)

	return &resp, nil
}

func (b *menuBiz) Update(ctx context.Context, ID uint, req *v1.UpdateMenuRequest) (*v1.MenuInfo, error) {
	menuM, err := b.ds.Menus().Get(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
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

	if err := b.ds.Menus().Update(ctx, menuM); err != nil {
		return nil, err
	}

	var resp v1.MenuInfo
	_ = copier.Copy(&resp, req)

	return &resp, nil
}

func (b *menuBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.Menus().Delete(ctx, ID)
}

func (b *menuBiz) All(ctx context.Context) ([]*v1.MenuInfo, error) {
	list, err := b.ds.Apis().All(ctx)
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
