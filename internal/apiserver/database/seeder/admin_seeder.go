package seeder

import (
	"context"

	"bingo/internal/apiserver/store"
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

	return nil
}
