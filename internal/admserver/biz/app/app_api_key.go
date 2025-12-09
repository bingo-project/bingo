package app

import (
	"context"
	"regexp"

	"github.com/dromara/carbon/v2"
	"github.com/duke-git/lancet/v2/pointer"
	"github.com/jinzhu/copier"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type ApiKeyBiz interface {
	List(ctx context.Context, req *v1.ListApiKeyRequest) (*v1.ListApiKeyResponse, error)
	Create(ctx context.Context, req *v1.CreateApiKeyRequest) (*v1.ApiKeyInfo, error)
	Get(ctx context.Context, ID uint) (*v1.ApiKeyInfo, error)
	Update(ctx context.Context, ID uint, req *v1.UpdateApiKeyRequest) (*v1.ApiKeyInfo, error)
	Delete(ctx context.Context, ID uint) error
}

type apiKeyBiz struct {
	ds store.IStore
}

var _ ApiKeyBiz = (*apiKeyBiz)(nil)

func NewApiKey(ds store.IStore) *apiKeyBiz {
	return &apiKeyBiz{ds: ds}
}

func (b *apiKeyBiz) List(ctx context.Context, req *v1.ListApiKeyRequest) (*v1.ListApiKeyResponse, error) {
	count, list, err := b.ds.ApiKey().ListWithRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list apiKeys", "err", err)

		return nil, err
	}

	data := make([]v1.ApiKeyInfo, 0)
	for _, item := range list {
		var apiKey v1.ApiKeyInfo
		_ = copier.Copy(&apiKey, item)

		data = append(data, apiKey)
	}

	return &v1.ListApiKeyResponse{Total: count, Data: data}, nil
}

func (b *apiKeyBiz) Create(ctx context.Context, req *v1.CreateApiKeyRequest) (*v1.ApiKeyInfo, error) {
	// Check app
	app, err := b.ds.App().GetByAppID(ctx, req.AppID)
	if err != nil {
		return nil, errno.ErrAppNotFound
	}

	apiKeyM := model.ApiKey{
		UID:         app.UID,
		AppID:       app.AppID,
		Name:        req.Name,
		Status:      model.ApiKeyStatus(req.Status),
		ACL:         req.ACL,
		Description: req.Description,
		ExpiredAt:   pointer.Of(carbon.Parse(req.ExpiredAt).StdTime()),
	}

	err = b.ds.ApiKey().Create(ctx, &apiKeyM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.ApiKeyInfo
	_ = copier.Copy(&resp, apiKeyM)

	return &resp, nil
}

func (b *apiKeyBiz) Get(ctx context.Context, ID uint) (*v1.ApiKeyInfo, error) {
	apiKey, err := b.ds.ApiKey().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	var resp v1.ApiKeyInfo
	_ = copier.Copy(&resp, apiKey)

	return &resp, nil
}

func (b *apiKeyBiz) Update(ctx context.Context, ID uint, req *v1.UpdateApiKeyRequest) (*v1.ApiKeyInfo, error) {
	apiKeyM, err := b.ds.ApiKey().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	if req.Name != nil {
		apiKeyM.Name = *req.Name
	}
	if req.Status != nil {
		apiKeyM.Status = model.ApiKeyStatus(*req.Status)
	}
	if req.ACL != nil {
		apiKeyM.ACL = req.ACL
	}
	if req.Description != nil {
		apiKeyM.Description = *req.Description
	}
	if req.ExpiredAt != nil {
		apiKeyM.ExpiredAt = pointer.Of(carbon.Parse(*req.ExpiredAt).StdTime())
	}

	if err := b.ds.ApiKey().Update(ctx, apiKeyM); err != nil {
		return nil, err
	}

	var resp v1.ApiKeyInfo
	_ = copier.Copy(&resp, apiKeyM)

	return &resp, nil
}

func (b *apiKeyBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.ApiKey().DeleteByID(ctx, ID)
}
