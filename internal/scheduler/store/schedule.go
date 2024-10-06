package store

import (
	"context"

	"gorm.io/gorm"

	"bingo/internal/pkg/model"
)

type ScheduleStore interface {
	AllEnabled(ctx context.Context) ([]*model.Schedule, error)
}

type schedules struct {
	db *gorm.DB
}

var _ ScheduleStore = (*schedules)(nil)

func NewSchedules(db *gorm.DB) *schedules {
	return &schedules{db: db}
}

func (s *schedules) AllEnabled(ctx context.Context) (ret []*model.Schedule, err error) {
	err = s.db.WithContext(ctx).
		Where("status = ?", model.ScheduleStatusEnabled).
		Find(&ret).
		Error

	return
}
