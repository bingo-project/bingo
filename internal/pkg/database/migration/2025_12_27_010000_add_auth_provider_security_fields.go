// ABOUTME: Migration to add scopes and pkce_enabled fields to uc_auth_provider table.
// ABOUTME: Supports OAuth security enhancements (PKCE) and configurable scopes.

package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type addAuthProviderSecurityFields struct {
	Scopes      string `gorm:"column:scopes;type:varchar(500)"`
	PKCEEnabled bool   `gorm:"column:pkce_enabled;default:false"`
}

func (addAuthProviderSecurityFields) TableName() string {
	return "uc_auth_provider"
}

func (addAuthProviderSecurityFields) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&addAuthProviderSecurityFields{})
}

func (addAuthProviderSecurityFields) Down(migrator gorm.Migrator) {
	_ = migrator.DropColumn(&addAuthProviderSecurityFields{}, "scopes")
	_ = migrator.DropColumn(&addAuthProviderSecurityFields{}, "pkce_enabled")
}

func init() {
	migrate.Add("2025_12_27_010000_add_auth_provider_security_fields", addAuthProviderSecurityFields{}.Up, addAuthProviderSecurityFields{}.Down)
}
