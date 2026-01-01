// ABOUTME: AI role business logic for admin management.
// ABOUTME: Provides full CRUD operations for AI role presets with no status restrictions.
package ai

import (
	"context"
	"errors"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// AiRoleBiz defines AI role management interface for admin.
type AiRoleBiz interface {
	Create(ctx context.Context, req *v1.CreateAiRoleRequest) (*v1.AiRoleInfo, error)
	Get(ctx context.Context, roleID string) (*v1.AiRoleInfo, error)
	List(ctx context.Context, req *v1.ListAiRoleRequest) (*v1.ListAiRoleResponse, error)
	Update(ctx context.Context, roleID string, req *v1.UpdateAiRoleRequest) (*v1.AiRoleInfo, error)
	Delete(ctx context.Context, roleID string) error
}

type aiRoleBiz struct {
	ds store.IStore
}

var _ AiRoleBiz = (*aiRoleBiz)(nil)

func NewAiRole(ds store.IStore) AiRoleBiz {
	return &aiRoleBiz{ds: ds}
}

// toRoleInfo converts model.AiRoleM to v1.AiRoleInfo.
func toRoleInfo(m *model.AiRoleM) *v1.AiRoleInfo {
	return &v1.AiRoleInfo{
		RoleID:       m.RoleID,
		Name:         m.Name,
		Description:  m.Description,
		Icon:         m.Icon,
		Category:     string(m.Category),
		SystemPrompt: m.SystemPrompt,
		Model:        m.Model,
		Temperature:  m.Temperature,
		MaxTokens:    m.MaxTokens,
		Sort:         m.Sort,
		Status:       string(m.Status),
	}
}

func (b *aiRoleBiz) Create(ctx context.Context, req *v1.CreateAiRoleRequest) (*v1.AiRoleInfo, error) {
	// Check if role already exists
	existing, err := b.ds.AiRole().GetByRoleID(ctx, req.RoleID)
	if err == nil && existing != nil {
		return nil, errno.ErrResourceAlreadyExists.WithMessage("role_id already exists: %s", req.RoleID)
	}

	// Set default category if not provided
	category := model.AiRoleCategoryGeneral
	if req.Category != "" {
		category = model.AiRoleCategory(req.Category)
	}

	role := &model.AiRoleM{
		RoleID:       req.RoleID,
		Name:         req.Name,
		Description:  req.Description,
		Icon:         req.Icon,
		Category:     category,
		SystemPrompt: req.SystemPrompt,
		Model:        req.Model,
		Temperature:  req.Temperature,
		MaxTokens:    req.MaxTokens,
		Sort:         req.Sort,
		Status:       model.AiRoleStatusActive,
	}

	if err := b.ds.AiRole().Create(ctx, role); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("create ai role: %v", err)
	}

	log.C(ctx).Infow("ai role created", "role_id", role.RoleID, "name", role.Name)

	return toRoleInfo(role), nil
}

func (b *aiRoleBiz) Get(ctx context.Context, roleID string) (*v1.AiRoleInfo, error) {
	role, err := b.ds.AiRole().GetByRoleID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIRoleNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get ai role: %v", err)
	}

	return toRoleInfo(role), nil
}

func (b *aiRoleBiz) List(ctx context.Context, req *v1.ListAiRoleRequest) (*v1.ListAiRoleResponse, error) {
	var category model.AiRoleCategory
	if req.Category != "" {
		category = model.AiRoleCategory(req.Category)
	}

	// Admin can list all statuses, no default filtering
	var status model.AiRoleStatus
	if req.Status != "" {
		status = model.AiRoleStatus(req.Status)
	}

	roles, err := b.ds.AiRole().ListByCategory(ctx, category, status)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("list ai roles: %v", err)
	}

	data := make([]v1.AiRoleInfo, len(roles))
	for i, r := range roles {
		data[i] = *toRoleInfo(r)
	}

	return &v1.ListAiRoleResponse{
		Total: int64(len(roles)),
		Data:  data,
	}, nil
}

func (b *aiRoleBiz) Update(ctx context.Context, roleID string, req *v1.UpdateAiRoleRequest) (*v1.AiRoleInfo, error) {
	role, err := b.ds.AiRole().GetByRoleID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIRoleNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get ai role: %v", err)
	}

	// Update fields
	_ = copier.CopyWithOption(role, req, copier.Option{IgnoreEmpty: true})
	if req.Category != "" {
		role.Category = model.AiRoleCategory(req.Category)
	}
	if req.Status != "" {
		role.Status = model.AiRoleStatus(req.Status)
	}

	if err := b.ds.AiRole().Update(ctx, role); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("update ai role: %v", err)
	}

	log.C(ctx).Infow("ai role updated", "role_id", role.RoleID)

	return toRoleInfo(role), nil
}

func (b *aiRoleBiz) Delete(ctx context.Context, roleID string) error {
	role, err := b.ds.AiRole().GetByRoleID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.ErrAIRoleNotFound
		}

		return errno.ErrDBRead.WithMessage("get ai role: %v", err)
	}

	role.Status = model.AiRoleStatusDisabled
	if err := b.ds.AiRole().Update(ctx, role, "status"); err != nil {
		return errno.ErrDBWrite.WithMessage("delete ai role: %v", err)
	}

	log.C(ctx).Infow("ai role deleted", "role_id", role.RoleID)

	return nil
}
