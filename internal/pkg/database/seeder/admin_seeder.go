package seeder

import (
	"context"

	"bingo/internal/pkg/known"
	"bingo/internal/pkg/model"
	"bingo/internal/pkg/store"
	"bingo/pkg/auth"
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
	authz, _ := auth.NewAuthz(store.S.DB(ctx))
	_, err = authz.AddNamedPolicy("p", known.RolePrefix+known.RoleRoot, "*", auth.AclDefaultMethods)
	if err != nil {
		return err
	}

	return nil
}
