package migration

import (
	"gorm.io/gorm"

	"github.com/bingo-project/bingoctl/pkg/migrate"

	"bingo/internal/apiserver/model"
)

type CreateSysCfgConfigTable struct {
	model.Base

	Name        string `gorm:"type:varchar(255);not null;default:''"`
	Key         string `gorm:"type:varchar(255);not null;default:''"`
	Value       string `gorm:"type:json;not null'"`
	OperatorID  string `gorm:"type:int;not null;default:'0'"`
	Description string `gorm:"type:varchar(1024);not null;default:''"`
}

func (CreateSysCfgConfigTable) TableName() string {
	return "sys_cfg_config"
}

func (CreateSysCfgConfigTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysCfgConfigTable{})
}

func (CreateSysCfgConfigTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysCfgConfigTable{})
}

func init() {
	migrate.Add("2024_02_10_200947_create_sys_cfg_config_table", CreateSysCfgConfigTable{}.Up, CreateSysCfgConfigTable{}.Down)
}
