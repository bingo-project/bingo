package auth

import (
	"context"
	"regexp"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"golang.org/x/oauth2"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type AuthProviderBiz interface {
	List(ctx context.Context, req *v1.ListAuthProviderRequest) (*v1.ListAuthProviderResponse, error)
	Create(ctx context.Context, req *v1.CreateAuthProviderRequest) (*v1.AuthProviderInfo, error)
	Get(ctx context.Context, ID uint) (*v1.AuthProviderInfo, error)
	Update(ctx context.Context, ID uint, req *v1.UpdateAuthProviderRequest) (*v1.AuthProviderInfo, error)
	Delete(ctx context.Context, ID uint) error
	FindEnabled(ctx context.Context) ([]*v1.AuthProviderBrief, error)
}

type authProviderBiz struct {
	ds store.IStore
}

var _ AuthProviderBiz = (*authProviderBiz)(nil)

func NewAuthProvider(ds store.IStore) *authProviderBiz {
	return &authProviderBiz{ds: ds}
}

func (b *authProviderBiz) List(ctx context.Context, req *v1.ListAuthProviderRequest) (*v1.ListAuthProviderResponse, error) {
	count, list, err := b.ds.AuthProvider().ListWithRequest(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("Failed to list authProviders", "err", err)

		return nil, err
	}

	data := make([]v1.AuthProviderInfo, 0)
	for _, item := range list {
		var authProvider v1.AuthProviderInfo
		_ = copier.Copy(&authProvider, item)

		data = append(data, authProvider)
	}

	return &v1.ListAuthProviderResponse{Total: count, Data: data}, nil
}

func (b *authProviderBiz) Create(ctx context.Context, req *v1.CreateAuthProviderRequest) (*v1.AuthProviderInfo, error) {
	var authProviderM model.AuthProvider
	_ = copier.Copy(&authProviderM, req)

	err := b.ds.AuthProvider().Create(ctx, &authProviderM)
	if err != nil {
		// Check exists
		if match, _ := regexp.MatchString("Duplicate entry '.*' for key", err.Error()); match {
			return nil, errno.ErrResourceAlreadyExists
		}

		return nil, err
	}

	var resp v1.AuthProviderInfo
	_ = copier.Copy(&resp, authProviderM)

	return &resp, nil
}

func (b *authProviderBiz) Get(ctx context.Context, ID uint) (*v1.AuthProviderInfo, error) {
	authProvider, err := b.ds.AuthProvider().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	var resp v1.AuthProviderInfo
	_ = copier.Copy(&resp, authProvider)

	return &resp, nil
}

func (b *authProviderBiz) Update(ctx context.Context, ID uint, req *v1.UpdateAuthProviderRequest) (*v1.AuthProviderInfo, error) {
	authProviderM, err := b.ds.AuthProvider().GetByID(ctx, ID)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	if req.Name != nil {
		authProviderM.Name = *req.Name
	}
	if req.Status != nil {
		authProviderM.Status = model.AuthProviderStatus(*req.Status)
	}
	if req.IsDefault != nil {
		authProviderM.IsDefault = *req.IsDefault
	}
	if req.AppID != nil {
		authProviderM.AppID = *req.AppID
	}
	if req.ClientID != nil {
		authProviderM.ClientID = *req.ClientID
	}
	if req.ClientSecret != nil {
		authProviderM.ClientSecret = *req.ClientSecret
	}
	if req.RedirectURL != nil {
		authProviderM.RedirectURL = *req.RedirectURL
	}
	if req.AuthURL != nil {
		authProviderM.AuthURL = *req.AuthURL
	}
	if req.TokenURL != nil {
		authProviderM.TokenURL = *req.TokenURL
	}
	if req.LogoutURI != nil {
		authProviderM.LogoutURI = *req.LogoutURI
	}
	if req.Info != nil {
		authProviderM.Info = *req.Info
	}

	if err := b.ds.AuthProvider().Update(ctx, authProviderM); err != nil {
		return nil, err
	}

	var resp v1.AuthProviderInfo
	_ = copier.Copy(&resp, authProviderM)

	return &resp, nil
}

func (b *authProviderBiz) Delete(ctx context.Context, ID uint) error {
	return b.ds.AuthProvider().DeleteByID(ctx, ID)
}

func (b *authProviderBiz) FindEnabled(ctx context.Context) (ret []*v1.AuthProviderBrief, err error) {
	list, err := b.ds.AuthProvider().FindEnabled(ctx)
	if err != nil {
		return nil, err
	}

	data := make([]*v1.AuthProviderBrief, 0)
	for _, item := range list {
		var authProvider v1.AuthProviderBrief
		_ = copier.Copy(&authProvider, item)

		// Get oauth config
		conf := oauth2.Config{
			ClientID:     item.ClientID,
			ClientSecret: item.ClientSecret,
			RedirectURL:  item.RedirectURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  item.AuthURL,
				TokenURL: item.TokenURL,
			},
		}

		// Get Auth URL
		authProvider.AuthURL = conf.AuthCodeURL(uuid.New().String())

		data = append(data, &authProvider)
	}

	return data, err
}
