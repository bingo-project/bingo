package app

import (
	"context"
	"regexp"

	"github.com/bingo-project/component-base/log"
	"github.com/jinzhu/copier"

	v1 "bingo/internal/apiserver/http/request/v1"
	"bingo/internal/apiserver/store"
	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/model"
)

type AppBiz interface {
	List(ctx context.Context, req *v1.ListAppRequest) (*v1.ListAppResponse, error)
	Create(ctx context.Context, req *v1.CreateAppRequest) (*v1.AppInfo, error)
	Get(ctx context.Context, appID string) (*v1.AppInfo, error)
	Update(ctx context.Context, appID string, req *v1.UpdateAppRequest) (*v1.AppInfo, error)
	Delete(ctx context.Context, appID string) error
}

type appBiz struct {
	ds store.IStore
}

var _ AppBiz = (*appBiz)(nil)

func NewApp(ds store.IStore) *appBiz {
	return &appBiz{ds: ds}
}

func (b *appBiz) List(ctx context.Context, req *v1.ListAppRequest) (*v1.ListAppResponse, error) {
	count, list, err := b.ds.Apps().List(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list apps", "err", err)

		return nil, err
	}

	data := make([]v1.AppInfo, 0)
	for _, item := range list {
		var app v1.AppInfo
		_ = copier.Copy(&app, item)

		data = append(data, app)
	}

	return &v1.ListAppResponse{Total: count, Data: data}, nil
}

func (b *appBiz) Create(ctx context.Context, req *v1.CreateAppRequest) (*v1.AppInfo, error) {
	var appM model.App
	_ = copier.Copy(&appM, req)

	// Check owner
	_, err := b.ds.Users().GetByUID(ctx, req.UID)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	err = b.ds.Apps().Create(ctx, &appM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.AppInfo
	_ = copier.Copy(&resp, appM)

	return &resp, nil
}

func (b *appBiz) Get(ctx context.Context, appID string) (*v1.AppInfo, error) {
	app, err := b.ds.Apps().Get(ctx, appID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	var resp v1.AppInfo
	_ = copier.Copy(&resp, app)

	return &resp, nil
}

func (b *appBiz) Update(ctx context.Context, appID string, req *v1.UpdateAppRequest) (*v1.AppInfo, error) {
	appM, err := b.ds.Apps().Get(ctx, appID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	if req.Name != nil {
		appM.Name = *req.Name
	}
	if req.Description != nil {
		appM.Description = *req.Description
	}
	if req.Logo != nil {
		appM.Logo = *req.Logo
	}

	if err := b.ds.Apps().Update(ctx, appM); err != nil {
		return nil, err
	}

	var resp v1.AppInfo
	_ = copier.Copy(&resp, appM)

	return &resp, nil
}

func (b *appBiz) Delete(ctx context.Context, appID string) error {
	return b.ds.Apps().Delete(ctx, appID)
}
