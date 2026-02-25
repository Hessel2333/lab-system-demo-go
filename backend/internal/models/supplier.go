package models

import "gorm.io/gorm"

type Supplier struct {
	gorm.Model
	Name          string  `json:"name"`
	Type          string  `json:"type"` // instrument, reagent, consumable, general
	ContactPerson string  `json:"contact_person"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	Address       string  `json:"address"`
	Rating        float64 `json:"rating"`         // 0-5 stars
	ResponseSpeed float64 `json:"response_speed"` // 0-5 stars
	Status        string  `json:"status"`         // active, blacklist
}
