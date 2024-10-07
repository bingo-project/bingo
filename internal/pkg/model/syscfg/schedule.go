package syscfg

import "bingo/internal/pkg/model"

type Schedule struct {
	model.Base

	Name        string         `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Job         string         `gorm:"column:job;type:varchar(255);not null;uniqueIndex:uk_job,priority:1" json:"job"`
	Spec        string         `gorm:"column:spec;type:varchar(255);not null" json:"spec"`
	Status      ScheduleStatus `gorm:"column:status;type:tinyint;not null;default:1;comment:Status, 1-enabled, 2-disabled" json:"status"` // Status, 1-enabled, 2-disabled
	Description string         `gorm:"column:description;type:varchar(1000);not null" json:"description"`
}

func (*Schedule) TableName() string {
	return "sys_schedule"
}

// ScheduleStatus 1-enabled, 2-disabled
type ScheduleStatus int

const (
	ScheduleStatusEnabled  ScheduleStatus = 1
	ScheduleStatusDisabled ScheduleStatus = 2
)
