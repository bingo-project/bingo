package seeder

import (
	"context"

	"bingo/internal/apiserver/global"
	"bingo/internal/apiserver/store"
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

	// Init admin account.
	err := store.S.Admins().InitData(ctx)
	if err != nil {
		return err
	}

	// Init permission
	authz, _ := auth.NewAuthz(store.S.DB())
	_, err = authz.AddNamedPolicy("p", global.RolePrefix+global.RoleRoot, auth.AclDefaultMethods)
	if err != nil {
		return err
	}

	return nil
}
