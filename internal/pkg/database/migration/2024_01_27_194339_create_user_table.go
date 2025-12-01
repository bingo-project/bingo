package migration

import (
	"time"

	"gorm.io/gorm"

	"github.com/bingo-project/bingoctl/pkg/migrate"

	"bingo/internal/pkg/model"
)

type CreateUserTable struct {
	model.Base

	UID           string     `gorm:"type:varchar(255);uniqueIndex:uk_uid"`
	CountryCode   string     `gorm:"type:varchar(255);not null;default:''"`
	Nickname      string     `gorm:"type:varchar(255);not null;default:''"`
	Username      string     `gorm:"type:varchar(255);uniqueIndex:uk_username;default:null"`
	Email         string     `gorm:"type:varchar(255);uniqueIndex:uk_email;default:null"`
	Phone         string     `gorm:"type:varchar(255);uniqueIndex:uk_phone;default:null"`
	Status        int64      `gorm:"type:tinyint;not null;default:1;comment:Status, 1-enabled, 2-disabled"`
	KycStatus     int64      `gorm:"type:tinyint;not null;default:0;comment:KYC status, 0-not verify, 1-pending, 2-verified, 3-failed"`
	GoogleKey     string     `gorm:"type:varchar(255);not null;default:''"`
	GoogleStatus  string     `gorm:"type:enum('unbind','disabled','enabled');not null;default:unbind"`
	Pid           string     `gorm:"type:varchar(255);index:idx_pid;not null;default:''"`
	InviteCount   uint64     `gorm:"type:int;not null;default:0"`
	Depth         int64      `gorm:"type:int;not null;default:0"`
	Age           int64      `gorm:"type:tinyint;not null;default:0"`
	Gender        string     `gorm:"type:enum('secret', 'male', 'female');not null;default:'secret';comment:Gender, male female secret"`
	Avatar        string     `gorm:"type:varchar(255);not null;default:''"`
	Password      string     `gorm:"type:varchar(255);not null;default:''"`
	PayPassword   string     `gorm:"type:varchar(255);not null;default:''"`
	LastLoginTime *time.Time `gorm:"type:timestamp;default:null"`
	LastLoginIP   string     `gorm:"type:varchar(255);not null;default:''"`
	LastLoginType string     `gorm:"type:varchar(255);not null;default:''"`
}

func (CreateUserTable) TableName() string {
	return "uc_user"
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
