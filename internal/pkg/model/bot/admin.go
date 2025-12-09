package bot

import "github.com/bingo-project/bingo/internal/pkg/model"

type Admin struct {
	model.Base

	Source string `gorm:"column:source;type:varchar(255);not null;uniqueIndex:uk_source_user,priority:1" json:"source"`
	UserID string `gorm:"column:user_id;type:varchar(255);not null;uniqueIndex:uk_source_user,priority:2" json:"userId"`
	Role   Role   `gorm:"column:role;type:varchar(255);not null" json:"role"`
}

func (*Admin) TableName() string {
	return "sys_bot_admin"
}

type Role string

const (
	RoleRoot  = "root"
	RoleAdmin = "admin"
)
