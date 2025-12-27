// ABOUTME: Seeder for OAuth provider templates.
// ABOUTME: Pre-populates Google, Apple, GitHub, Discord, Twitter configurations.

package seeder

import (
	"context"
	"encoding/json"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

type AuthProviderSeeder struct{}

func (AuthProviderSeeder) Signature() string {
	return "AuthProviderSeeder"
}

func (AuthProviderSeeder) Run() error {
	ctx := context.Background()
	providers := getOAuthProviderTemplates()

	for _, p := range providers {
		// Check if exists by name
		existing, _ := store.S.AuthProvider().FindByName(ctx, p.Name)
		if existing != nil {
			continue
		}

		// Create new provider
		if err := store.S.AuthProvider().Create(ctx, p); err != nil {
			return err
		}
	}

	return nil
}

func getOAuthProviderTemplates() []*model.AuthProvider {
	return []*model.AuthProvider{
		{
			Name:         model.AuthProviderGoogle,
			Status:       model.AuthProviderStatusDisabled,
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserInfoURL:  "https://www.googleapis.com/oauth2/v3/userinfo",
			Scopes:       "openid email profile",
			PKCEEnabled:  true,
			FieldMapping: mustJSON(map[string]string{
				"account_id": "sub",
				"email":      "email",
				"nickname":   "name",
				"avatar":     "picture",
			}),
		},
		{
			Name:         model.AuthProviderApple,
			Status:       model.AuthProviderStatusDisabled,
			AuthURL:      "https://appleid.apple.com/auth/authorize",
			TokenURL:     "https://appleid.apple.com/auth/token",
			Scopes:       "name email",
			PKCEEnabled:  true,
			FieldMapping: mustJSON(map[string]string{
				"account_id": "sub",
				"email":      "email",
			}),
			Info: mustJSON(map[string]string{
				"team_id":     "",
				"key_id":      "",
				"private_key": "",
			}),
		},
		{
			Name:         model.AuthProviderGithub,
			Status:       model.AuthProviderStatusDisabled,
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserInfoURL:  "https://api.github.com/user",
			Scopes:       "read:user user:email",
			PKCEEnabled:  false,
			FieldMapping: mustJSON(map[string]string{
				"account_id": "id",
				"username":   "login",
				"nickname":   "name",
				"email":      "email",
				"avatar":     "avatar_url",
				"bio":        "bio",
			}),
		},
		{
			Name:         model.AuthProviderDiscord,
			Status:       model.AuthProviderStatusDisabled,
			AuthURL:      "https://discord.com/api/oauth2/authorize",
			TokenURL:     "https://discord.com/api/oauth2/token",
			UserInfoURL:  "https://discord.com/api/users/@me",
			Scopes:       "identify email",
			PKCEEnabled:  true,
			FieldMapping: mustJSON(map[string]string{
				"account_id": "id",
				"username":   "username",
				"nickname":   "global_name",
				"email":      "email",
				"avatar":     "avatar",
			}),
		},
		{
			Name:         model.AuthProviderTwitter,
			Status:       model.AuthProviderStatusDisabled,
			AuthURL:      "https://twitter.com/i/oauth2/authorize",
			TokenURL:     "https://api.twitter.com/2/oauth2/token",
			UserInfoURL:  "https://api.twitter.com/2/users/me",
			Scopes:       "users.read tweet.read",
			PKCEEnabled:  true,
			FieldMapping: mustJSON(map[string]string{
				"account_id": "data.id",
				"username":   "data.username",
				"nickname":   "data.name",
			}),
			ExtraHeaders: mustJSON(map[string]string{
				"User-Agent": "BingoApp/1.0",
			}),
		},
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
