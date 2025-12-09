package seeder

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/known"
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

	// Init admin account.
	err := store.S.Admin().Create(ctx, &admin)
	if err != nil {
		return err
	}

	// Init permission
	authz, _ := auth.NewAuthorizer(store.S.DB(ctx), nil)
	_, err = authz.Enforcer().AddNamedPolicy("p", known.RolePrefix+known.RoleRoot, "*", auth.AclDefaultMethods)
	if err != nil {
		return err
	}

	return nil
}
