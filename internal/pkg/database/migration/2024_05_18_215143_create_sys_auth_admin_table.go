package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/model"
)

type CreateSysAuthAdminTable struct {
	model.Base

	Username string  `gorm:"uniqueIndex:uk_username;type:varchar(255);not null"`
	Password string  `gorm:"type:varchar(255);not null;default:''"`
	Nickname string  `gorm:"type:varchar(255);not null;default:''"`
	Email    *string `gorm:"uniqueIndex:uk_email;type:varchar(255);default:null"`
	Phone    *string `gorm:"uniqueIndex:uk_phone;type:varchar(255);default:null"`
	Avatar   string  `gorm:"type:varchar(255);not null;default:''"`
	Status   uint    `gorm:"type:tinyint;default:1;comment:状态：1正常，2冻结"`
	RoleName string  `gorm:"index:idx_role;type:varchar(255);not null;default:'';comment:当前角色"`
}

func (CreateSysAuthAdminTable) TableName() string {
	return "sys_auth_admin"
}

func (CreateSysAuthAdminTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateSysAuthAdminTable{})
}

func (CreateSysAuthAdminTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateSysAuthAdminTable{})
}

func init() {
	migrate.Add("2024_05_18_215143_create_sys_auth_admin_table", CreateSysAuthAdminTable{}.Up, CreateSysAuthAdminTable{}.Down)
}
