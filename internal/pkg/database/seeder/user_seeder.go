// ABOUTME: Test user seeder for development environment.
// ABOUTME: Creates test users with predefined credentials for quick testing.
package seeder

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

type UserSeeder struct {
}

// Signature The name and signature of the seeder.
func (UserSeeder) Signature() string {
	return "UserSeeder"
}

// Run seed the application's database.
func (UserSeeder) Run() error {
	ctx := context.Background()

	users := []model.UserM{
		{
			Username:    "test",
			Email:       "test@example.com",
			Phone:       "13800138000",
			CountryCode: "86",
			Nickname:    "Test User",
			Password:    "123456",
			Status:      model.UserStatusEnabled,
		},
		{
			Username:    "test2",
			Email:       "test2@example.com",
			Phone:       "13800138001",
			CountryCode: "86",
			Nickname:    "Test User 2",
			Password:    "123456",
			Status:      model.UserStatusEnabled,
		},
	}

	for _, user := range users {
		where := &model.UserM{Username: user.Username}
		if err := store.S.User().FirstOrCreate(ctx, where, &user); err != nil {
			return err
		}
	}

	return nil
}
