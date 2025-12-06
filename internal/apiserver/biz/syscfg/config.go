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

type ConfigBiz interface {
	List(ctx context.Context, req *v1.ListConfigRequest) (*v1.ListConfigResponse, error)
	Create(ctx context.Context, req *v1.CreateConfigRequest) (*v1.ConfigInfo, error)
	Get(ctx context.Context, ID uint) (*v1.ConfigInfo, error)
	Update(ctx context.Context, ID uint, req *v1.UpdateConfigRequest) (*v1.ConfigInfo, error)
	Delete(ctx context.Context, ID uint) error
}

type configBiz struct {
	ds store.IStore
}

var _ ConfigBiz = (*configBiz)(nil)

func NewConfig(ds store.IStore) *configBiz {
	return &configBiz{ds: ds}
}

func (b *configBiz) List(ctx context.Context, req *v1.ListConfigRequest) (*v1.ListConfigResponse, error) {
	count, list, err := b.ds.SysConfig().ListWithRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list configs", "err", err)

		return nil, err
	}

	data := make([]v1.ConfigInfo, 0)
	for _, item := range list {
		var config v1.ConfigInfo
		_ = copier.Copy(&config, item)

		data = append(data, config)
	}

	return &v1.ListConfigResponse{Total: count, Data: data}, nil
}

func (b *configBiz) Create(ctx context.Context, req *v1.CreateConfigRequest) (*v1.ConfigInfo, error) {
	var configM model.Config
	_ = copier.Copy(&configM, req)

	err := b.ds.SysConfig().Create(ctx, &configM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.ConfigInfo
	_ = copier.Copy(&resp, configM)

	return &resp, nil
}

func (b *configBiz) Get(ctx context.Context, ID uint) (*v1.ConfigInfo, error) {
	config, err := b.ds.SysConfig().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	var resp v1.ConfigInfo
	_ = copier.Copy(&resp, config)

	return &resp, nil
}

func (b *configBiz) Update(ctx context.Context, ID uint, req *v1.UpdateConfigRequest) (*v1.ConfigInfo, error) {
	configM, err := b.ds.SysConfig().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrResourceNotFound
	}

	if req.Name != nil {
		configM.Name = *req.Name
	}
	if req.Description != nil {
		configM.Description = *req.Description
	}
	if req.Key != nil {
		configM.Key = model.CfgKey(*req.Key)
	}
	if req.Value != nil {
		configM.Value = *req.Value
	}
	if req.OperatorID != nil {
		configM.OperatorID = *req.OperatorID
	}

	if err := b.ds.SysConfig().Update(ctx, configM); err != nil {
		return nil, err
	}

	var resp v1.ConfigInfo
	_ = copier.Copy(&resp, configM)

	return &resp, nil
}

func (b *configBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.SysConfig().DeleteByID(ctx, ID)
}
