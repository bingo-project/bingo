package migration

import (
	"gorm.io/gorm"

	"github.com/bingo-project/bingoctl/pkg/migrate"

	"bingo/internal/apiserver/model"
)

type CreateSysCfgAppTable struct {
	model.Base

	Name        string `gorm:"type:varchar(255);not null"`
	Version     string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:varchar(1000);not null"`
	AboutUs     string `gorm:"type:varchar(2000);not null"`
	Logo        string `gorm:"type:varchar(255);not null"`
	Enabled     int32  `gorm:"type:tinyint;not null;comment:Is enabled"`
}

func (CreateSysCfgAppTable) TableName() string {
	return "sys_cfg_app"
}

func (CreateSysCfgAppTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysCfgAppTable{})
}

func (CreateSysCfgAppTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysCfgAppTable{})
}

func init() {
	migrate.Add("2024_02_10_173725_create_sys_cfg_app_table", CreateSysCfgAppTable{}.Up, CreateSysCfgAppTable{}.Down)
}
