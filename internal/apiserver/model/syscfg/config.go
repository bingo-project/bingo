package syscfg

import "bingo/internal/apiserver/model"

type Config struct {
	model.Base

	Name        string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description string `gorm:"column:description;type:varchar(1024);not null" json:"description"`
	Key         CfgKey `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value       string `gorm:"column:value;type:json" json:"value"`
	OperatorID  int32  `gorm:"column:operator_id;type:tinyint;not null" json:"operatorId"`
}

func (*Config) TableName() string {
	return "sys_cfg_config"
}

type CfgKey string
type ServerStatus string

const (
	CfgKeyServer CfgKey = "server" // server config

	ServerStatusOK          ServerStatus = "ok"          // server is ok.
	ServerStatusMaintenance ServerStatus = "maintenance" // server under maintenance.
)

type ServerConfig struct {
	Status ServerStatus `json:"status"` // server status: ok, maintenance
}
