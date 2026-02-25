package models

import "gorm.io/gorm"

// UserQualification represents the permission for a user to reserve a specific instrument
type UserQualification struct {
	gorm.Model
	UserID       uint       `json:"user_id"`
	User         User       `json:"user"`
	InstrumentID uint       `json:"instrument_id"`
	Instrument   Instrument `json:"instrument"`
	Status       bool       `json:"status" gorm:"default:true"` // true means authorized
}
