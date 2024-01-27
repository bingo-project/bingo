package migration

import (
	"gorm.io/gorm"

	"github.com/bingo-project/bingoctl/pkg/migrate"

	"bingo/internal/pkg/model"
)

type CreateUserTable struct {
	model.Base

	Username string `gorm:"uniqueIndex:uk_username;type:varchar(255);;not null"`
	Password string `gorm:"type:varchar(255);not null;default:''"`
	Nickname string `gorm:"type:varchar(255);default:''"`
	Email    string `gorm:"uniqueIndex:uk_email;type:varchar(255)"`
	Phone    string `gorm:"uniqueIndex:uk_phone;type:varchar(255)"`
}

func (CreateUserTable) TableName() string {
	return "user"
}

func (CreateUserTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateUserTable{})
}

func (CreateUserTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateUserTable{})
}

func init() {
	migrate.Add("2024_01_27_194339_create_user_table", CreateUserTable{}.Up, CreateUserTable{}.Down)
}
