package seeder

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

type AdminSeeder struct {
}

// Signature The name and signature of the seeder.
func (AdminSeeder) Signature() string {
	return "AdminSeeder"
}

// Run seed the application's database.
func (AdminSeeder) Run() error {
	ctx := context.Background()

	admin := model.AdminM{
		Username: "root",
		Password: "123456",
		Nickname: "Root",
		Email:    nil,
		Phone:    nil,
		RoleName: "root",
	}

	// Init admin account
	where := &model.AdminM{Username: admin.Username}
	if err := store.S.Admin().FirstOrCreate(ctx, where, &admin); err != nil {
		return err
	}

	// Associate super-admin role for role switching
	roles, _ := store.S.SysRole().GetByNames(ctx, []string{"super-admin"})
	if len(roles) > 0 {
		adminM, _ := store.S.Admin().GetByUsername(ctx, admin.Username)
		if adminM != nil {
			adminM.Roles = roles
			_ = store.S.Admin().UpdateWithRoles(ctx, adminM)
		}
	}

	return nil
}
