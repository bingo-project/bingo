package migration

import (
	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"

	"bingo/internal/pkg/model"
)

type CreateUserAccountTable struct {
	model.Base

	UID       string `gorm:"type:varchar(255);index:idx_uid"`
	Provider  string `gorm:"type:varchar(255);not null;default:''"`
	AccountID string `gorm:"type:varchar(255);not null;default:''"`
	Username  string `gorm:"type:varchar(255);not null;default:''"`
	Nickname  string `gorm:"type:varchar(255);not null;default:''"`
	Email     string `gorm:"type:varchar(255);not null;default:''"`
	Bio       string `gorm:"type:varchar(255);not null;default:''"`
	Avatar    string `gorm:"type:varchar(255);not null;default:''"`
	Nonce     string `gorm:"type:varchar(255);not null;default:''"`
}

func (CreateUserAccountTable) TableName() string {
	return "uc_user_account"
}

func (CreateUserAccountTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateUserAccountTable{})
}

func (CreateUserAccountTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateUserAccountTable{})
}

func init() {
	migrate.Add("2024_05_15_233734_create_user_account_table", CreateUserAccountTable{}.Up, CreateUserAccountTable{}.Down)
}
