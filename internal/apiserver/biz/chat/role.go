// ABOUTME: AI role business logic implementation.
// ABOUTME: Handles role query operations and validations.

package chat

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

// AiRoleBiz defines AI role query interface for users.
type AiRoleBiz interface {
	Get(ctx context.Context, roleID string) (*v1.AiRoleInfo, error)
	List(ctx context.Context, req *v1.ListAiRoleRequest) (*v1.ListAiRoleResponse, error)
}

type aiRoleBiz struct {
	ds store.IStore
}

var _ AiRoleBiz = (*aiRoleBiz)(nil)

func NewAiRole(ds store.IStore) *aiRoleBiz {
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

func (b *aiRoleBiz) Get(ctx context.Context, roleID string) (*v1.AiRoleInfo, error) {
	role, err := b.ds.AiRole().GetByRoleID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.ErrAIRoleNotFound
		}

		return nil, errno.ErrDBRead.WithMessage("get ai role: %v", err)
	}

	// Only return active roles
	if role.Status != model.AiRoleStatusActive {
		return nil, errno.ErrAIRoleNotFound
	}

	return toRoleInfo(role), nil
}

func (b *aiRoleBiz) List(ctx context.Context, req *v1.ListAiRoleRequest) (*v1.ListAiRoleResponse, error) {
	var category model.AiRoleCategory
	if req.Category != "" {
		category = model.AiRoleCategory(req.Category)
	}

	// Force filter to active only, ignore status parameter
	status := model.AiRoleStatusActive

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
