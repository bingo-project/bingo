package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"

	"bingo/internal/pkg/model"
)

type CreateAppApiKeyTable struct {
	model.Base

	UID         string    `gorm:"type:varchar(255);index:idx_uid;not null;default:''"`
	AppID       string    `gorm:"type:varchar(255);index:idx_app_id;not null;default:''"`
	Name        string    `gorm:"type:varchar(255);not null;default:''"`
	AccessKey   string    `gorm:"type:varchar(255);not null;default:''"`
	SecretKey   string    `gorm:"type:varchar(255);not null;default:''"`
	Status      string    `gorm:"type:tinyint;not null;default:1;comment:Status, 1-enabled, 2-disabled"`
	ACL         string    `gorm:"type:json;default:null"`
	Description string    `gorm:"type:varchar(1000);not null;default:''"`
	ExpiredAt   time.Time `gorm:"type:DATETIME(3);index:idx_expired_at"`
}

func (CreateAppApiKeyTable) TableName() string {
	return "app_api_key"
}

func (CreateAppApiKeyTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAppApiKeyTable{})
}

func (CreateAppApiKeyTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAppApiKeyTable{})
}

func init() {
	migrate.Add("2024_05_21_210642_create_app_api_key_table", CreateAppApiKeyTable{}.Up, CreateAppApiKeyTable{}.Down)
}
