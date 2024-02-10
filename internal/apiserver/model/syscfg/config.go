package syscfg

import (
	"gorm.io/gorm"
)

type Config struct {
	gorm.Model

	Name        string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description string `gorm:"column:description;type:varchar(1024);not null" json:"description"`
	Key         string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value       string `gorm:"column:value;type:json" json:"value"`
	OperatorID  int32  `gorm:"column:operator_id;type:tinyint;not null" json:"operatorId"`
}

func (*Config) TableName() string {
	return "sys_cfg_config"
}
