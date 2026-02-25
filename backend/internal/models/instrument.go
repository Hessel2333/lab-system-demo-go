package models

import (
	"time"

	"gorm.io/gorm"
)

// Instrument Lifecycle Stages
const (
	StagePlanning    = "planning"
	StageProcurement = "procurement"
	StageArrival     = "arrival"
	StageActive      = "active"
	StageMaintenance = "maintenance"
	StageRetired     = "retired"
)

type Instrument struct {
	gorm.Model
	Name            string    `json:"name"`
	DeviceModel     string    `json:"model"`
	Brand           string    `json:"brand"`
	Status          string    `json:"status"` // active, in_use, maintenance, fault, retired, arrival
	Location        string    `json:"location"`
	PurchaseDate    time.Time `json:"purchase_date"` // Arrival/Acceptance Date
	PlanningDate    time.Time `json:"planning_date"`
	ProcurementDate time.Time `json:"procurement_date"`
	Admin           string    `json:"admin"`

	// Relations
	DepartmentID *uint       `json:"department_id"` // Requesting Department
	Department   *Department `json:"department" gorm:"foreignKey:DepartmentID"`
	SupplierID   *uint       `json:"supplier_id"` // Winning Supplier
	Supplier     *Supplier   `json:"supplier" gorm:"foreignKey:SupplierID"`

	// Stats
	RunTime      int `json:"run_time"` // Hours
	Health       int `json:"health"`   // Percentage (0-100)
	Reservations int `json:"reservations_count"`

	// Lifecycle
	LifecycleStage    string  `json:"lifecycle_stage"`
	ProcurementBudget float64 `json:"budget"`             // Budget for procurement
	ApplicationReason string  `json:"application_reason"` // Reason for purchase

	// Documentation
	Documents []InstrumentDoc `json:"documents" gorm:"serializer:json"`
}

type InstrumentDoc struct {
	Name       string `json:"name"`
	Type       string `json:"type"` // pdf, word, img, other
	Url        string `json:"url"`
	UploadDate string `json:"upload_date"`
}
