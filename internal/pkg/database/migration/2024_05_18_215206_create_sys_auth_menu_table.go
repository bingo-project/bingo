package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/model"
)

type CreateSysAuthMenuTable struct {
	model.Base

	ParentID  uint   `gorm:"index:idx_parent;type:int;not null;default:0"`
	Name      string `gorm:"type:varchar(255);not null;default:''"`
	Path      string `gorm:"index:idx_path;type:varchar(255);not null;default:''"`
	Sort      int    `gorm:"type:int;not null;default:0"`
	Title     string `gorm:"type:varchar(255);not null;default:''"`
	Icon      string `gorm:"type:varchar(255);not null;default:''"`
	Hidden    bool   `gorm:"type:tinyint;not null;default:0;comment:Is Hidden"`
	Component string `gorm:"type:varchar(255);not null;default:''"`
	Redirect  string `gorm:"type:varchar(255);not null;default:''"`
}

func (CreateSysAuthMenuTable) TableName() string {
	return "sys_auth_menu"
}

func (CreateSysAuthMenuTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysAuthMenuTable{})
}

func (CreateSysAuthMenuTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysAuthMenuTable{})
}

func init() {
	migrate.Add("2024_05_18_215206_create_sys_auth_menu_table", CreateSysAuthMenuTable{}.Up, CreateSysAuthMenuTable{}.Down)
}
