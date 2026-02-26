package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedReagents creates rich initial demo data for the reagent module
func SeedReagents(c *gin.Context) {
	tx := database.DB.Begin()

	// 0. Clear old data to prevent duplicates
	tx.Exec("DELETE FROM reagent_logs")
	tx.Exec("DELETE FROM reagent_items")
	tx.Exec("DELETE FROM procurement_batch_items")
	tx.Exec("DELETE FROM procurement_batches")
	tx.Exec("DELETE FROM reagent_requests")
	// tx.Exec("DELETE FROM reagent_catalogs") // Let's keep catalogs or recreate them if not exists to avoid violating foreign keys from other places if any

	// 0.1 Ensure Organization exists (Fixed IDs)
	if err := InternalSeedOrganization(tx); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed organization: " + err.Error()})
		return
	}

	if err := InternalSeedUsers(tx); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed users: " + err.Error()})
		return
	}

	// Fetch some default users (from Dept 8 - 分析团队) to create requests
	var researcher models.User
	var researcher2 models.User
	tx.Where("username = ?", "zhangming").First(&researcher)
	tx.Where("username = ?", "lihua").First(&researcher2)

	// 1. Create Diverse Catalogs
	catalogs := []models.ReagentCatalog{
		// Organic Solvents
		{CASNumber: "64-17-5", Name: "无水乙醇", Alias: "Ethanol", Formula: "C2H6O", Category: "有机溶剂", IsControlled: false, Storage: "E309", AlertThreshold: 10, Unit: "500ml",
			ChemicalLabels: `["危险化学品"]`, Aliases: "乙醇,酒精,Ethanol,EtOH", StorageCondition: "阴凉干燥处,远离火源", PhysicalState: "液体"},
		{CASNumber: "67-64-1", Name: "丙酮", Alias: "Acetone", Formula: "C3H6O", Category: "有机溶剂", IsControlled: true, Storage: "F103", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "二甲基酮,Acetone", StorageCondition: "阴凉通风处,远离火源和氧化剂", PhysicalState: "液体"},
		{CASNumber: "67-56-1", Name: "甲醇", Alias: "Methanol", Formula: "CH4O", Category: "有机溶剂", IsControlled: true, Storage: "E309", AlertThreshold: 8, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "木精,木醇,Methanol,MeOH", StorageCondition: "阴凉通风处,远离火源,密封保存", PhysicalState: "液体"},

		// Acids and Bases
		{CASNumber: "7647-01-0", Name: "盐酸", Alias: "Hydrochloric Acid", Formula: "HCl", Category: "无机酸", IsControlled: true, Storage: "E307", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "氢氯酸,Hydrochloric Acid,HCl", StorageCondition: "阴凉通风处,耐酸柜存放", PhysicalState: "液体"},
		{CASNumber: "7664-93-9", Name: "硫酸 (98%)", Alias: "Sulfuric Acid", Formula: "H2SO4", Category: "无机酸", IsControlled: true, Storage: "E307", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制爆化学品"]`, Aliases: "硫酸,矾油,Sulfuric Acid,H2SO4", StorageCondition: "阴凉干燥处,耐酸专柜", PhysicalState: "液体"},

		// Bio & Analytical Standard
		{CASNumber: "7647-14-5", Name: "氯化钠", Alias: "Sodium Chloride", Formula: "NaCl", Category: "生化试剂", IsControlled: false, Storage: "E309", AlertThreshold: 5, Unit: "500g",
			ChemicalLabels: `["普通化学品"]`, Aliases: "食盐,NaCl,Sodium Chloride", StorageCondition: "常温干燥处", PhysicalState: "固体"},

		// 高纯水 & 气体
		{CASNumber: "7732-18-5", Name: "超纯水", Alias: "Ultrapure Water", Formula: "H2O", Category: "通用试剂", IsControlled: false, Storage: "E309", AlertThreshold: 20, Unit: "1L",
			ChemicalLabels: `["普通化学品"]`, Aliases: "去离子水,MilliQ水,Deionized Water", StorageCondition: "常温密封保存", PhysicalState: "液体"},
	}

	for i := range catalogs {
		if err := tx.Where(models.ReagentCatalog{CASNumber: catalogs[i].CASNumber}).FirstOrCreate(&catalogs[i]).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed catalogs: " + err.Error()})
			return
		}
	}

	// 2. Create Reagent Requests (Historical and Current)
	// 历史完成的数据 (BPM-A 闭环 -> 已接单)
	historyRequests := []models.ReagentRequest{
		{RequestorID: researcher.ID, ReagentCatalogID: catalogs[0].ID, Quantity: 10, Status: "已接单", Remarks: "上个月常规补充", OrderReference: "YPK-HIST-001"},
		{RequestorID: researcher2.ID, ReagentCatalogID: catalogs[1].ID, Quantity: 5, Status: "已接单", Remarks: "项目攻坚消耗大", OrderReference: "YPK-HIST-002"},
		{RequestorID: researcher.ID, ReagentCatalogID: catalogs[3].ID, Quantity: 3, Status: "已接单", Remarks: "滴定分析用", OrderReference: "YPK-HIST-003"},
	}

	for i := range historyRequests {
		if err := tx.Create(&historyRequests[i]).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed history requests"})
			return
		}

		catCapacity := 500.0
		// Generate ReagentItems for these closed requests
		for j := 0; j < historyRequests[i].Quantity; j++ {
			statusObj := "在库"
			remaining := float64(10 + rand.Intn(490))             // Random remaining
			location := historyRequests[i].ReagentCatalog.Storage // default location

			// Make some exhausted
			if rand.Intn(100) < 20 {
				statusObj = "已耗尽"
				remaining = 0
			}

			item := models.ReagentItem{
				UUID:             uuid.New().String(),
				ReagentRequestID: historyRequests[i].ID,
				ReagentCatalogID: historyRequests[i].ReagentCatalogID,
				Status:           statusObj,
				Location:         location,
				Capacity:         catCapacity,
				RemainingVolume:  remaining,
				BatchNumber:      fmt.Sprintf("BATCH-2026-%d", rand.Intn(1000)),
				ExpiryDate:       time.Now().AddDate(1, 0, 0),
			}
			if err := tx.Create(&item).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed items"})
				return
			}

			// Initial Log
			tx.Create(&models.ReagentLog{
				ReagentItemID: item.UUID,
				UserID:        researcher.ID,
				Action:        "入库登记",
				Quantity:      catCapacity,
				Remarks:       "历史订单初始入库",
			})

			// Exhausted Log
			if item.Status == "已耗尽" {
				tx.Create(&models.ReagentLog{
					ReagentItemID: item.UUID,
					UserID:        researcher.ID,
					Action:        "空瓶核销",
					Quantity:      0,
					Remarks:       "由于使用完毕进行核销",
				})
			}
		}
	}

	// 3. Create Pending Requests (These are the targets for our Excel match!)
	pendingRequests := []models.ReagentRequest{
		{RequestorID: researcher.ID, ReagentCatalogID: catalogs[2].ID, Quantity: 5, Status: "待采购", Remarks: "急用，提取溶剂快用完了"}, // 甲醇
		{RequestorID: researcher2.ID, ReagentCatalogID: catalogs[4].ID, Quantity: 2, Status: "待审批", Remarks: "酸消解实验即将开始"},  // 硫酸
		{RequestorID: researcher.ID, ReagentCatalogID: catalogs[6].ID, Quantity: 10, Status: "待采购", Remarks: "清洗仪纯水补充"},    // 超纯水
	}

	for i := range pendingRequests {
		if err := tx.Create(&pendingRequests[i]).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed pending requests"})
			return
		}
	}

	// 4. Create Procurement Batches and Items (For "待点验" testing)
	testBatch := models.ProcurementBatch{
		OrderNumber: fmt.Sprintf("PO-TEST-%d", rand.Intn(9000)),
		Status:      "已确认",
	}
	tx.Create(&testBatch)

	tx.Create(&models.ProcurementBatchItem{
		BatchID:          testBatch.ID,
		ReagentName:      "无水乙醇",
		CASNumber:        "64-17-5",
		Quantity:         5,
		Unit:             "瓶",
		Supplier:         "国药试剂",
		MatchedCatalogID: &catalogs[0].ID,
		ReceiveStatus:    "待收货",
		ReceivedQuantity: 0,
	})

	tx.Create(&models.ProcurementBatchItem{
		BatchID:          testBatch.ID,
		ReagentName:      "甲醇",
		CASNumber:        "67-56-1",
		Quantity:         10,
		Unit:             "瓶",
		Supplier:         "阿拉丁试剂",
		MatchedCatalogID: &catalogs[2].ID,
		ReceiveStatus:    "部分收货",
		ReceivedQuantity: 3,
	})

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Reagent demo data seeded successfully! Remember to run /api/debug/seed_cabinets to attach physical cabinets."})
}

// SeedTeamInventory 为各研发团队注入在库试剂 seed 数据（真实写入数据库，非 Mock）
func SeedTeamInventory(c *gin.Context) {
	tx := database.DB.Begin()

	// 清理团队相关的模拟数据 (可选)
	tx.Exec("DELETE FROM users WHERE username LIKE 'leader_%'")

	// 各团队负责人用户 ID（挂载到我们生成的真实用户身上）
	type teamSeed struct {
		userID    uint
		teamLabel string
		deptID    uint
	}
	teams := []teamSeed{
		{userID: 101, teamLabel: "分析团队", deptID: 8},   // 张明
		{userID: 201, teamLabel: "研发团队A", deptID: 9},  // 赵强
		{userID: 301, teamLabel: "研发团队B", deptID: 10}, // 周杰
		{userID: 401, teamLabel: "研发团队C", deptID: 11}, // 郑六
	}

	// 各团队的试剂清单（CAS 号 → 数量 → 库位）
	type reagentEntry struct {
		casNumber string
		quantity  int
		location  string
		remarks   string
	}
	teamReagents := map[uint][]reagentEntry{
		101: {
			{"64-17-5", 3, "E309", "分析用无水乙醇"},
			{"67-64-1", 2, "F103", "色谱流动相丙酮"},
		},
		201: {
			{"67-56-1", 3, "E309", "提取用甲醇"},
			{"108-88-3", 2, "F103", "合成用甲苯"},
		},
		301: {
			{"64-17-5", 4, "E309", "洗涤用无水乙醇"},
			{"7647-14-5", 3, "E309", "生理盐水氯化钠"},
		},
		401: {
			{"67-64-1", 3, "F103", "反应溶剂丙酮"},
			{"6381-92-6", 2, "F309", "螯合剂 EDTA 二钠"},
		},
	}

	for _, t := range teams {
		entries, ok := teamReagents[t.userID]
		if !ok {
			continue
		}
		for _, entry := range entries {
			// 找到 catalog
			var cat models.ReagentCatalog
			if err := tx.Where("cas_number = ?", entry.casNumber).First(&cat).Error; err != nil {
				continue // 如果catalog尚未seed，直接跳过
			}

			// 创建团队特属申购单
			req := models.ReagentRequest{
				RequestorID:      t.userID,
				ReagentCatalogID: cat.ID,
				Quantity:         entry.quantity,
				Status:           "已接单",
				Remarks:          entry.remarks,
				OrderReference:   fmt.Sprintf("YPK-TEAM-%d", rand.Intn(9000)+1000),
			}
			if err := tx.Create(&req).Error; err != nil {
				continue
			}

			// 为每瓶创建 ReagentItem
			for i := 0; i < entry.quantity; i++ {
				capacity := 500.0
				remaining := capacity
				if i == entry.quantity-1 && entry.quantity > 1 {
					remaining = float64(100 + rand.Intn(300)) // 最后一瓶是被领用过的
				}

				item := models.ReagentItem{
					UUID:             uuid.New().String(),
					ReagentRequestID: req.ID,
					ReagentCatalogID: cat.ID,
					Status:           "在库",
					Location:         entry.location,
					Capacity:         capacity,
					RemainingVolume:  remaining,
					BatchNumber:      fmt.Sprintf("BATCH-%d-%d", time.Now().Year(), rand.Intn(9000)+1000),
					ExpiryDate:       time.Now().AddDate(1, rand.Intn(12), 0),
				}

				tx.Create(&item)

				// 制造完整的声明周期日志轴
				// 1. 入库登记
				tx.Create(&models.ReagentLog{
					ReagentItemID: item.UUID,
					UserID:        t.userID,
					Action:        "入库登记",
					Quantity:      capacity,
					Remarks:       "总库调拨分配入库",
				})

				// 2. 扫码领用 (只有用掉部分容量的那一瓶)
				if remaining < capacity {
					tx.Create(&models.ReagentLog{
						ReagentItemID: item.UUID,
						UserID:        t.userID,
						Action:        "领用消耗",
						Quantity:      capacity - remaining,
						Remarks:       "日常实验消耗记录",
					})
				}
			}
		}
	}

	// === 添加一些特殊的“已到货(待入库)”测试条目（用于验证待入库跟进列表以及超时警告）===
	// 制造一条今天刚点验的 (给当前默认登录用户 ID=1 Admin)：
	reqRecent := models.ReagentRequest{
		RequestorID:      1, // Admin (Current Frontend User)
		ReagentCatalogID: 1, // 无水乙醇
		Quantity:         1,
		Status:           "已接单",
		Remarks:          "测试刚点验的数据",
		OrderReference:   "YPK-TEST-NEW",
	}
	tx.Create(&reqRecent)
	tx.Create(&models.ReagentItem{
		UUID:             uuid.New().String(),
		ReagentRequestID: reqRecent.ID,
		ReagentCatalogID: reqRecent.ReagentCatalogID,
		Status:           "已到货",
		Location:         ReagentStagingArea,
		Capacity:         500.0,
		RemainingVolume:  500.0,
		BatchNumber:      fmt.Sprintf("BATCH-TEST-%d", rand.Intn(9000)),
		ExpiryDate:       time.Now().AddDate(1, 0, 0),
		CreatedAt:        time.Now(), // 刚刚创建
	})

	// 制造一条超过24小时未入库的：
	reqOverdue := models.ReagentRequest{
		RequestorID:      1, // Admin (Current Frontend User)
		ReagentCatalogID: 2, // 丙酮
		Quantity:         2,
		Status:           "已接单",
		Remarks:          "测试超时未领回数据",
		OrderReference:   "YPK-TEST-OLD",
	}
	tx.Create(&reqOverdue)
	tx.Create(&models.ReagentItem{
		UUID:             uuid.New().String(),
		ReagentRequestID: reqOverdue.ID,
		ReagentCatalogID: reqOverdue.ReagentCatalogID,
		Status:           "已到货",
		Location:         "采购部实验桌收发台",
		Capacity:         500.0,
		RemainingVolume:  500.0,
		BatchNumber:      fmt.Sprintf("BATCH-TEST-%d", rand.Intn(9000)),
		ExpiryDate:       time.Now().AddDate(1, 0, 0),
		CreatedAt:        time.Now().Add(-26 * time.Hour), // 26小时前创建的
	})

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Team inventory seed data with complete lifecycle created successfully"})
}

// SeedOrganization 手动触发组织架构初始化
func SeedOrganization(c *gin.Context) {
	tx := database.DB.Begin()
	if err := InternalSeedOrganization(tx); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Organization seed data created successfully (Fixed IDs 7-12)"})
}

// InternalSeedOrganization 内部组织架构生成逻辑，确保 ID 锚点一致
func InternalSeedOrganization(tx *gorm.DB) error {
	// 部门 ID：7=新材料研究院, 8=分析团队, 9=研发A, 10=研发B, 11=研发C, 12=条保部
	depts := []models.Department{
		{Model: gorm.Model{ID: 7}, Name: "新材料研究院", Type: "institute"},
	}

	// 先创建根节点
	for _, d := range depts {
		if err := tx.Where(models.Department{Model: gorm.Model{ID: d.ID}}).FirstOrCreate(&d).Error; err != nil {
			return err
		}
	}

	// 创建子节点
	parentID := uint(7)
	subDepts := []models.Department{
		{Model: gorm.Model{ID: 8}, Name: "分析团队", Type: "team", ParentID: &parentID},
		{Model: gorm.Model{ID: 9}, Name: "研发团队A", Type: "team", ParentID: &parentID},
		{Model: gorm.Model{ID: 10}, Name: "研发团队B", Type: "team", ParentID: &parentID},
		{Model: gorm.Model{ID: 11}, Name: "研发团队C", Type: "team", ParentID: &parentID},
		{Model: gorm.Model{ID: 12}, Name: "条件保障部", Type: "dept", ParentID: &parentID},
	}

	for _, d := range subDepts {
		if err := tx.Where(models.Department{Model: gorm.Model{ID: d.ID}}).FirstOrCreate(&d).Error; err != nil {
			return err
		}
	}
	return nil
}

// InternalSeedUsers 生成所有固定的标准团队成员与领导层用户
func InternalSeedUsers(tx *gorm.DB) error {
	users := []models.User{
		{Model: gorm.Model{ID: 1}, Username: "admin", RealName: "系统管理员", DepartmentID: 7, Role: "admin"},
		{Model: gorm.Model{ID: 2}, Username: "caigou1", RealName: "刘采购", DepartmentID: 12, Role: "procurement"},
		{Model: gorm.Model{ID: 3}, Username: "caigou2", RealName: "王采购", DepartmentID: 12, Role: "procurement"},

		// 分析团队 (Dept 8)
		{Model: gorm.Model{ID: 101}, Username: "zhangming", RealName: "张明", DepartmentID: 8, Role: "team_leader"},
		{Model: gorm.Model{ID: 102}, Username: "lihua", RealName: "李华", DepartmentID: 8, Role: "researcher"},
		{Model: gorm.Model{ID: 103}, Username: "wangwei", RealName: "王伟", DepartmentID: 8, Role: "researcher"},

		// 研发A组 (Dept 9)
		{Model: gorm.Model{ID: 201}, Username: "zhaoqiang", RealName: "赵强", DepartmentID: 9, Role: "team_leader"},
		{Model: gorm.Model{ID: 202}, Username: "sunjing", RealName: "孙静", DepartmentID: 9, Role: "researcher"},

		// 研发B组 (Dept 10)
		{Model: gorm.Model{ID: 301}, Username: "zhoujie", RealName: "周杰", DepartmentID: 10, Role: "team_leader"},
		{Model: gorm.Model{ID: 302}, Username: "wufan", RealName: "吴凡", DepartmentID: 10, Role: "researcher"},

		// 研发C组 (Dept 11)
		{Model: gorm.Model{ID: 401}, Username: "zhengliu", RealName: "郑六", DepartmentID: 11, Role: "team_leader"},
	}

	for _, expectedUser := range users {
		var count int64
		tx.Unscoped().Model(&models.User{}).Where("id = ?", expectedUser.ID).Count(&count)
		if count > 0 {
			tx.Unscoped().Model(&models.User{}).Where("id = ?", expectedUser.ID).Updates(map[string]interface{}{
				"username":      expectedUser.Username,
				"real_name":     expectedUser.RealName,
				"department_id": expectedUser.DepartmentID,
				"role":          expectedUser.Role,
				"deleted_at":    nil,
			})
		} else {
			if err := tx.Create(&expectedUser).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
