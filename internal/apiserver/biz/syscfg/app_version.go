package syscfg

import (
	"context"
	"regexp"

	"github.com/jinzhu/copier"

	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	model "bingo/internal/pkg/model/syscfg"
	"bingo/internal/pkg/store"
	v1 "bingo/pkg/api/apiserver/v1/syscfg"
)

type AppVersionBiz interface {
	List(ctx context.Context, req *v1.ListAppVersionRequest) (*v1.ListAppVersionResponse, error)
	Create(ctx context.Context, req *v1.CreateAppVersionRequest) (*v1.AppVersionInfo, error)
	Get(ctx context.Context, ID uint) (*v1.AppVersionInfo, error)
	Update(ctx context.Context, ID uint, req *v1.UpdateAppVersionRequest) (*v1.AppVersionInfo, error)
	Delete(ctx context.Context, ID uint) error
}

type appVersionBiz struct {
	ds store.IStore
}

var _ AppVersionBiz = (*appVersionBiz)(nil)

func NewAppVersion(ds store.IStore) *appVersionBiz {
	return &appVersionBiz{ds: ds}
}

func (b *appVersionBiz) List(ctx context.Context, req *v1.ListAppVersionRequest) (*v1.ListAppVersionResponse, error) {
	count, list, err := b.ds.AppVersion().ListWithRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list apps", "err", err)

		return nil, err
	}

	data := make([]v1.AppVersionInfo, 0)
	for _, item := range list {
		var app v1.AppVersionInfo
		_ = copier.Copy(&app, item)

		data = append(data, app)
	}

	return &v1.ListAppVersionResponse{Total: count, Data: data}, nil
}

func (b *appVersionBiz) Create(ctx context.Context, req *v1.CreateAppVersionRequest) (*v1.AppVersionInfo, error) {
	var appM model.AppVersion
	_ = copier.Copy(&appM, req)

	err := b.ds.AppVersion().Create(ctx, &appM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.AppVersionInfo
	_ = copier.Copy(&resp, appM)

	return &resp, nil
}

func (b *appVersionBiz) Get(ctx context.Context, ID uint) (*v1.AppVersionInfo, error) {
	app, err := b.ds.AppVersion().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	var resp v1.AppVersionInfo
	_ = copier.Copy(&resp, app)

	return &resp, nil
}

func (b *appVersionBiz) Update(ctx context.Context, ID uint, req *v1.UpdateAppVersionRequest) (*v1.AppVersionInfo, error) {
	appM, err := b.ds.AppVersion().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	if req.Name != nil {
		appM.Name = *req.Name
	}
	if req.Version != nil {
		appM.Version = *req.Version
	}
	if req.Description != nil {
		appM.Description = *req.Description
	}
	if req.AboutUs != nil {
		appM.AboutUs = *req.AboutUs
	}
	if req.Logo != nil {
		appM.Logo = *req.Logo
	}
	if req.Enabled != nil {
		appM.Enabled = *req.Enabled
	}

	if err := b.ds.AppVersion().Update(ctx, appM); err != nil {
		return nil, err
	}

	var resp v1.AppVersionInfo
	_ = copier.Copy(&resp, appM)

	return &resp, nil
}

func (b *appVersionBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.AppVersion().DeleteByID(ctx, ID)
}
