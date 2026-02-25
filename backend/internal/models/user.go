package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username     string     `json:"username" gorm:"uniqueIndex"`
	RealName     string     `json:"real_name"`
	DepartmentID uint       `json:"department_id"`
	Role         string     `json:"role"` // team_leader, member, specialist
	Department   Department `json:"department"`
}
