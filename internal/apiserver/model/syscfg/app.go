package syscfg

import "bingo/internal/apiserver/model"

type App struct {
	model.Base

	Name        string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Version     string `gorm:"column:version;type:varchar(255);not null" json:"version"`
	Description string `gorm:"column:description;type:varchar(1000);not null" json:"description"`
	AboutUs     string `gorm:"column:about_us;type:varchar(2000);not null" json:"aboutUs"`
	Logo        string `gorm:"column:logo;type:varchar(255);not null" json:"logo"`
	Enabled     int32  `gorm:"column:enabled;type:tinyint;not null;comment:Is enabled" json:"enabled"` // Is enabled
}

func (*App) TableName() string {
	return "sys_cfg_app"
}
