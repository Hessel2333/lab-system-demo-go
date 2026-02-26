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

	// BPM-A: 采购闭环相关
	OrderReference  string     `json:"order_reference"`  // 外部平台订单号（如易派客）
	OrderAttachment string     `json:"order_attachment"` // 下单截图/凭证附件路径
	ClosedAt        *time.Time `json:"closed_at"`        // BPM-A 闭环时间

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

	Logs []ReagentLog `gorm:"foreignKey:ReagentItemID;references:UUID" json:"logs"`

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

// --------------- BPM-B: 采购批次导入 ---------------

// ProcurementBatch 采购批次（对应一次 Excel 导入）
type ProcurementBatch struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	UploaderID  uint   `json:"uploader_id"`
	Uploader    User   `gorm:"foreignKey:UploaderID" json:"uploader"`
	SourceFile  string `json:"source_file"`  // 上传的原始文件路径
	Period      string `json:"period"`       // 所属周期，如 "2026-01"
	OrderNumber string `json:"order_number"` // 订单编号，防重入
	Status      string `json:"status"`       // 解析中 / 待确认 / 已确认

	Items []ProcurementBatchItem `gorm:"foreignKey:BatchID" json:"items"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// ProcurementBatchItem 批次明细行（每行对应 Excel 中的一条采购记录）
type ProcurementBatchItem struct {
	ID      uint             `gorm:"primaryKey" json:"id"`
	BatchID uint             `json:"batch_id"`
	Batch   ProcurementBatch `gorm:"foreignKey:BatchID" json:"batch"`
	RowHash string           `gorm:"index;size:128" json:"row_hash"` // 行级去重键

	// 原始 Excel 数据
	ReagentName      string  `json:"reagent_name"` // Excel 中的商品名称
	CASNumber        string  `json:"cas_number"`   // 解析出的 CAS 号
	Quantity         int     `json:"quantity"`
	Unit             string  `json:"unit"`
	UnitPrice        float64 `json:"unit_price"`
	Supplier         string  `json:"supplier"`          // 供应商（来自 Excel）
	MaterialCategory string  `json:"material_category"` // 物资类别
	ProductCategory  string  `json:"product_category"`  // 商品类别

	// 匹配结果
	MatchedCatalogID *uint  `json:"matched_catalog_id"`                  // 匹配到的品目 ID
	MatchedRequestID *uint  `json:"matched_request_id"`                  // 匹配到的申购单 ID
	MatchedUserID    *uint  `json:"matched_user_id"`                     // 匹配到的直接使用人 ID (无申购单直接发库用)
	MatchStatus      string `json:"match_status"`                        // 自动匹配 / 手动匹配 / 未匹配
	ReceiveStatus    string `json:"receive_status" gorm:"default:'待收货'"` // 待收货 / 部分收货 / 已收货
	ReceivedQuantity int    `json:"received_quantity" gorm:"default:0"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// --------------- 领用审批与双人双锁 ---------------

// ReagentDispenseRequest 试剂领用申请单
type ReagentDispenseRequest struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	RequesterID   uint        `json:"requester_id"`
	Requester     User        `gorm:"foreignKey:RequesterID" json:"requester"`
	ReagentItemID string      `gorm:"type:char(36)" json:"reagent_item_id"`
	ReagentItem   ReagentItem `gorm:"foreignKey:ReagentItemID;references:UUID" json:"reagent_item"`

	Amount  float64 `json:"amount"`  // 领取量
	Purpose string  `json:"purpose"` // 用途 / 关联实验
	Status  string  `json:"status"`  // 待审批 / 已通过 / 已驳回 / 待双签 / 已完成

	// 团队长审批
	LeaderID         *uint      `json:"leader_id"`
	Leader           User       `gorm:"foreignKey:LeaderID" json:"leader"`
	LeaderApprovedAt *time.Time `json:"leader_approved_at"`
	LeaderRejectMsg  string     `json:"leader_reject_msg"` // 驳回原因

	// 管控品双人双锁
	KeyHolderAID          *uint      `json:"key_holder_a_id"`
	KeyHolderA            User       `gorm:"foreignKey:KeyHolderAID" json:"key_holder_a"`
	KeyHolderBID          *uint      `json:"key_holder_b_id"`
	KeyHolderB            User       `gorm:"foreignKey:KeyHolderBID" json:"key_holder_b"`
	KeyHolderAConfirmedAt *time.Time `json:"key_holder_a_confirmed_at"`
	KeyHolderBConfirmedAt *time.Time `json:"key_holder_b_confirmed_at"`
	KeyHolderRejectMsg    string     `json:"key_holder_reject_msg"` // 钥匙持有人驳回原因
	ExpiresAt             *time.Time `json:"expires_at"`            // 双签超时（如 24h）

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
