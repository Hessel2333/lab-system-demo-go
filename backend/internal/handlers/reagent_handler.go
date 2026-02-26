package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var reagentStateTransitions = map[string]map[string]bool{
	"已到货": {
		"在库": true,
	},
	"在库": {
		"已耗尽": true,
	},
	"已耗尽": {},
}

func isValidItemTransition(from, to string) bool {
	if from == to {
		return true
	}
	next, ok := reagentStateTransitions[from]
	if !ok {
		return false
	}
	return next[to]
}

func getUserIDOrDefault(c *gin.Context, fallback int) uint {
	raw := strings.TrimSpace(c.GetHeader("X-User-ID"))
	if raw == "" {
		return uint(fallback)
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return uint(fallback)
	}
	return uint(id)
}

func procurementItemFingerprint(item models.ProcurementBatchItem) string {
	if rowHash := strings.TrimSpace(item.RowHash); rowHash != "" {
		return strings.ToLower(rowHash)
	}
	return strings.Join([]string{
		strings.ToLower(strings.TrimSpace(item.ReagentName)),
		strings.ToLower(strings.TrimSpace(item.CASNumber)),
		fmt.Sprintf("%d", item.Quantity),
		strings.ToLower(strings.TrimSpace(item.Unit)),
		fmt.Sprintf("%.4f", item.UnitPrice),
		strings.ToLower(strings.TrimSpace(item.Supplier)),
		strings.ToLower(strings.TrimSpace(item.MaterialCategory)),
		strings.ToLower(strings.TrimSpace(item.ProductCategory)),
	}, "|")
}

type importDedupStats struct {
	TotalItems         int      `json:"total_items"`
	CreatedCount       int      `json:"created_count"`
	SkippedCount       int      `json:"skipped_count"`
	DuplicateInFile    int      `json:"duplicate_in_file"`
	DuplicateInSystem  int      `json:"duplicate_in_system"`
	DuplicateNameHints []string `json:"duplicate_name_hints"`
}

func dedupeProcurementItems(orderNumber string, items []models.ProcurementBatchItem) ([]models.ProcurementBatchItem, importDedupStats) {
	stats := importDedupStats{
		TotalItems: len(items),
	}
	orderNumber = strings.TrimSpace(orderNumber)

	existingFP := map[string]bool{}
	if orderNumber != "" {
		var importedItems []models.ProcurementBatchItem
		database.DB.
			Joins("JOIN procurement_batches pb ON pb.id = procurement_batch_items.batch_id").
			Where("pb.order_number = ?", orderNumber).
			Find(&importedItems)
		for _, it := range importedItems {
			fp := strings.TrimSpace(it.RowHash)
			if fp == "" {
				fp = procurementItemFingerprint(it)
			}
			existingFP[fp] = true
		}
	}

	localFP := map[string]bool{}
	dupNameSet := map[string]bool{}
	newItems := make([]models.ProcurementBatchItem, 0, len(items))
	for _, item := range items {
		fp := procurementItemFingerprint(item)
		if existingFP[fp] {
			stats.DuplicateInSystem++
			stats.SkippedCount++
			name := strings.TrimSpace(item.ReagentName)
			if name != "" && len(stats.DuplicateNameHints) < 8 && !dupNameSet[name] {
				stats.DuplicateNameHints = append(stats.DuplicateNameHints, name)
				dupNameSet[name] = true
			}
			continue
		}
		if localFP[fp] {
			stats.DuplicateInFile++
			stats.SkippedCount++
			continue
		}
		localFP[fp] = true
		item.RowHash = fp
		newItems = append(newItems, item)
	}

	stats.CreatedCount = len(newItems)
	return newItems, stats
}

// --- Reagent Catalog ---

func GetReagentCatalogs(c *gin.Context) {
	var catalogs []models.ReagentCatalog
	tx := database.DB

	// 搜索过滤：按名称、CAS、别称模糊匹配
	if search := c.Query("search"); search != "" {
		q := "%" + search + "%"
		tx = tx.Where("name LIKE ? OR cas_number LIKE ? OR aliases LIKE ? OR alias LIKE ?", q, q, q, q)
	}

	// 标签过滤：chemical_labels JSON 中包含该标签
	if label := c.Query("label"); label != "" {
		tx = tx.Where("chemical_labels LIKE ?", "%"+label+"%")
	}

	if err := tx.Find(&catalogs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, catalogs)
}

// GetReagentDashboardStats returns aggregated metrics and recent logs for the dashboard
func GetReagentDashboardStats(c *gin.Context) {
	var totalItems int64
	var inStorageItems int64
	var pendingRequests int64
	var lowStockAlerts int64
	var pendingCheckInItems int64
	var pendingReceiveLines int64
	var controlledInStorage int64
	var nearExpiry30Days int64
	var activeCatalogs int64
	var consumedVolume7d float64

	database.DB.Model(&models.ReagentItem{}).Where("status != ?", "已耗尽").Count(&totalItems)
	database.DB.Model(&models.ReagentItem{}).Where("reagent_items.status = ?", "在库").Count(&inStorageItems)
	database.DB.Model(&models.ReagentRequest{}).Where("status IN ?", []string{"待审批", "待采购"}).Count(&pendingRequests)
	database.DB.Model(&models.ReagentItem{}).Where("status = ?", "已到货").Count(&pendingCheckInItems)
	database.DB.Model(&models.ProcurementBatchItem{}).Where("receive_status != ?", "已收货").Count(&pendingReceiveLines)
	database.DB.Model(&models.ReagentItem{}).
		Joins("JOIN reagent_catalogs c ON c.id = reagent_items.reagent_catalog_id").
		Where("reagent_items.status = ? AND c.is_controlled = ?", "在库", true).
		Count(&controlledInStorage)
	database.DB.Model(&models.ReagentItem{}).
		Where("status = ? AND expiry_date >= date('now') AND expiry_date <= date('now', '+30 day')", "在库").
		Count(&nearExpiry30Days)
	database.DB.Model(&models.ReagentItem{}).
		Where("status != ?", "已耗尽").
		Distinct("reagent_catalog_id").
		Count(&activeCatalogs)
	database.DB.Model(&models.ReagentLog{}).
		Where("action IN ? AND created_at >= date('now', '-7 days')", []string{"领用消耗", "空瓶核销"}).
		Select("COALESCE(SUM(quantity), 0)").Scan(&consumedVolume7d)

	// Low Stock Alert (Checking catalogs against active items count)
	// For simplicity, count how many catalogs have fewer active items than their threshold
	// Assuming AlertThreshold > 0
	type CatalogStock struct {
		ID             uint
		Name           string
		AlertThreshold int
		Count          int
	}
	var stocks []CatalogStock

	// Complex query simplified: get all items grouped by catalog, compare with catalog threshold
	database.DB.Raw(`
		SELECT c.id, c.name, c.alert_threshold, COUNT(i.uuid) as count
		FROM reagent_catalogs c
		LEFT JOIN reagent_items i ON c.id = i.reagent_catalog_id AND i.status != '已耗尽'
		WHERE c.deleted_at IS NULL
		GROUP BY c.id, c.name, c.alert_threshold
		HAVING COUNT(i.uuid) <= c.alert_threshold AND c.alert_threshold > 0
	`).Scan(&stocks)
	lowStockAlerts = int64(len(stocks))

	// Get Recent Activity Logs (last 5)
	var recentLogs []models.ReagentLog
	database.DB.Preload("User").Preload("ReagentItem.ReagentCatalog").
		Order("reagent_logs.created_at desc").Limit(5).Find(&recentLogs)

	// categoryDistribution for charts
	type CategoryStat struct {
		Category string `json:"category"`
		Count    int    `json:"count"`
	}
	var catStats []CategoryStat
	database.DB.Raw(`
		SELECT c.category, COUNT(i.uuid) as count 
		FROM reagent_items i 
		JOIN reagent_catalogs c ON i.reagent_catalog_id = c.id 
		WHERE i.status != '已耗尽' 
		GROUP BY c.category
	`).Scan(&catStats)

	// recentUsageTrend (last 7 days consumption)
	type TrendStat struct {
		Date  string `json:"date"`
		Count int    `json:"count"`
	}
	var trendStats []TrendStat

	// SQLite specific date truncation for trend line (counting how many items were 'used' per day)
	database.DB.Raw(`
		SELECT strftime('%Y-%m-%d', created_at) as date, COUNT(*) as count
		FROM reagent_logs 
		WHERE action = '空瓶核销' AND created_at >= date('now', '-7 days')
		GROUP BY date
		ORDER BY date ASC
	`).Scan(&trendStats)

	c.JSON(http.StatusOK, gin.H{
		"total_items":           totalItems,
		"in_storage_items":      inStorageItems,
		"pending_requests":      pendingRequests,
		"low_stock_alerts":      lowStockAlerts,
		"pending_checkin_items": pendingCheckInItems,
		"pending_receive_lines": pendingReceiveLines,
		"controlled_in_storage": controlledInStorage,
		"near_expiry_30d":       nearExpiry30Days,
		"active_catalogs":       activeCatalogs,
		"consumed_volume_7d":    consumedVolume7d,
		"recent_logs":           recentLogs,
		"alerts":                stocks,
		"category_distribution": catStats,
		"recent_usage_trend":    trendStats,
	})
}

func CreateReagentCatalog(c *gin.Context) {
	var input models.ReagentCatalog
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func UpdateReagentCatalog(c *gin.Context) {
	id := c.Param("id")
	var catalog models.ReagentCatalog
	if err := database.DB.First(&catalog, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "品目不存在"})
		return
	}

	var input models.ReagentCatalog
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新可修改的字段
	catalog.Name = input.Name
	catalog.CASNumber = input.CASNumber
	catalog.Alias = input.Alias
	catalog.Formula = input.Formula
	catalog.Category = input.Category
	catalog.IsControlled = input.IsControlled
	catalog.Description = input.Description
	catalog.Storage = input.Storage
	catalog.AlertThreshold = input.AlertThreshold
	catalog.Unit = input.Unit
	catalog.ChemicalLabels = input.ChemicalLabels
	catalog.Aliases = input.Aliases
	catalog.StorageCondition = input.StorageCondition
	catalog.PhysicalState = input.PhysicalState

	// 根据标签自动判断 IsControlled
	if catalog.ChemicalLabels != "" && catalog.ChemicalLabels != "[]" && catalog.ChemicalLabels != "[\"普通化学品\"]" {
		catalog.IsControlled = true
	}

	if err := database.DB.Save(&catalog).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, catalog)
}

func DeleteReagentCatalog(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.ReagentCatalog{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "品目已删除"})
}

// StockCheck 查询某试剂的库存概况，供申购/审核时参考
func StockCheck(c *gin.Context) {
	casNumber := c.Query("cas_number")
	name := c.Query("name")

	if casNumber == "" && name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请提供 cas_number 或 name 参数"})
		return
	}

	// 查找品目
	var catalog models.ReagentCatalog
	tx := database.DB
	if casNumber != "" {
		tx = tx.Where("cas_number = ?", casNumber)
	} else {
		tx = tx.Where("name LIKE ? OR aliases LIKE ?", "%"+name+"%", "%"+name+"%")
	}
	if err := tx.First(&catalog).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到匹配的品目"})
		return
	}

	// 在库数量
	var inStock int64
	database.DB.Model(&models.ReagentItem{}).
		Where("reagent_catalog_id = ? AND reagent_items.status = ?", catalog.ID, "在库").
		Count(&inStock)

	// 待到货数量（已批准或采购中的申购单的总瓶数）
	var pendingArrival int64
	database.DB.Model(&models.ReagentRequest{}).
		Where("reagent_catalog_id = ? AND status IN (?, ?)", catalog.ID, "采购中", "已到货").
		Select("COALESCE(SUM(quantity), 0)").Scan(&pendingArrival)

	// 待审申购数量
	var pendingRequests int64
	database.DB.Model(&models.ReagentRequest{}).
		Where("reagent_catalog_id = ? AND reagent_requests.status = ?", catalog.ID, "待处理").
		Count(&pendingRequests)

	// 最近消耗时间
	var lastConsumed string
	database.DB.Model(&models.ReagentLog{}).
		Joins("JOIN reagent_items ri ON ri.uuid = reagent_logs.reagent_item_id").
		Where("ri.reagent_catalog_id = ? AND reagent_logs.action = ?", catalog.ID, "空瓶核销").
		Order("reagent_logs.created_at DESC").
		Limit(1).
		Pluck("strftime('%Y-%m-%d', reagent_logs.created_at)", &lastConsumed)

	// 简易 AI 建议（基于规则）
	advice := ""
	if inStock == 0 && pendingArrival == 0 {
		advice = "⚠️ 当前零库存，建议尽快审批采购"
	} else if inStock <= int64(catalog.AlertThreshold) {
		advice = "⚠️ 库存已低于预警阈值（" + fmt.Sprintf("%d", catalog.AlertThreshold) + " 瓶），建议优先审批"
	} else if inStock >= 10 {
		advice = "✅ 库存充足，建议评估是否有必要新增采购"
	} else {
		advice = "📦 库存正常，可按需采购"
	}

	// 管控品额外提示
	if catalog.IsControlled {
		advice += "｜🔒 该试剂为管控品，采购需双人审批"
	}

	c.JSON(http.StatusOK, gin.H{
		"catalog": gin.H{
			"id":              catalog.ID,
			"name":            catalog.Name,
			"cas_number":      catalog.CASNumber,
			"chemical_labels": catalog.ChemicalLabels,
			"unit":            catalog.Unit,
			"is_controlled":   catalog.IsControlled,
		},
		"in_stock":         inStock,
		"pending_arrival":  pendingArrival,
		"pending_requests": pendingRequests,
		"last_consumed_at": lastConsumed,
		"advice":           advice,
	})
}

// --- Reagent Requests ---

func GetReagentRequests(c *gin.Context) {
	var requests []models.ReagentRequest
	if err := database.DB.Preload("ReagentCatalog").Preload("Requestor").Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, requests)
}

func CreateReagentRequest(c *gin.Context) {
	var input models.ReagentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var catalog models.ReagentCatalog
	if err := database.DB.First(&catalog, input.ReagentCatalogID).Error; err == nil && catalog.IsControlled {
		input.Status = "待审批" // 管控品走团队长审批
	} else {
		input.Status = "待采购" // 普通品直达采购员
	}

	// Assuming RequestorID is passed or retrieved from context (mocking for now)
	if input.RequestorID == 0 {
		input.RequestorID = 1 // Default to Admin for demo
	}

	if err := database.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func LeaderApproveReagentRequest(c *gin.Context) {
	id := c.Param("id")
	var req models.ReagentRequest

	if err := database.DB.First(&req, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	if req.Status != "待审批" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有待审批状态的申请可以审核"})
		return
	}

	var body struct {
		Approved  bool   `json:"approved"`
		RejectMsg string `json:"reject_msg"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.Approved {
		req.Status = "待采购"
	} else {
		req.Status = "已驳回"
		// req.LeaderRejectMsg = body.RejectMsg // 可扩展记录驳回原因
	}

	if err := database.DB.Save(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update request format"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "审批完成"})
}

// MarkReagentRequestOrdered transitions a request from 待采购 to 已接单 (BPM-A 闭环)
func ApproveReagentRequest(c *gin.Context) {
	id := c.Param("id")
	var req models.ReagentRequest

	if err := database.DB.First(&req, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	if req.Status != "待采购" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只有待采购的申请可以标记已接单"})
		return
	}

	// 接收可选的外部订单凭证
	var body struct {
		OrderReference  string `json:"order_reference"`
		OrderAttachment string `json:"order_attachment"`
	}
	c.ShouldBindJSON(&body)

	now := time.Now()
	req.Status = "已接单" // BPM-A 闭环节点
	req.OrderReference = body.OrderReference
	req.OrderAttachment = body.OrderAttachment
	req.ClosedAt = &now

	if err := database.DB.Save(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as ordered"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已标记为接单/已下单", "closed_at": now})
}

// --- Reagent Items (Inventory) ---

// GetTeamInventory 按团队分组返回在库试剂台账
func GetTeamInventory(c *gin.Context) {
	// 支持按团队筛选
	departmentID := c.Query("department_id")

	// 查询在库 items
	var items []models.ReagentItem

	tx := database.DB.
		Preload("ReagentCatalog").
		Preload("Cabinet").
		Preload("ReagentRequest").
		Preload("ReagentRequest.Requestor").
		Preload("ReagentRequest.Requestor.Department").
		Where("reagent_items.status = ?", "在库")

	if departmentID != "" {
		// JOIN 到 user 表过滤团队
		tx = tx.Joins("JOIN reagent_requests rr ON rr.id = reagent_items.reagent_request_id").
			Joins("JOIN users u ON u.id = rr.requestor_id").
			Where("u.department_id = ?", departmentID)
	}

	if err := tx.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 按团队分组
	type TeamGroup struct {
		DepartmentID   uint                 `json:"department_id"`
		DepartmentName string               `json:"department_name"`
		Items          []models.ReagentItem `json:"items"`
		TotalCount     int                  `json:"total_count"`
	}

	groupMap := make(map[uint]*TeamGroup)
	for _, item := range items {
		var deptID uint = 0
		var deptName string = "公共库 / 未分配所属团队"

		if item.ReagentRequest.ID != 0 && item.ReagentRequest.Requestor.ID != 0 && item.ReagentRequest.Requestor.Department.ID != 0 {
			deptID = item.ReagentRequest.Requestor.DepartmentID
			deptName = item.ReagentRequest.Requestor.Department.Name
		}

		if _, ok := groupMap[deptID]; !ok {
			groupMap[deptID] = &TeamGroup{
				DepartmentID:   deptID,
				DepartmentName: deptName,
				Items:          []models.ReagentItem{},
			}
		}
		groupMap[deptID].Items = append(groupMap[deptID].Items, item)
		groupMap[deptID].TotalCount++
	}

	// 转换为有序 slice
	result := []*TeamGroup{}
	for _, g := range groupMap {
		result = append(result, g)
	}

	c.JSON(http.StatusOK, result)
}

func GetReagentItems(c *gin.Context) {
	var items []models.ReagentItem
	tx := database.DB.Preload("ReagentCatalog").Preload("Cabinet").Preload("ReagentRequest").Preload("ReagentRequest.Requestor")

	// 如果需要按申购人过滤，利用 GORM 关联进行 Join
	requestorID := c.Query("requestor_id")
	if requestorID != "" {
		tx = tx.Joins("JOIN reagent_requests rr ON rr.id = reagent_items.reagent_request_id").
			Where("rr.requestor_id = ?", requestorID)
	}

	// Optional Query Params
	if status := c.Query("status"); status != "" {
		tx = tx.Where("reagent_items.status = ?", status)
	}
	if requestID := c.Query("request_id"); requestID != "" {
		tx = tx.Where("reagent_items.reagent_request_id = ?", requestID)
	}
	if err := tx.Order("reagent_items.created_at DESC").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func GetReagentItemByUUID(c *gin.Context) {
	uuid := c.Param("uuid")
	var item models.ReagentItem
	tx := database.DB.
		Preload("ReagentCatalog").
		Preload("Cabinet").
		Preload("ReagentRequest").
		Preload("ReagentRequest.Requestor").
		Preload("Logs", func(db *gorm.DB) *gorm.DB {
			return db.Order("reagent_logs.created_at DESC")
		}).
		Preload("Logs.User").
		Where("uuid = ?", uuid)

	if err := tx.First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func UpdateReagentItemStatus(c *gin.Context) {
	uuid := c.Param("uuid")
	var input struct {
		Status    string `json:"status"`
		Location  string `json:"location"`
		CabinetID uint   `json:"cabinet_id"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var item models.ReagentItem
	if err := database.DB.Where("uuid = ?", uuid).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if input.Status != "" && !isValidItemTransition(item.Status, input.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("非法状态流转: %s -> %s", item.Status, input.Status),
		})
		return
	}

	if input.Status == "在库" && strings.TrimSpace(input.Location) == "" && strings.TrimSpace(item.Location) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "入库必须指定库位"})
		return
	}

	// Determine Action for logging based on status change
	action := "变更信息"
	remarks := ""
	if input.Status != "" && input.Status != item.Status {
		item.Status = input.Status
		switch item.Status {
		case "在库":
			action = "扫码入库"
			cabinetHint := ""
			if input.CabinetID > 0 {
				cabinetHint = fmt.Sprintf("，放入试剂柜#%d", input.CabinetID)
			}
			remarks = "移至库位: " + input.Location + cabinetHint
		case "已耗尽":
			action = "空瓶核销"
			remarks = "标记为耗尽并核销回收"
			item.RemainingVolume = 0
		default:
			action = "状态变更"
			remarks = "状态更改为 " + item.Status
		}
	}

	if input.Location != "" && input.Location != item.Location {
		item.Location = input.Location
		if action == "变更信息" {
			action = "库位移动"
			remarks = "库位变更为 " + input.Location
		}
	}

	if input.CabinetID > 0 {
		item.CabinetID = input.CabinetID
	}

	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create log entry if there was a meaningful change
	if action != "变更信息" {
		log := models.ReagentLog{
			ReagentItemID: item.UUID,
			UserID:        1, // Mock User ID
			Action:        action,
			Quantity:      0, // Keeping 0 for simple status updates for now
			Remarks:       remarks,
		}
		if err := database.DB.Create(&log).Error; err != nil {
			fmt.Printf("Failed to create log for item %s: %v\n", item.UUID, err)
		}
	}

	// 注意：此处不回写申购单状态。
	// ReagentRequest（BPM-A）的终态是「已接单」，由采购员确认时设置。
	// ReagentItem 的物理状态属于 BPM-B，两个流程通过外键关联，状态机独立。

	c.JSON(http.StatusOK, item)
}

// CheckInReagentItem 研发扫码后确认入库位置（专用动作，区别于通用状态修改）
func CheckInReagentItem(c *gin.Context) {
	uuid := c.Param("uuid")
	var input struct {
		LabRoom   string `json:"lab_room"`
		CabinetID uint   `json:"cabinet_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var item models.ReagentItem
	if err := database.DB.Preload("ReagentCatalog").Where("uuid = ?", uuid).First(&item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if item.Status != "已到货" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅“已到货”状态允许扫码入库"})
		return
	}

	var cabinet models.ReagentCabinet
	if err := database.DB.First(&cabinet, input.CabinetID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "试剂柜不存在"})
		return
	}

	if input.LabRoom != "" && input.LabRoom != cabinet.Location {
		c.JSON(http.StatusBadRequest, gin.H{"error": "实验室与试剂柜位置不一致"})
		return
	}

	if item.ReagentCatalog.IsControlled && cabinet.CabinetType != "易制毒制爆试剂柜" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "管控试剂必须入库至管控柜"})
		return
	}

	finalLocation := fmt.Sprintf("%s / %s", cabinet.Location, cabinet.Name)

	item.Status = "在库"
	item.Location = finalLocation
	item.CabinetID = input.CabinetID
	if err := database.DB.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log := models.ReagentLog{
		ReagentItemID: item.UUID,
		UserID:        1, // mock user
		Action:        "扫码入库",
		Quantity:      0,
		Remarks:       fmt.Sprintf("扫码入库至 %s", finalLocation),
	}
	_ = database.DB.Create(&log).Error

	c.JSON(http.StatusOK, item)
}

// PrintReagentLabels 批量打印标签（MVP：记录打印日志并返回可打印数据）
func PrintReagentLabels(c *gin.Context) {
	var input struct {
		UUIDs []string `json:"uuids" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var items []models.ReagentItem
	if err := database.DB.Preload("ReagentCatalog").Where("uuid IN ?", input.UUIDs).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No items found"})
		return
	}

	type labelPayload struct {
		UUID        string `json:"uuid"`
		Name        string `json:"name"`
		CASNumber   string `json:"cas_number"`
		BatchNumber string `json:"batch_number"`
		Status      string `json:"status"`
	}

	labels := make([]labelPayload, 0, len(items))
	for _, item := range items {
		labels = append(labels, labelPayload{
			UUID:        item.UUID,
			Name:        item.ReagentCatalog.Name,
			CASNumber:   item.ReagentCatalog.CASNumber,
			BatchNumber: item.BatchNumber,
			Status:      item.Status,
		})

		_ = database.DB.Create(&models.ReagentLog{
			ReagentItemID: item.UUID,
			UserID:        1,
			Action:        "二维码打印",
			Quantity:      0,
			Remarks:       "已生成并打印标签",
		}).Error
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "labels ready",
		"count":   len(labels),
		"labels":  labels,
	})
}

// ConsumeReagentItem 记录试剂的使用和余量扣减
func ConsumeReagentItem(c *gin.Context) {
	uuid := c.Param("uuid")
	var input struct {
		ConsumeVolume float64 `json:"consume_volume"`
		Remarks       string  `json:"remarks"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	var item models.ReagentItem
	if err := tx.Where("uuid = ?", uuid).First(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if item.Status != "在库" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only items '在库' can be consumed"})
		return
	}

	if input.ConsumeVolume <= 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Consume volume must be positive"})
		return
	}

	// Calculate new volume
	newVolume := item.RemainingVolume - input.ConsumeVolume
	if newVolume <= 0 {
		newVolume = 0
		item.Status = "已耗尽"
	}
	item.RemainingVolume = newVolume

	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log consumption
	action := "领用消耗"
	if item.Status == "已耗尽" {
		action = "空瓶核销"
	}
	remarks := input.Remarks
	if remarks == "" {
		remarks = fmt.Sprintf("本次消耗 %.2f，剩余 %.2f", input.ConsumeVolume, newVolume)
	}

	log := models.ReagentLog{
		ReagentItemID: item.UUID,
		UserID:        1, // Mock User ID
		Action:        action,
		Quantity:      input.ConsumeVolume,
		Remarks:       remarks,
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, item)
}

// DepleteReagentItem 统一“耗尽核销”动作（供台账/详情/扫码复用）
func DepleteReagentItem(c *gin.Context) {
	uuid := c.Param("uuid")
	var input struct {
		Remarks string `json:"remarks"`
	}
	_ = c.ShouldBindJSON(&input)

	tx := database.DB.Begin()

	var item models.ReagentItem
	if err := tx.Where("uuid = ?", uuid).First(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	if item.Status != "在库" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅“在库”状态允许耗尽核销"})
		return
	}

	consumed := item.RemainingVolume
	item.RemainingVolume = 0
	item.Status = "已耗尽"
	if err := tx.Save(&item).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	remarks := strings.TrimSpace(input.Remarks)
	if remarks == "" {
		remarks = "标记为耗尽并核销回收"
	}
	log := models.ReagentLog{
		ReagentItemID: item.UUID,
		UserID:        1,
		Action:        "空瓶核销",
		Quantity:      consumed,
		Remarks:       remarks,
	}
	if err := tx.Create(&log).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, item)
}

// --- AI Mock ---

func ParseReagentRequestAI(c *gin.Context) {
	var input struct {
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "未配置 GEMINI_API_KEY"})
		return
	}

	modelName := os.Getenv("GEMINI_MODEL_NAME")
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}

	// Call Real Gemini API
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, apiKey)

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": input.Message},
				},
			},
		},
		"systemInstruction": map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": `You are a reagent parser in a laboratory system. 
Parse the user's natural language request for reagents into JSON.
You MUST reply with ONLY a JSON object exactly matching this schema:
{
  "parsed_catalog": {
    "cas_number": "string (CAS format, e.g., 67-64-1)",
    "name": "string (Reagent name in Chinese, e.g., 丙酮)",
    "unit": "string (e.g., 500ml, 1kg, 1box)",
    "is_controlled": boolean (true if hazardous/controlled)
  },
  "quantity": integer,
  "request_type": "string (enum: 日常, 储备, 紧急, default: 日常)",
  "expected_delivery": "string (YYYY-MM-DD or fuzzy like '尽快', '下周' etc)",
  "project_name": "string (Associated project name)",
  "project_id": "string (Associated project code or ID)",
  "confidence": float (between 0.0 and 1.0)
}`},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseMimeType": "application/json",
		},
	}

	reqBytes, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode request"})
		return
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 请求失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("AI 接口返回错误 HTTP %d: %s", resp.StatusCode, string(bodyBytes))})
		return
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法解析 AI 响应数据"})
		return
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 没有返回有效内容"})
		return
	}

	jsonText := geminiResp.Candidates[0].Content.Parts[0].Text

	// Parse JSON strictly into map
	var parsedResult map[string]interface{}
	if err := json.Unmarshal([]byte(jsonText), &parsedResult); err != nil {
		fmt.Println("Failed to parse Gemini JSON:", jsonText)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 返回的数据格式不正确"})
		return
	}

	c.JSON(http.StatusOK, parsedResult)
}

// ===================== BPM-B: 采购批次导入 =====================

// CreateProcurementBatch 创建新的采购批次（上传 Excel 后调用）
func CreateProcurementBatch(c *gin.Context) {
	var input struct {
		Period      string                        `json:"period"`
		OrderNumber string                        `json:"order_number"`
		Items       []models.ProcurementBatchItem `json:"items"`
		DryRun      bool                          `json:"dry_run"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uploaderID := getUserIDOrDefault(c, 2)
	orderNumber := strings.TrimSpace(input.OrderNumber)

	var latestOrderBatch models.ProcurementBatch
	if orderNumber != "" {
		database.DB.Where("order_number = ?", orderNumber).Order("created_at DESC").First(&latestOrderBatch)
	}
	newItems, dedupStats := dedupeProcurementItems(orderNumber, input.Items)
	if input.DryRun {
		c.JSON(http.StatusOK, gin.H{
			"message":              "解析完成",
			"skipped_count":        dedupStats.SkippedCount,
			"created_count":        dedupStats.CreatedCount,
			"confirmable":          dedupStats.CreatedCount > 0,
			"duplicate_in_file":    dedupStats.DuplicateInFile,
			"duplicate_in_system":  dedupStats.DuplicateInSystem,
			"duplicate_name_hints": dedupStats.DuplicateNameHints,
		})
		return
	}

	// 同订单重复上传但无新增行：返回最近批次，避免报错
	if len(newItems) == 0 && latestOrderBatch.ID > 0 {
		database.DB.Preload("Items").Preload("Uploader").First(&latestOrderBatch, latestOrderBatch.ID)
		c.JSON(http.StatusOK, gin.H{
			"message":              "无新增条目，已跳过重复导入",
			"skipped_count":        dedupStats.SkippedCount,
			"created_count":        0,
			"confirmable":          false,
			"duplicate_in_file":    dedupStats.DuplicateInFile,
			"duplicate_in_system":  dedupStats.DuplicateInSystem,
			"duplicate_name_hints": dedupStats.DuplicateNameHints,
			"id":                   latestOrderBatch.ID,
			"items":                latestOrderBatch.Items,
			"batch":                latestOrderBatch,
		})
		return
	}

	batch := models.ProcurementBatch{
		UploaderID:  uploaderID,
		Period:      strings.TrimSpace(input.Period),
		OrderNumber: orderNumber,
		Status:      "待确认",
	}
	if err := database.DB.Create(&batch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create batch"})
		return
	}

	// 自动匹配逻辑：尝试用 CAS 号或名称模糊匹配品目字典
	createdCount := 0
	for i := range newItems {
		item := &newItems[i]
		item.BatchID = batch.ID
		item.MatchStatus = "未匹配"

		// 自动忽略非试剂及非化工类材料
		if item.MaterialCategory != "" && item.MaterialCategory != "化工" && !strings.Contains(item.ProductCategory, "试剂") {
			item.MatchStatus = "已忽略"
		} else {
			// 尝试 CAS 号精确匹配
			if item.CASNumber != "" {
				var catalog models.ReagentCatalog
				if err := database.DB.Where("cas_number = ?", item.CASNumber).First(&catalog).Error; err == nil {
					item.MatchedCatalogID = &catalog.ID
					item.MatchStatus = "自动匹配"

					// 进一步尝试匹配到最近的待采购 / 已接单的申购单
					var request models.ReagentRequest
					if err := database.DB.Where("reagent_catalog_id = ? AND status IN ?", catalog.ID, []string{"待采购", "已接单"}).
						Order("reagent_requests.created_at DESC").First(&request).Error; err == nil {
						item.MatchedRequestID = &request.ID
					}
				}
			}

			// 如果 CAS 匹配失败，尝试名称模糊匹配
			if item.MatchStatus == "未匹配" && item.ReagentName != "" {
				var catalog models.ReagentCatalog
				if err := database.DB.Where("name LIKE ? OR aliases LIKE ?",
					"%"+item.ReagentName+"%", "%"+item.ReagentName+"%").
					First(&catalog).Error; err == nil {
					item.MatchedCatalogID = &catalog.ID
					item.MatchStatus = "自动匹配"

					// 模糊匹配成功的，也尝试找一下申购单
					var request models.ReagentRequest
					if err := database.DB.Where("reagent_catalog_id = ? AND status IN ?", catalog.ID, []string{"待采购", "已接单"}).
						Order("reagent_requests.created_at DESC").First(&request).Error; err == nil {
						item.MatchedRequestID = &request.ID
					}
				}
			}
		}

		database.DB.Create(item)
		createdCount++
	}

	// 重新加载含 Items 的完整批次
	database.DB.Preload("Items").Preload("Uploader").First(&batch, batch.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message":              "导入完成",
		"skipped_count":        dedupStats.SkippedCount,
		"created_count":        createdCount,
		"confirmable":          createdCount > 0,
		"duplicate_in_file":    dedupStats.DuplicateInFile,
		"duplicate_in_system":  dedupStats.DuplicateInSystem,
		"duplicate_name_hints": dedupStats.DuplicateNameHints,
		"id":                   batch.ID,
		"items":                batch.Items,
		"batch":                batch,
	})
}

// GetProcurementBatches 获取所有采购批次列表
func GetProcurementBatches(c *gin.Context) {
	var batches []models.ProcurementBatch
	database.DB.Preload("Uploader").Preload("Items").Order("procurement_batches.created_at DESC").Find(&batches)
	c.JSON(http.StatusOK, batches)
}

// GetProcurementBatchItems 获取某批次的明细
func GetProcurementBatchItems(c *gin.Context) {
	batchID := c.Param("id")
	var items []models.ProcurementBatchItem
	database.DB.Where("batch_id = ?", batchID).Find(&items)
	c.JSON(http.StatusOK, items)
}

// UpdateProcurementBatchItem 手动修正明细行的匹配关系
func UpdateProcurementBatchItem(c *gin.Context) {
	itemID := c.Param("item_id")
	var item models.ProcurementBatchItem
	if err := database.DB.First(&item, itemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	var body struct {
		MatchedCatalogID *uint  `json:"matched_catalog_id"`
		MatchedRequestID *uint  `json:"matched_request_id"`
		MatchedUserID    *uint  `json:"matched_user_id"`
		CASNumber        string `json:"cas_number"`
		MatchStatus      string `json:"match_status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if body.MatchStatus == "已忽略" {
		item.MatchStatus = "已忽略"
		item.MatchedCatalogID = nil
		item.MatchedRequestID = nil
		item.MatchedUserID = nil
	} else {
		item.MatchedRequestID = body.MatchedRequestID
		item.MatchedCatalogID = body.MatchedCatalogID
		item.MatchedUserID = body.MatchedUserID

		// 若提供了申购单但没提供品目，自动回填品目
		if item.MatchedRequestID != nil && item.MatchedCatalogID == nil {
			var req models.ReagentRequest
			if err := database.DB.First(&req, *item.MatchedRequestID).Error; err == nil {
				item.MatchedCatalogID = &req.ReagentCatalogID
			}
		}

		if body.CASNumber != "" {
			item.CASNumber = body.CASNumber
		}
		item.MatchStatus = "手动匹配"
	}

	database.DB.Save(&item)
	c.JSON(http.StatusOK, item)
}

// ConfirmProcurementBatch 确认批次并触发到货赋码
func ConfirmProcurementBatch(c *gin.Context) {
	batchID := c.Param("id")
	var batch models.ProcurementBatch
	if err := database.DB.Preload("Items").First(&batch, batchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Batch not found"})
		return
	}

	if batch.Status != "待确认" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Batch already confirmed"})
		return
	}

	createdItems := 0
	for _, item := range batch.Items {
		if item.MatchedCatalogID == nil {
			continue // 跳过未匹配的行
		}

		// BPM-A/B 完全解耦：只累计数量，不修改申购单状态
		// 采购批次只记录关联关系，申购单的状态由 BPM-A 自己单独推进
		createdItems += item.Quantity
	}

	batch.Status = "已确认"
	database.DB.Save(&batch)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Batch confirmed, pending receiving",
		"items_created": createdItems,
	})
}

// ==========================================
// BPM-B Phase 3: 到货清点与赋码 (三段式解耦设计)
// ==========================================

// GetPendingReceives 获取所有已确认下单，但还未完全点货物理签收的批次明细
func GetPendingReceives(c *gin.Context) {
	var items []models.ProcurementBatchItem
	// 查找归属于"已确认"大状态批次中的明细，且明细本身的收货状态不是"已收货"，且必须关联了 catalog
	if err := database.DB.
		Joins("JOIN procurement_batches pb ON pb.id = procurement_batch_items.batch_id").
		Preload("Batch").
		Where("pb.status = ? AND procurement_batch_items.receive_status != ? AND procurement_batch_items.matched_catalog_id IS NOT NULL", "已确认", "已收货").
		Order("procurement_batch_items.created_at ASC").
		Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

// ReceiveBatchItem 采购人员在“收货工作台”对着单一点货，触发物理资产生成与原始单据闭环
func ReceiveBatchItem(c *gin.Context) {
	itemID := c.Param("itemId")
	var input struct {
		Quantity int `json:"quantity"` // 实收点验瓶数
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	var item models.ProcurementBatchItem
	if err := database.DB.Preload("Batch").First(&item, itemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Batch item not found"})
		return
	}

	if item.MatchedCatalogID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item has not been matched to a catalog"})
		return
	}

	if item.ReceiveStatus == "已收货" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item is already fully received"})
		return
	}

	remaining := item.Quantity - item.ReceivedQuantity
	if input.Quantity > remaining || input.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receive quantity"})
		return
	}

	// BPM-A/B 完全解耦：不再创建中间申购单，也不修改原始申购单状态
	// 仅做软关联：如果有匹配的申购单，就把 ReagentRequestID 写入到瓶资产上，方便复盘查询
	var reqID uint
	if item.MatchedRequestID != nil {
		reqID = *item.MatchedRequestID
	}

	createdItemsCount := 0
	createdUUIDs := make([]string, 0, input.Quantity)
	for q := 0; q < input.Quantity; q++ {
		reagentItem := models.ReagentItem{
			ReagentCatalogID: *item.MatchedCatalogID,
			Status:           "已到货", // 呆在暂存区
			Location:         ReagentStagingArea,
		}
		if reqID > 0 {
			reagentItem.ReagentRequestID = reqID
		}
		database.DB.Create(&reagentItem)
		createdItemsCount++
		createdUUIDs = append(createdUUIDs, reagentItem.UUID)
	}

	item.ReceivedQuantity += input.Quantity
	if item.ReceivedQuantity >= item.Quantity {
		item.ReceiveStatus = "已收货"
		// BPM-A/B 解耦：不再修改申购单状态，申购单由 BPM-A 自行流转
	} else {
		item.ReceiveStatus = "部分收货"
	}

	database.DB.Save(&item)

	c.JSON(http.StatusOK, gin.H{
		"message":            "Items received and marked for staging",
		"created_items":      createdItemsCount,
		"created_uuids":      createdUUIDs,
		"new_receive_status": item.ReceiveStatus,
	})
}

// ===================== 领用审批与双人双锁 =====================

// CreateDispenseRequest 研发提交领用申请
func CreateDispenseRequest(c *gin.Context) {
	var input struct {
		ReagentItemID string  `json:"reagent_item_id" binding:"required"`
		Amount        float64 `json:"amount" binding:"required"`
		Purpose       string  `json:"purpose"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requesterID := getUserIDOrDefault(c, 1)

	// 检查该试剂是否存在且在库
	var reagentItem models.ReagentItem
	if err := database.DB.Preload("ReagentCatalog").Where("uuid = ?", input.ReagentItemID).First(&reagentItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reagent item not found"})
		return
	}
	if reagentItem.Status != "在库" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅“在库”试剂可发起领用申请"})
		return
	}
	if !reagentItem.ReagentCatalog.IsControlled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "普通试剂不走领用申请，请直接登记消耗/耗尽"})
		return
	}
	if input.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "领取量必须大于 0"})
		return
	}
	if reagentItem.RemainingVolume > 0 && input.Amount > reagentItem.RemainingVolume {
		c.JSON(http.StatusBadRequest, gin.H{"error": "领取量超过当前余量"})
		return
	}

	dispenseReq := models.ReagentDispenseRequest{
		RequesterID:   requesterID,
		ReagentItemID: input.ReagentItemID,
		Amount:        input.Amount,
		Purpose:       input.Purpose,
		Status:        "待审批",
	}

	// 管控品需要设置双签超时
	expires := time.Now().Add(24 * time.Hour)
	dispenseReq.ExpiresAt = &expires

	if err := database.DB.Create(&dispenseReq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dispense request"})
		return
	}

	c.JSON(http.StatusCreated, dispenseReq)
}

// GetDispenseRequests 获取领用申请列表
func GetDispenseRequests(c *gin.Context) {
	var requests []models.ReagentDispenseRequest
	query := database.DB.Preload("Requester").Preload("ReagentItem").Preload("ReagentItem.ReagentCatalog").
		Preload("Leader").Preload("KeyHolderA").Preload("KeyHolderB")

	// 按角色过滤
	role := c.Query("role")
	userID := getUserIDOrDefault(c, 1)

	switch role {
	case "leader":
		query = query.Where("status IN ?", []string{"待审批", "已通过", "待双签", "已完成", "已驳回"})
	case "key_holder":
		query = query.Where("(key_holder_a_id = ? OR key_holder_b_id = ?) AND reagent_dispense_requests.status = ?", userID, userID, "待双签")
	default:
		query = query.Where("requester_id = ?", userID)
	}

	query.Order("reagent_dispense_requests.created_at DESC").Find(&requests)
	c.JSON(http.StatusOK, requests)
}

// LeaderApproveDispense 团队长审批领用申请
func LeaderApproveDispense(c *gin.Context) {
	id := c.Param("id")
	var dispReq models.ReagentDispenseRequest
	if err := database.DB.Preload("ReagentItem.ReagentCatalog").First(&dispReq, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dispense request not found"})
		return
	}

	if dispReq.Status != "待审批" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request is not pending approval"})
		return
	}

	var body struct {
		Approved     bool   `json:"approved"`
		RejectMsg    string `json:"reject_msg"`
		KeyHolderAID *uint  `json:"key_holder_a_id"`
		KeyHolderBID *uint  `json:"key_holder_b_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	leaderIDUint := getUserIDOrDefault(c, 101)
	now := time.Now()

	if !body.Approved {
		dispReq.Status = "已驳回"
		dispReq.LeaderID = &leaderIDUint
		dispReq.LeaderRejectMsg = body.RejectMsg
		database.DB.Save(&dispReq)
		c.JSON(http.StatusOK, gin.H{"message": "Request rejected", "status": "已驳回"})
		return
	}

	dispReq.LeaderID = &leaderIDUint
	dispReq.LeaderApprovedAt = &now

	// 判断是否为管控品 → 需要双签
	if dispReq.ReagentItem.ReagentCatalog.IsControlled {
		dispReq.Status = "待双签"
		dispReq.KeyHolderAID = body.KeyHolderAID
		dispReq.KeyHolderBID = body.KeyHolderBID
		expires := now.Add(24 * time.Hour)
		dispReq.ExpiresAt = &expires
		database.DB.Save(&dispReq)
		c.JSON(http.StatusOK, gin.H{"message": "Approved, awaiting dual key confirmation", "status": "待双签"})
		return
	}

	// 普通品直接通过
	dispReq.Status = "已通过"
	database.DB.Save(&dispReq)
	c.JSON(http.StatusOK, gin.H{"message": "Request approved", "status": "已通过"})
}

// KeyHolderConfirmDispense 钥匙持有人确认/驳回取用
func KeyHolderConfirmDispense(c *gin.Context) {
	id := c.Param("id")
	var dispReq models.ReagentDispenseRequest
	if err := database.DB.First(&dispReq, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Dispense request not found"})
		return
	}

	if dispReq.Status != "待双签" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request is not awaiting dual key confirmation"})
		return
	}

	// 检查是否超时
	if dispReq.ExpiresAt != nil && time.Now().After(*dispReq.ExpiresAt) {
		dispReq.Status = "已驳回"
		dispReq.KeyHolderRejectMsg = "双签已超时（24小时），自动驳回"
		database.DB.Save(&dispReq)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dual key confirmation expired"})
		return
	}

	var body struct {
		Confirmed bool   `json:"confirmed"`
		RejectMsg string `json:"reject_msg"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	holderIDUint := getUserIDOrDefault(c, 1)
	now := time.Now()

	if !body.Confirmed {
		dispReq.Status = "已驳回"
		dispReq.KeyHolderRejectMsg = body.RejectMsg
		database.DB.Save(&dispReq)
		c.JSON(http.StatusOK, gin.H{"message": "Dual key rejected", "status": "已驳回"})
		return
	}

	// 确认逻辑：判断当前操作者是 A 还是 B
	isA := dispReq.KeyHolderAID != nil && *dispReq.KeyHolderAID == holderIDUint
	isB := dispReq.KeyHolderBID != nil && *dispReq.KeyHolderBID == holderIDUint

	if !isA && !isB {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a key holder for this request"})
		return
	}

	if isA {
		dispReq.KeyHolderAConfirmedAt = &now
	}
	if isB {
		dispReq.KeyHolderBConfirmedAt = &now
	}

	// 检查双方是否都已确认
	if dispReq.KeyHolderAConfirmedAt != nil && dispReq.KeyHolderBConfirmedAt != nil {
		dispReq.Status = "已完成"
	}

	database.DB.Save(&dispReq)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Key holder confirmed",
		"status":      dispReq.Status,
		"a_confirmed": dispReq.KeyHolderAConfirmedAt != nil,
		"b_confirmed": dispReq.KeyHolderBConfirmedAt != nil,
	})
}
