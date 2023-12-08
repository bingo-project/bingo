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

type ApiBiz interface {
	List(ctx context.Context, req *v1.ListApiRequest) (*v1.ListResponse, error)
	Create(ctx context.Context, req *v1.CreateApiRequest) (*v1.ApiInfo, error)
	Get(ctx context.Context, ID uint) (*v1.ApiInfo, error)
	Update(ctx context.Context, ID uint, req *v1.UpdateApiRequest) (*v1.ApiInfo, error)
	Delete(ctx context.Context, ID uint) error

	All(ctx context.Context) ([]*v1.ApiInfo, error)
}

type apiBiz struct {
	ds store.IStore
}

var _ ApiBiz = (*apiBiz)(nil)

func NewApi(ds store.IStore) *apiBiz {
	return &apiBiz{ds: ds}
}

func (b *apiBiz) List(ctx context.Context, req *v1.ListApiRequest) (*v1.ListResponse, error) {
	count, list, err := b.ds.Api().List(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list apis", "err", err)

		return nil, err
	}

	data := make([]*v1.ApiInfo, 0, len(list))
	for _, item := range list {
		var api v1.ApiInfo
		_ = copier.Copy(&api, item)

		data = append(data, &api)
	}

	return &v1.ListResponse{Total: count, Data: data}, nil
}

func (b *apiBiz) Create(ctx context.Context, req *v1.CreateApiRequest) (*v1.ApiInfo, error) {
	var apiM model.ApiM
	_ = copier.Copy(&apiM, req)

	err := b.ds.Api().Create(ctx, &apiM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.ApiInfo
	_ = copier.Copy(&resp, apiM)

	return &resp, nil
}

func (b *apiBiz) Get(ctx context.Context, ID uint) (*v1.ApiInfo, error) {
	api, err := b.ds.Api().Get(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	var resp v1.ApiInfo
	_ = copier.Copy(&resp, api)

	return &resp, nil
}

func (b *apiBiz) Update(ctx context.Context, ID uint, req *v1.UpdateApiRequest) (*v1.ApiInfo, error) {
	apiM, err := b.ds.Api().Get(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	if req.Method != nil {
		apiM.Method = *req.Method
	}

	if req.Path != nil {
		apiM.Path = *req.Path
	}

	if req.Group != nil {
		apiM.Group = *req.Group
	}

	if req.Description != nil {
		apiM.Description = *req.Description
	}

	if err := b.ds.Api().Update(ctx, apiM); err != nil {
		return nil, err
	}

	var resp v1.ApiInfo
	_ = copier.Copy(&resp, req)

	return &resp, nil
}

func (b *apiBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.Api().Delete(ctx, ID)
}

func (b *apiBiz) All(ctx context.Context) ([]*v1.ApiInfo, error) {
	list, err := b.ds.Api().All(ctx)
	if err != nil {
		log.C(ctx).Errorw("Failed to list apis from storage", "err", err)

		return nil, err
	}

	data := make([]*v1.ApiInfo, 0, len(list))
	for _, item := range list {
		var api v1.ApiInfo
		_ = copier.Copy(&api, item)

		data = append(data, &api)
	}

	return data, nil
}
