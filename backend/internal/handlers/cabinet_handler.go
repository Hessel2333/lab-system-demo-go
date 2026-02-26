package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"
)

// GetReagentCabinets 获取试剂柜列表，支持按 dept_id 或 cabinet_type 筛选
func GetReagentCabinets(c *gin.Context) {
	var cabinets []models.ReagentCabinet
	tx := database.DB.Model(&models.ReagentCabinet{})

	if deptID := c.Query("dept_id"); deptID != "" {
		tx = tx.Where("department_id = ?", deptID)
	}
	if cabinetType := c.Query("type"); cabinetType != "" {
		tx = tx.Where("cabinet_type = ?", cabinetType)
	}

	if err := tx.Order("department_id ASC, name ASC").Find(&cabinets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cabinets)
}

// CreateReagentCabinet 创建新试剂柜
func CreateReagentCabinet(c *gin.Context) {
	var input struct {
		Name         string `json:"name" binding:"required"`
		CabinetType  string `json:"cabinet_type" binding:"required"`
		DepartmentID uint   `json:"department_id"`
		Location     string `json:"location"`
		Notes        string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cabinet := models.ReagentCabinet{
		Name:         input.Name,
		CabinetType:  input.CabinetType,
		DepartmentID: input.DepartmentID,
		Location:     input.Location,
		Notes:        input.Notes,
	}
	if err := database.DB.Create(&cabinet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cabinet)
}

// UpdateReagentCabinet 更新试剂柜信息
func UpdateReagentCabinet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var cabinet models.ReagentCabinet
	if err := database.DB.First(&cabinet, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "cabinet not found"})
		return
	}
	var input struct {
		Name         string `json:"name"`
		CabinetType  string `json:"cabinet_type"`
		DepartmentID *uint  `json:"department_id"`
		Location     string `json:"location"`
		Notes        string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.Name != "" {
		cabinet.Name = input.Name
	}
	if input.CabinetType != "" {
		cabinet.CabinetType = input.CabinetType
	}
	if input.DepartmentID != nil {
		cabinet.DepartmentID = *input.DepartmentID
	}
	if input.Location != "" {
		cabinet.Location = input.Location
	}
	cabinet.Notes = input.Notes

	if err := database.DB.Save(&cabinet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cabinet)
}

// DeleteReagentCabinet 删除试剂柜（软删除）
func DeleteReagentCabinet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// 检查是否有试剂实体仍在使用此柜
	var count int64
	database.DB.Model(&models.ReagentItem{}).Where("cabinet_id = ? AND reagent_items.status = ?", id, "在库").Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("该试剂柜当前仍有 %d 瓶在库试剂，无法删除", count)})
		return
	}

	if err := database.DB.Delete(&models.ReagentCabinet{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// SeedCabinets 初始化基础试剂柜点位，并批量为现有试剂条目分配柜子
func SeedCabinets(c *gin.Context) {
	// 先清理旧数据：重置所有 item 的 cabinet_id，然后硬删所有已有柜子
	database.DB.Model(&models.ReagentItem{}).Where("cabinet_id > 0").Update("cabinet_id", 0)
	database.DB.Unscoped().Where("1=1").Delete(&models.ReagentCabinet{})

	// 部门 ID：7=新材料研究院, 8=分析团队, 9=研发A, 10=研发B, 11=研发C, 12=条保部
	type cabinetSeed struct {
		Name         string
		CabinetType  string
		DepartmentID uint
		Location     string
		Notes        string
	}
	seeds := []cabinetSeed{
		// === 普通试剂柜（每个团队各一个）===
		{"分析团队-普通试剂柜A", "普通试剂柜", 8, "E309", ""},
		{"研发A组-普通试剂柜A", "普通试剂柜", 9, "F103", ""},
		{"研发B组-普通试剂柜A", "普通试剂柜", 10, "F309", ""},
		{"研发C组-普通试剂柜A", "普通试剂柜", 11, "G101", ""},
		{"新材料院-普通试剂柜A", "普通试剂柜", 7, "E309", ""},
		{"条保部-普通试剂柜A", "普通试剂柜", 12, "E309", ""},

		// === 易制毒制爆试剂柜（统一在 F311，公共共享，共 4 个）===
		{"F311-管控柜1号", "易制毒制爆试剂柜", 0, "F311", "双人双锁"},
		{"F311-管控柜2号", "易制毒制爆试剂柜", 0, "F311", "双人双锁"},
		{"F311-管控柜3号", "易制毒制爆试剂柜", 0, "F311", "双人双锁"},
		{"F311-管控柜4号", "易制毒制爆试剂柜", 0, "F311", "双人双锁"},
	}

	// 创建柜子
	cabinetMap := map[string]*models.ReagentCabinet{} // deptID_type → cabinet（普通柜按团队）
	var controlledCabinets []*models.ReagentCabinet   // 管控柜列表（轮流分配）
	for _, s := range seeds {
		cab := models.ReagentCabinet{
			Name:         s.Name,
			CabinetType:  s.CabinetType,
			DepartmentID: s.DepartmentID,
			Location:     s.Location,
			Notes:        s.Notes,
		}
		if err := database.DB.Create(&cab).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建柜子失败: " + err.Error()})
			return
		}
		if s.CabinetType == "普通试剂柜" {
			key := fmt.Sprintf("%d_普通试剂柜", s.DepartmentID)
			cabinetMap[key] = &cab
		} else {
			controlledCabinets = append(controlledCabinets, &cab)
		}
	}

	// 批量为现有 ReagentItem 分配柜子
	var items []models.ReagentItem
	database.DB.Preload("ReagentCatalog").
		Preload("ReagentRequest").
		Preload("ReagentRequest.Requestor").
		Find(&items)

	updatedCount := 0
	controlledIdx := 0
	for i := range items {
		item := &items[i]
		isControlled := item.ReagentCatalog.IsControlled

		var targetCabinetID uint
		if isControlled {
			// 管控品轮流分配到 4 个管控柜
			if len(controlledCabinets) > 0 {
				targetCabinetID = controlledCabinets[controlledIdx%len(controlledCabinets)].ID
				controlledIdx++
			}
		} else {
			// 普通品按团队分配
			deptID := item.ReagentRequest.Requestor.DepartmentID
			key := fmt.Sprintf("%d_普通试剂柜", deptID)
			if cab, ok := cabinetMap[key]; ok {
				targetCabinetID = cab.ID
			}
		}

		if targetCabinetID > 0 {
			database.DB.Model(item).Update("cabinet_id", targetCabinetID)
			updatedCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "试剂柜初始化成功",
		"cabinets":      len(seeds),
		"items_updated": updatedCount,
	})
}
