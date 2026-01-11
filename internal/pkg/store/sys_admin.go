// ABOUTME: System admin store implementation using generic store pattern.
// ABOUTME: Provides CRUD and expansion methods for admin user management.

package store

import (
	"context"

	"github.com/bingo-project/component-base/util/gormutil"

	"github.com/bingo-project/bingo/internal/pkg/model"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	genericstore "github.com/bingo-project/bingo/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
)

// AdminStore defines the interface for admin operations.
type AdminStore interface {
	Create(ctx context.Context, obj *model.AdminM) error
	Update(ctx context.Context, obj *model.AdminM, fields ...string) error
	Delete(ctx context.Context, opts *where.Options) error
	Get(ctx context.Context, opts *where.Options) (*model.AdminM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.AdminM, error)
	FirstOrCreate(ctx context.Context, where any, obj *model.AdminM) error

	AdminExpansion
}

// AdminExpansion defines additional methods for admin operations.
type AdminExpansion interface {
	ListWithRequest(ctx context.Context, req *v1.ListAdminRequest) (int64, []*model.AdminM, error)
	GetByUsername(ctx context.Context, username string) (*model.AdminM, error)
	DeleteByUsername(ctx context.Context, username string) error
	CheckExist(ctx context.Context, admin *model.AdminM) (bool, error)
	HasRole(ctx context.Context, admin *model.AdminM, roleName string) bool
	GetUserInfo(ctx context.Context, username string) (*model.AdminM, error)
	UpdateWithRoles(ctx context.Context, admin *model.AdminM) error
}

type adminStore struct {
	*genericstore.Store[model.AdminM]
}

var _ AdminStore = (*adminStore)(nil)

func NewAdminStore(store *datastore) *adminStore {
	return &adminStore{
		Store: genericstore.NewStore[model.AdminM](store, NewLogger()),
	}
}

// ListWithRequest lists admins based on request parameters.
func (s *adminStore) ListWithRequest(ctx context.Context, req *v1.ListAdminRequest) (int64, []*model.AdminM, error) {
	opts := where.NewWhere()

	if req.Status != "" {
		opts = opts.F("status", req.Status)
	}
	if req.RoleName != "" {
		opts = opts.F("role_name", req.RoleName)
	}

	db := s.DB(ctx, opts).Preload("Role").Preload("Roles")

	// Add keyword search
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ? OR phone LIKE ?",
			keyword, keyword, keyword, keyword)
	}

	var ret []*model.AdminM
	count, err := gormutil.Paginate(db, &req.ListOptions, &ret)

	return count, ret, err
}

// GetByUsername retrieves an admin by username.
func (s *adminStore) GetByUsername(ctx context.Context, username string) (*model.AdminM, error) {
	return s.Get(ctx, where.F("username", username))
}

// DeleteByUsername deletes an admin by username.
func (s *adminStore) DeleteByUsername(ctx context.Context, username string) error {
	return s.Delete(ctx, where.F("username", username))
}

// CheckExist checks if an admin exists by username, email, or phone.
func (s *adminStore) CheckExist(ctx context.Context, admin *model.AdminM) (bool, error) {
	var id uint

	if admin.Username != "" {
		s.DB(ctx).Model(admin).Where("username = ?", admin.Username).Select("id").First(&id)
		if id > 0 {
			return true, nil
		}
	}

	if admin.Email != nil {
		s.DB(ctx).Model(admin).Where("email = ?", admin.Email).Select("id").First(&id)
		if id > 0 {
			return true, nil
		}
	}

	s.DB(ctx).Model(admin).Where("phone = ?", admin.Phone).Select("id").First(&id)

	return id > 0, nil
}

// HasRole checks if an admin has a specific role.
func (s *adminStore) HasRole(ctx context.Context, admin *model.AdminM, roleName string) bool {
	count := s.DB(ctx).Model(admin).Where("role_name = ?", roleName).Association("Roles").Count()

	return count > 0
}

// GetUserInfo retrieves an admin with roles preloaded.
func (s *adminStore) GetUserInfo(ctx context.Context, username string) (*model.AdminM, error) {
	var ret model.AdminM
	err := s.DB(ctx).
		Preload("Role").
		Preload("Roles").
		Where("username = ?", username).
		First(&ret).
		Error

	return &ret, err
}

// UpdateWithRoles updates an admin and replaces its roles association.
func (s *adminStore) UpdateWithRoles(ctx context.Context, admin *model.AdminM) error {
	err := s.DB(ctx).Model(admin).Association("Roles").Replace(admin.Roles)
	if err != nil {
		return err
	}

	return s.DB(ctx).Save(admin).Error
}
