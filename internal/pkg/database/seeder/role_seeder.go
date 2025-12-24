// ABOUTME: Seeds built-in roles for the system.
// ABOUTME: Creates default roles like admin with proper status.

package seeder

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

var defaultRoles = []model.RoleM{
	{Name: "super-admin", Description: "Super admin", Status: "enabled"},
	{Name: "admin", Description: "System administrator", Status: "enabled"},
}

type RoleSeeder struct{}

func (RoleSeeder) Signature() string {
	return "RoleSeeder"
}

func (RoleSeeder) Run() error {
	ctx := context.Background()

	for _, role := range defaultRoles {
		where := &model.RoleM{Name: role.Name}
		if err := store.S.SysRole().FirstOrCreate(ctx, where, &role); err != nil {
			return err
		}
	}

	return nil
}
