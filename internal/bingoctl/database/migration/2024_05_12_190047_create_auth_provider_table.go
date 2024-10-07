package migration

import (
	"gorm.io/gorm"

	"github.com/bingo-project/bingoctl/pkg/migrate"

	"bingo/internal/pkg/model"
)

type CreateAuthProviderTable struct {
	model.Base

	Name         string `gorm:"type:varchar(255);not null;default:'';comment:Auth provider name"`
	Status       int64  `gorm:"type:tinyint;not null;default:1;comment:Status, 1-enabled, 2-disabled"`
	IsDefault    int64  `gorm:"type:tinyint;not null;default:0;comment:Is default provider, 0-not, 1-yes"`
	AppID        string `gorm:"type:varchar(255);not null;default:'';comment:App ID"`
	ClientID     string `gorm:"type:varchar(255);not null;default:'';comment:Client ID"`
	ClientSecret string `gorm:"type:varchar(1024);not null;default:'';comment:Client secret"`
	TokenType    string `gorm:"type:varchar(1024);not null;default:'';comment:Token type"`
	RedirectURL  string `gorm:"type:varchar(1024);not null;default:'';comment:Redirect URL"`
	AuthURL      string `gorm:"type:varchar(1024);not null;default:'';comment:Auth URL"`
	TokenURL     string `gorm:"type:varchar(1024);not null;default:'';comment:Token URL"`
	LogoutURI    string `gorm:"type:varchar(1024);not null;default:'';comment:Logout URI"`
	Info         string `gorm:"type:json;comment:Ext info"`
}

func (CreateAuthProviderTable) TableName() string {
	return "uc_auth_provider"
}

func (CreateAuthProviderTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAuthProviderTable{})
}

func (CreateAuthProviderTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAuthProviderTable{})
}

func init() {
	migrate.Add("2024_05_12_190047_create_auth_provider_table", CreateAuthProviderTable{}.Up, CreateAuthProviderTable{}.Down)
}
