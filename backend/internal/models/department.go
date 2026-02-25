package models

import "gorm.io/gorm"

type Department struct {
	gorm.Model
	Name     string       `json:"name"`
	ParentID *uint        `json:"parent_id"` // Pointer to allow null for root
	Type     string       `json:"type"`      // institute, team, department
	Children []Department `json:"children" gorm:"foreignKey:ParentID"`
}
