// ABOUTME: Migration to add OAuth generalization fields to uc_auth_provider table.
// ABOUTME: Adds user_info_url, field_mapping, token_in_query, and extra_headers columns.

package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

// addAuthProviderOAuthFields defines the new columns to add to uc_auth_provider.
type addAuthProviderOAuthFields struct {
	UserInfoURL  string `gorm:"column:user_info_url;type:varchar(500)"`
	FieldMapping string `gorm:"column:field_mapping;type:text"`
	TokenInQuery bool   `gorm:"column:token_in_query;default:false"`
	ExtraHeaders string `gorm:"column:extra_headers;type:text"`
}

func (addAuthProviderOAuthFields) TableName() string {
	return "uc_auth_provider"
}

func (addAuthProviderOAuthFields) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&addAuthProviderOAuthFields{})
}

func (addAuthProviderOAuthFields) Down(migrator gorm.Migrator) {
	_ = migrator.DropColumn(&addAuthProviderOAuthFields{}, "user_info_url")
	_ = migrator.DropColumn(&addAuthProviderOAuthFields{}, "field_mapping")
	_ = migrator.DropColumn(&addAuthProviderOAuthFields{}, "token_in_query")
	_ = migrator.DropColumn(&addAuthProviderOAuthFields{}, "extra_headers")
}

func init() {
	migrate.Add("2025_12_26_120000_add_auth_provider_oauth_fields", addAuthProviderOAuthFields{}.Up, addAuthProviderOAuthFields{}.Down)
}
