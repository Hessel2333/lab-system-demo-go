package handlers

import (
	"math/rand"
	"net/http"
	"time"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// FixInstrumentDates backfills missing purchase dates
func FixInstrumentDates(c *gin.Context) {
	var instruments []models.Instrument
	// 0. Seed Suppliers if none exist
	var supplierCount int64
	database.DB.Model(&models.Supplier{}).Count(&supplierCount)
	if supplierCount == 0 {
		suppliers := []models.Supplier{
			{Name: "安捷伦科技(中国)有限公司", Type: "instrument", ContactPerson: "张经理", Rating: 4.8, ResponseSpeed: 4.9, Status: "active"},
			{Name: "赛默飞世尔科技", Type: "instrument", ContactPerson: "李总监", Rating: 4.7, ResponseSpeed: 4.6, Status: "active"},
			{Name: "岛津企业管理(中国)有限公司", Type: "instrument", ContactPerson: "王工", Rating: 4.5, ResponseSpeed: 4.2, Status: "active"},
			{Name: "国药集团化学试剂有限公司", Type: "reagent", ContactPerson: "赵专员", Rating: 4.2, ResponseSpeed: 4.0, Status: "active"},
		}
		database.DB.Create(&suppliers)
	}

	database.DB.Find(&instruments)

	var suppliers []models.Supplier
	database.DB.Find(&suppliers)

	var departments []models.Department
	database.DB.Find(&departments)

	count := 0
	for _, ins := range instruments {
		updates := make(map[string]interface{})

		// 1. Ensure PurchaseDate (Arrival) exists
		if ins.PurchaseDate.IsZero() || ins.PurchaseDate.Year() < 2000 {
			days := rand.Intn(730)
			ins.PurchaseDate = time.Date(2022, 1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, days)
			updates["purchase_date"] = ins.PurchaseDate
		}

		// 2. Backfill Procurement Date (3-12 months before Arrival)
		if ins.ProcurementDate.IsZero() {
			months := rand.Intn(9) + 3 // 3 to 12
			ins.ProcurementDate = ins.PurchaseDate.AddDate(0, -months, 0)
			updates["procurement_date"] = ins.ProcurementDate
		}

		// 3. Backfill Planning Date (1-2 months before Procurement)
		if ins.PlanningDate.IsZero() {
			months := rand.Intn(2) + 1 // 1 to 3
			ins.PlanningDate = ins.ProcurementDate.AddDate(0, -months, 0)
			updates["planning_date"] = ins.PlanningDate
		}

		// 4. Link Department (Random) - Force Update to ensure validity
		if len(departments) > 0 {
			// Prefer non-root departments
			dept := departments[rand.Intn(len(departments))]
			if dept.ParentID != nil {
				updates["department_id"] = dept.ID
			} else if len(departments) > 1 {
				updates["department_id"] = departments[1].ID
			} else {
				updates["department_id"] = dept.ID
			}
		}

		// 5. Link Supplier (Random) - Force Update
		if len(suppliers) > 0 {
			sup := suppliers[rand.Intn(len(suppliers))]
			updates["supplier_id"] = sup.ID
		}

		if len(updates) > 0 {
			database.DB.Model(&ins).Updates(updates)
			count++
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Fixed dates", "updated_count": count})
}

// GetInstruments list all instruments
func GetInstruments(c *gin.Context) {
	var instruments []models.Instrument
	result := database.DB.Find(&instruments)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, instruments)
}

// CreateInstrument adds a new instrument
func CreateInstrument(c *gin.Context) {
	var input models.Instrument
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Defaults
	if input.Status == "" {
		input.Status = "arrival"
	}
	if input.LifecycleStage == "" {
		input.LifecycleStage = models.StageArrival
	}
	input.PurchaseDate = time.Now()

	database.DB.Create(&input)
	c.JSON(http.StatusOK, input)
}

// GetInstrument returns a single instrument with manual relation fetching
func GetInstrument(c *gin.Context) {
	var instrument models.Instrument
	// First fetch instrument without preloads to ensure we get the ID
	if result := database.DB.First(&instrument, c.Param("id")); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instrument not found"})
		return
	}

	// Manually fetch relations to fix Preload/JSON issue
	if instrument.DepartmentID != nil {
		var dept models.Department
		if err := database.DB.First(&dept, *instrument.DepartmentID).Error; err == nil {
			instrument.Department = &dept
		}
	}

	if instrument.SupplierID != nil {
		var sup models.Supplier
		if err := database.DB.First(&sup, *instrument.SupplierID).Error; err == nil {
			instrument.Supplier = &sup
		}
	}

	// Load documents
	database.DB.Model(&instrument).Association("Documents").Find(&instrument.Documents)

	c.JSON(http.StatusOK, instrument)
}

// UpdateInstrumentStatus updates just the status
func UpdateInstrumentStatus(c *gin.Context) {
	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var instrument models.Instrument
	if result := database.DB.First(&instrument, c.Param("id")); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instrument not found"})
		return
	}

	instrument.Status = input.Status
	database.DB.Save(&instrument)
	c.JSON(http.StatusOK, instrument)
}

// SeedInstruments creates initial demo data
func SeedInstruments(c *gin.Context) {
	instruments := []models.Instrument{
		{Name: "高性能液相色谱仪", DeviceModel: "HPLC-2024", Brand: "Agilent", Status: "active", Location: "分析室 302", RunTime: 1240, Health: 98, Reservations: 45, LifecycleStage: models.StageActive},
		{Name: "台式高速冷冻离心机", DeviceModel: "CF-5000", Brand: "Thermo", Status: "maintenance", Location: "前处理室 101", RunTime: 450, Health: 85, Reservations: 12, LifecycleStage: models.StageMaintenance},
		{Name: "扫描电子显微镜", DeviceModel: "SEM-X1", Brand: "Zeiss", Status: "arrival", Location: "待定", RunTime: 0, Health: 100, Reservations: 0, LifecycleStage: models.StageArrival},
		{Name: "电子万能试验机", DeviceModel: "UTM-200", Brand: "Instron", Status: "active", Location: "力学室 205", RunTime: 890, Health: 92, Reservations: 28, LifecycleStage: models.StageActive},
	}

	for _, ins := range instruments {
		database.DB.FirstOrCreate(&ins, models.Instrument{Name: ins.Name})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Seeded successfully"})
}
