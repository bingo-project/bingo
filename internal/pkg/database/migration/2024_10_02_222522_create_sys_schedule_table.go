package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"

	"bingo/internal/pkg/model"
)

type CreateSysScheduleTable struct {
	model.Base

	Name        string `gorm:"type:varchar(255);not null;default:''"`
	Job         string `gorm:"type:varchar(255);uniqueIndex:uk_job;not null;default:''"`
	Spec        string `gorm:"type:varchar(255);not null;default:''"`
	Status      string `gorm:"type:tinyint;not null;default:1;comment:Status, 1-enabled, 2-disabled"`
	Description string `gorm:"type:varchar(1000);not null;default:''"`
}

func (CreateSysScheduleTable) TableName() string {
	return "sys_schedule"
}

func (CreateSysScheduleTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysScheduleTable{})
}

func (CreateSysScheduleTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysScheduleTable{})
}

func init() {
	migrate.Add("2024_10_02_222522_create_sys_schedule_table", CreateSysScheduleTable{}.Up, CreateSysScheduleTable{}.Down)
}
