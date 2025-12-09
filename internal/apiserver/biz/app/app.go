package app

import (
	"context"
	"regexp"

	"github.com/jinzhu/copier"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
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
	count, list, err := b.ds.App().ListWithRequest(ctx, req)
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
	_, err := b.ds.User().GetByUID(ctx, req.UID)
	if err != nil {
		return nil, errno.ErrUserNotFound
	}

	err = b.ds.App().Create(ctx, &appM)
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
	app, err := b.ds.App().GetByAppID(ctx, appID)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	var resp v1.AppInfo
	_ = copier.Copy(&resp, app)

	return &resp, nil
}

func (b *appBiz) Update(ctx context.Context, appID string, req *v1.UpdateAppRequest) (*v1.AppInfo, error) {
	appM, err := b.ds.App().GetByAppID(ctx, appID)
	if err != nil {
		return nil, errno.ErrNotFound
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

	if err := b.ds.App().Update(ctx, appM); err != nil {
		return nil, err
	}

	var resp v1.AppInfo
	_ = copier.Copy(&resp, appM)

	return &resp, nil
}

func (b *appBiz) Delete(ctx context.Context, appID string) error {
	return b.ds.App().DeleteByAppID(ctx, appID)
}
