// ABOUTME: Migration to change status field type from tinyint to varchar.
// ABOUTME: Values change from 1/2 to enabled/disabled for consistency.

package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

// changeAuthProviderStatusTypeOld represents the old schema with status as int.
type changeAuthProviderStatusTypeOld struct {
	Status int `gorm:"column:status;type:tinyint;not null;default:2"`
}

func (changeAuthProviderStatusTypeOld) TableName() string {
	return "uc_auth_provider"
}

// changeAuthProviderStatusTypeNew represents the new schema with status as string.
type changeAuthProviderStatusTypeNew struct {
	Status string `gorm:"column:status;type:varchar(20);not null;default:'disabled'"`
}

func (changeAuthProviderStatusTypeNew) TableName() string {
	return "uc_auth_provider"
}

type changeAuthProviderStatusType struct{}

func (changeAuthProviderStatusType) TableName() string {
	return "uc_auth_provider"
}

func (changeAuthProviderStatusType) Up(migrator gorm.Migrator) {
	// Use AlterColumn to change the status field type
	_ = migrator.AlterColumn(&changeAuthProviderStatusTypeNew{}, "status")
}

func (changeAuthProviderStatusType) Down(migrator gorm.Migrator) {
	// Revert back to tinyint
	_ = migrator.AlterColumn(&changeAuthProviderStatusTypeOld{}, "status")
}

func init() {
	migrate.Add("2025_12_27_020000_change_auth_provider_status_type", changeAuthProviderStatusType{}.Up, changeAuthProviderStatusType{}.Down)
}
