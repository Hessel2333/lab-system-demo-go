package models

import (
	"time"

	"gorm.io/gorm"
)

type Reservation struct {
	gorm.Model
	InstrumentID uint      `json:"instrument_id"`
	UserID       string    `json:"user_id"`
	UserName     string    `json:"user_name"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	Type         string    `json:"type"` // usage, maintenance
	Description  string    `json:"description"`
	Status       string    `json:"status"` // active, cancelled
}
