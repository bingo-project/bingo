package migration

import (
	"gorm.io/gorm"

	"github.com/bingo-project/bingoctl/pkg/migrate"

	"bingo/internal/pkg/model"
)

type CreateAppTable struct {
	model.Base

	UID         string `gorm:"type:varchar(255);index:idx_uid;not null;default:''"`
	AppID       string `gorm:"type:varchar(255);uniqueIndex:uk_app_id;not null;default:''"`
	Name        string `gorm:"type:varchar(255);not null;default:''"`
	Status      string `gorm:"type:tinyint;not null;default:1;comment:Status, 1-enabled, 2-disabled"`
	Description string `gorm:"type:varchar(1000);not null;default:''"`
	Logo        string `gorm:"type:varchar(1000);not null;default:''"`
}

func (CreateAppTable) TableName() string {
	return "app"
}

func (CreateAppTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateAppTable{})
}

func (CreateAppTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateAppTable{})
}

func init() {
	migrate.Add("2024_05_21_205542_create_app_table", CreateAppTable{}.Up, CreateAppTable{}.Down)
}
