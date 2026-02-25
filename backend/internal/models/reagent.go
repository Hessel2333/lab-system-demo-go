package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReagentCatalog represents the standard library of reagents
type ReagentCatalog struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	CASNumber        string `gorm:"uniqueIndex;not null" json:"cas_number"`
	Name             string `gorm:"not null" json:"name"`
	Alias            string `json:"alias"`
	Formula          string `json:"formula"`
	Category         string `json:"category"`      // Organic Solvent, Inorganic Salt, Acid, Base, etc.
	IsControlled     bool   `json:"is_controlled"` // Precursor Chemicals / Explosives
	Description      string `json:"description"`
	Storage          string `json:"storage"`           // e.g., "Room Temp", "4°C", "-20°C"
	AlertThreshold   int    `json:"alert_threshold"`   // Low stock alert threshold
	Unit             string `json:"unit"`              // e.g., "500ml", "100g"
	ChemicalLabels   string `json:"chemical_labels"`   // JSON array: ["危险化学品","易制毒化学品"]
	Aliases          string `json:"aliases"`           // 逗号分隔别称: "乙醇,酒精,Ethanol"
	StorageCondition string `json:"storage_condition"` // 详细储存条件: "阴凉干燥处,远离火源"
	PhysicalState    string `json:"physical_state"`    // 物态: 固体/液体/气体

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ReagentRequest represents a procurement request
type ReagentRequest struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	RequestorID      uint           `json:"requestor_id"`
	Requestor        User           `gorm:"foreignKey:RequestorID" json:"requestor"`
	ReagentCatalogID uint           `json:"reagent_catalog_id"`
	ReagentCatalog   ReagentCatalog `gorm:"foreignKey:ReagentCatalogID" json:"reagent_catalog"`
	Quantity         int            `json:"quantity"` // Number of bottles/units
	Status           string         `json:"status"`   // Pending, Approved, Ordered, Fulfilled, Rejected
	Remarks          string         `json:"remarks"`

	// Extended Fields
	RequestType      string `json:"request_type"`      // '日常', '储备', '紧急'
	ExpectedDelivery string `json:"expected_delivery"` // YYYY-MM-DD
	ProjectName      string `json:"project_name"`
	ProjectID        string `json:"project_id"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ReagentItem represents a specific physical bottle/item (Serialized)
type ReagentItem struct {
	UUID             string         `gorm:"primaryKey;type:char(36)" json:"uuid"`
	ReagentRequestID uint           `json:"reagent_request_id"`
	ReagentRequest   ReagentRequest `gorm:"foreignKey:ReagentRequestID" json:"reagent_request"`
	ReagentCatalogID uint           `json:"reagent_catalog_id"`
	ReagentCatalog   ReagentCatalog `gorm:"foreignKey:ReagentCatalogID" json:"reagent_catalog"`

	Status    string         `json:"status"`     // Arrived, InStorage, Used, Expired, Disposed
	Location  string         `json:"location"`   // Specific room/shelf code
	CabinetID uint           `json:"cabinet_id"` // 关联试剂柜 ID（0=未指定）
	Cabinet   ReagentCabinet `gorm:"foreignKey:CabinetID" json:"cabinet"`

	Capacity        float64 `json:"capacity"`         // Initial volume/weight
	RemainingVolume float64 `json:"remaining_volume"` // Current remaining volume/weight

	BatchNumber string    `json:"batch_number"`
	ExpiryDate  time.Time `json:"expiry_date"`

	CreatedAt time.Time      `json:"created_at"` // Arrival time
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ReagentLog represents an activity log for a specific reagent item
type ReagentLog struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	ReagentItemID string      `gorm:"type:char(36)" json:"reagent_item_id"`
	ReagentItem   ReagentItem `gorm:"foreignKey:ReagentItemID;references:UUID" json:"reagent_item"`
	UserID        uint        `json:"user_id"`
	User          User        `gorm:"foreignKey:UserID" json:"user"`
	Action        string      `json:"action"`   // "CheckIn", "Consume", "Dispose", "UpdateLocation"
	Quantity      float64     `json:"quantity"` // Amount consumed/changed
	Remarks       string      `json:"remarks"`

	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ReagentCabinet 表示实验室中的一个试剂柜点位
type ReagentCabinet struct {
	ID           uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string `gorm:"not null" json:"name"` // 柜子名称，如"分析团队-普通柜A"
	CabinetType  string `json:"cabinet_type"`         // "普通试剂柜" / "易制毒制爆试剂柜"
	DepartmentID uint   `json:"department_id"`        // 所属团队ID（0=公共）
	Location     string `json:"location"`             // 所在房间号，如"E309"
	Notes        string `json:"notes"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID for ReagentItem
func (item *ReagentItem) BeforeCreate(tx *gorm.DB) (err error) {
	if item.UUID == "" {
		item.UUID = uuid.New().String()
	}
	return
}
