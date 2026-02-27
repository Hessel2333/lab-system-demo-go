package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username             string     `json:"username" gorm:"uniqueIndex"`
	RealName             string     `json:"real_name"`
	DepartmentID         uint       `json:"department_id"`
	Role                 string     `json:"role"` // team_leader, member, specialist
	IsDispenseKeyHolderA bool       `json:"is_dispense_key_holder_a" gorm:"default:false"`
	IsDispenseKeyHolderB bool       `json:"is_dispense_key_holder_b" gorm:"default:false"`
	Department           Department `json:"department"`
}
