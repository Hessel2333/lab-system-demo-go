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
	tx.Exec("DELETE FROM reagent_requests")
	tx.Exec("DELETE FROM reagent_catalogs")

	// 1. Create Diverse Catalogs
	catalogs := []models.ReagentCatalog{
		// Organic Solvents
		{CASNumber: "64-17-5", Name: "无水乙醇", Alias: "Ethanol", Formula: "C2H6O", Category: "有机溶剂", IsControlled: false, Storage: "E309", AlertThreshold: 10, Unit: "500ml",
			ChemicalLabels: `["危险化学品"]`, Aliases: "乙醇,酒精,Ethanol,EtOH", StorageCondition: "阴凉干燥处,远离火源", PhysicalState: "液体"},
		{CASNumber: "67-64-1", Name: "丙酮", Alias: "Acetone", Formula: "C3H6O", Category: "有机溶剂", IsControlled: true, Storage: "F103", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "二甲基酮,Acetone", StorageCondition: "阴凉通风处,远离火源和氧化剂", PhysicalState: "液体"},
		{CASNumber: "67-56-1", Name: "甲醇", Alias: "Methanol", Formula: "CH4O", Category: "有机溶剂", IsControlled: true, Storage: "E309", AlertThreshold: 8, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "木精,木醇,Methanol,MeOH", StorageCondition: "阴凉通风处,远离火源,密封保存", PhysicalState: "液体"},
		{CASNumber: "108-88-3", Name: "甲苯", Alias: "Toluene", Formula: "C7H8", Category: "有机溶剂", IsControlled: true, Storage: "F103", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "甲基苯,Toluene", StorageCondition: "阴凉通风处,远离火源和氧化剂", PhysicalState: "液体"},

		// Acids and Bases
		{CASNumber: "7647-01-0", Name: "盐酸", Alias: "Hydrochloric Acid", Formula: "HCl", Category: "无机酸", IsControlled: true, Storage: "E307", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "氢氯酸,Hydrochloric Acid,HCl", StorageCondition: "阴凉通风处,耐酸柜存放", PhysicalState: "液体"},
		{CASNumber: "7664-93-9", Name: "硫酸", Alias: "Sulfuric Acid", Formula: "H2SO4", Category: "无机酸", IsControlled: true, Storage: "E307", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制爆化学品"]`, Aliases: "矾油,Sulfuric Acid,H2SO4", StorageCondition: "阴凉干燥处,耐酸专柜", PhysicalState: "液体"},
		{CASNumber: "1310-73-2", Name: "氢氧化钠", Alias: "Sodium Hydroxide", Formula: "NaOH", Category: "无机碱", IsControlled: false, Storage: "F309", AlertThreshold: 2, Unit: "500g",
			ChemicalLabels: `["危险化学品"]`, Aliases: "烧碱,火碱,苛性钠,NaOH", StorageCondition: "密封干燥保存,防潮", PhysicalState: "固体"},
		{CASNumber: "1336-21-6", Name: "氨水", Alias: "Ammonia Solution", Formula: "NH3·H2O", Category: "无机碱", IsControlled: false, Storage: "F309", AlertThreshold: 3, Unit: "500ml",
			ChemicalLabels: `["危险化学品"]`, Aliases: "氢氧化铵,Ammonia Solution", StorageCondition: "阴凉通风处,密封保存", PhysicalState: "液体"},

		// Bio & Analytical Standard
		{CASNumber: "7647-14-5", Name: "氯化钠", Alias: "Sodium Chloride", Formula: "NaCl", Category: "生化试剂", IsControlled: false, Storage: "E309", AlertThreshold: 5, Unit: "500g",
			ChemicalLabels: `["普通化学品"]`, Aliases: "食盐,NaCl,Sodium Chloride", StorageCondition: "常温干燥处", PhysicalState: "固体"},
		{CASNumber: "50-99-7", Name: "D-葡萄糖", Alias: "D-Glucose", Formula: "C6H12O6", Category: "生化试剂", IsControlled: false, Storage: "E309", AlertThreshold: 5, Unit: "500g",
			ChemicalLabels: `["普通化学品"]`, Aliases: "葡萄糖,Glucose,Dextrose", StorageCondition: "常温干燥处,密封保存", PhysicalState: "固体"},
		{CASNumber: "9002-93-1", Name: "Triton X-100", Alias: "Triton X-100", Formula: "C14H22O(C2H4O)n", Category: "表面活性剂", IsControlled: false, Storage: "F103", AlertThreshold: 2, Unit: "100ml",
			ChemicalLabels: `["普通化学品"]`, Aliases: "聚乙二醇辛基苯基醚,TX-100", StorageCondition: "常温避光保存", PhysicalState: "液体"},
		{CASNumber: "6381-92-6", Name: "EDTA二钠", Alias: "EDTA-Na2", Formula: "C10H14N2Na2O8", Category: "络合剂", IsControlled: false, Storage: "F309", AlertThreshold: 2, Unit: "250g",
			ChemicalLabels: `["普通化学品"]`, Aliases: "乙二胺四乙酸二钠,EDTA-2Na,EDTA Disodium Salt", StorageCondition: "常温干燥处", PhysicalState: "固体"},

		// ---- 扩展品目 ----

		// 色谱级溶剂
		{CASNumber: "75-05-8", Name: "乙腈", Alias: "Acetonitrile", Formula: "C2H3N", Category: "有机溶剂", IsControlled: true, Storage: "F103", AlertThreshold: 5, Unit: "4L",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "甲基氰,Acetonitrile,ACN", StorageCondition: "阴凉通风处,远离火源,密封避光", PhysicalState: "液体"},
		{CASNumber: "75-09-2", Name: "二氯甲烷", Alias: "Dichloromethane", Formula: "CH2Cl2", Category: "有机溶剂", IsControlled: true, Storage: "F103", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "DCM,Methylene Chloride", StorageCondition: "阴凉通风处,远离火源和氧化剂", PhysicalState: "液体"},
		{CASNumber: "67-63-0", Name: "异丙醇", Alias: "Isopropanol", Formula: "C3H8O", Category: "有机溶剂", IsControlled: false, Storage: "E309", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品"]`, Aliases: "IPA,2-丙醇,Isopropyl Alcohol", StorageCondition: "阴凉干燥处,远离火源", PhysicalState: "液体"},
		{CASNumber: "141-78-6", Name: "乙酸乙酯", Alias: "Ethyl Acetate", Formula: "C4H8O2", Category: "有机溶剂", IsControlled: false, Storage: "F103", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品"]`, Aliases: "醋酸乙酯,EA,Ethyl Acetate", StorageCondition: "阴凉通风处,远离火源", PhysicalState: "液体"},

		// 无机试剂
		{CASNumber: "7697-37-2", Name: "硝酸", Alias: "Nitric Acid", Formula: "HNO3", Category: "无机酸", IsControlled: true, Storage: "E307", AlertThreshold: 3, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制爆化学品"]`, Aliases: "Nitric Acid,HNO3", StorageCondition: "阴凉通风处,耐酸柜,远离有机物", PhysicalState: "液体"},
		{CASNumber: "7664-38-2", Name: "磷酸", Alias: "Phosphoric Acid", Formula: "H3PO4", Category: "无机酸", IsControlled: false, Storage: "E307", AlertThreshold: 3, Unit: "500ml",
			ChemicalLabels: `["危险化学品"]`, Aliases: "正磷酸,Phosphoric Acid,H3PO4", StorageCondition: "阴凉处,耐酸柜存放", PhysicalState: "液体"},
		{CASNumber: "7681-49-4", Name: "氟化钠", Alias: "Sodium Fluoride", Formula: "NaF", Category: "无机盐", IsControlled: true, Storage: "F309", AlertThreshold: 2, Unit: "100g",
			ChemicalLabels: `["危险化学品","剧毒化学品"]`, Aliases: "NaF,Sodium Fluoride", StorageCondition: "密封干燥保存,专柜存放,双人双锁", PhysicalState: "固体"},
		{CASNumber: "7440-38-2", Name: "砷", Alias: "Arsenic", Formula: "As", Category: "无机单质", IsControlled: true, Storage: "G101", AlertThreshold: 1, Unit: "25g",
			ChemicalLabels: `["危险化学品","剧毒化学品","限制化学品"]`, Aliases: "砒,Arsenic", StorageCondition: "双人双锁,专柜隔离存放", PhysicalState: "固体"},

		// 生化试剂
		{CASNumber: "69-52-3", Name: "氨苄青霉素钠", Alias: "Ampicillin Sodium", Formula: "C16H18N3NaO4S", Category: "抗生素", IsControlled: false, Storage: "F309", AlertThreshold: 3, Unit: "25g",
			ChemicalLabels: `["普通化学品"]`, Aliases: "Amp,Ampicillin Sodium Salt", StorageCondition: "-20°C冷冻保存", PhysicalState: "固体"},
		{CASNumber: "9005-64-5", Name: "Tween 20", Alias: "Tween 20", Formula: "C58H114O26", Category: "表面活性剂", IsControlled: false, Storage: "F103", AlertThreshold: 2, Unit: "500ml",
			ChemicalLabels: `["普通化学品"]`, Aliases: "聚山梨酯20,Polysorbate 20,吐温20", StorageCondition: "常温避光保存", PhysicalState: "液体"},
		{CASNumber: "77-86-1", Name: "Tris碱", Alias: "Tris Base", Formula: "C4H11NO3", Category: "缓冲剂", IsControlled: false, Storage: "E309", AlertThreshold: 3, Unit: "500g",
			ChemicalLabels: `["普通化学品"]`, Aliases: "三羟甲基氨基甲烷,Tris,THAM", StorageCondition: "常温干燥处", PhysicalState: "固体"},
		{CASNumber: "7365-45-9", Name: "HEPES", Alias: "HEPES", Formula: "C8H18N2O4S", Category: "缓冲剂", IsControlled: false, Storage: "E309", AlertThreshold: 2, Unit: "100g",
			ChemicalLabels: `["普通化学品"]`, Aliases: "4-羟乙基哌嗪乙磺酸,HEPES Buffer", StorageCondition: "常温干燥处,密封保存", PhysicalState: "固体"},

		// 指示剂与色谱标准品
		{CASNumber: "60-00-4", Name: "EDTA", Alias: "EDTA", Formula: "C10H16N2O8", Category: "络合剂", IsControlled: false, Storage: "F309", AlertThreshold: 2, Unit: "250g",
			ChemicalLabels: `["普通化学品"]`, Aliases: "乙二胺四乙酸,Ethylenediaminetetraacetic Acid", StorageCondition: "常温干燥处", PhysicalState: "固体"},
		{CASNumber: "7553-56-2", Name: "碘", Alias: "Iodine", Formula: "I2", Category: "无机单质", IsControlled: true, Storage: "E307", AlertThreshold: 2, Unit: "100g",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "碘单质,Iodine,I2", StorageCondition: "阴凉避光处,密封保存", PhysicalState: "固体"},

		// 高纯水 & 气体
		{CASNumber: "7732-18-5", Name: "超纯水", Alias: "Ultrapure Water", Formula: "H2O", Category: "通用试剂", IsControlled: false, Storage: "E309", AlertThreshold: 20, Unit: "1L",
			ChemicalLabels: `["普通化学品"]`, Aliases: "去离子水,MilliQ水,Deionized Water", StorageCondition: "常温密封保存", PhysicalState: "液体"},
		{CASNumber: "64-19-7", Name: "冰乙酸", Alias: "Glacial Acetic Acid", Formula: "CH3COOH", Category: "有机酸", IsControlled: false, Storage: "E307", AlertThreshold: 3, Unit: "500ml",
			ChemicalLabels: `["危险化学品"]`, Aliases: "醋酸,冰醋酸,Acetic Acid,GAA", StorageCondition: "阴凉通风处,远离氧化剂", PhysicalState: "液体"},
		{CASNumber: "110-54-3", Name: "正己烷", Alias: "n-Hexane", Formula: "C6H14", Category: "有机溶剂", IsControlled: true, Storage: "F103", AlertThreshold: 5, Unit: "500ml",
			ChemicalLabels: `["危险化学品","易制毒化学品"]`, Aliases: "己烷,Hexane,n-Hexane", StorageCondition: "阴凉通风处,远离火源和氧化剂", PhysicalState: "液体"},
		{CASNumber: "7778-50-9", Name: "重铬酸钾", Alias: "Potassium Dichromate", Formula: "K2Cr2O7", Category: "氧化剂", IsControlled: true, Storage: "G101", AlertThreshold: 2, Unit: "500g",
			ChemicalLabels: `["危险化学品","剧毒化学品","限制化学品"]`, Aliases: "红矾钾,K2Cr2O7,Potassium Dichromate", StorageCondition: "密封干燥保存,专柜隔离,双人双锁", PhysicalState: "固体"},
	}

	for i := range catalogs {
		if err := tx.Where(models.ReagentCatalog{CASNumber: catalogs[i].CASNumber}).FirstOrCreate(&catalogs[i]).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed catalogs: " + err.Error()})
			return
		}
	}

	// Fetch User (ensure at least user id=1 exists, mock creation if missing)
	var user models.User
	if err := tx.Where(models.User{Username: "admin"}).FirstOrCreate(&user, models.User{
		Model:        gorm.Model{ID: 1},
		Username:     "admin",
		RealName:     "Admin",
		Role:         "team_leader",
		DepartmentID: 8,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mock user"})
		return
	}

	// 2. Create Reagent Requests (Historical and Current) with varied catalog indexes
	requests := []models.ReagentRequest{
		{RequestorID: user.ID, ReagentCatalogID: catalogs[0].ID, Quantity: 20, Status: "已入库", Remarks: "季度常规采购"},
		{RequestorID: user.ID, ReagentCatalogID: catalogs[1].ID, Quantity: 4, Status: "已入库", Remarks: "色谱分析用"},
		{RequestorID: user.ID, ReagentCatalogID: catalogs[4].ID, Quantity: 6, Status: "已入库", Remarks: "消解实验用酸"},
		{RequestorID: user.ID, ReagentCatalogID: catalogs[8].ID, Quantity: 2, Status: "已入库", Remarks: "配制生理盐水缓冲液"},
		{RequestorID: user.ID, ReagentCatalogID: catalogs[9].ID, Quantity: 1, Status: "已入库", Remarks: "细胞培养"},
		{RequestorID: user.ID, ReagentCatalogID: catalogs[2].ID, Quantity: 5, Status: "待处理", Remarks: "溶剂即将耗尽"},
		{RequestorID: user.ID, ReagentCatalogID: catalogs[11].ID, Quantity: 2, Status: "采购中", Remarks: "滴定分析标准品补充"},
	}

	for i := range requests {
		if err := tx.FirstOrCreate(&requests[i], models.ReagentRequest{Remarks: requests[i].Remarks}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed requests"})
			return
		}
	}

	// 3. Create Items (Only for Fulfilled requests)
	for _, req := range requests {
		if req.Status != "已入库" {
			continue
		}

		// Find relevant catalog to get capacity
		var cat models.ReagentCatalog
		tx.First(&cat, req.ReagentCatalogID)

		capacity := 500.0 // Default

		for j := 0; j < req.Quantity; j++ {
			// Randomize Status
			statusObj := "在库"
			location := cat.Storage
			remaining := capacity

			randNum := rand.Intn(100)
			if randNum < 20 {
				statusObj = "已耗尽"
				remaining = 0
			} else if randNum < 40 {
				statusObj = "已到货"
				location = "分拣区(临时区)"
			} else if randNum < 50 {
				statusObj = "在库"
				remaining = capacity

				// Randomize realistic labs
				labOpts := []string{"E309", "E307", "F103", "F309"}
				location = labOpts[rand.Intn(len(labOpts))]
			}

			item := models.ReagentItem{
				UUID:             uuid.New().String(),
				ReagentRequestID: req.ID,
				ReagentCatalogID: req.ReagentCatalogID,
				Status:           statusObj,
				Location:         location,
				Capacity:         capacity,
				RemainingVolume:  remaining,
				BatchNumber:      fmt.Sprintf("BATCH-%d-%d", time.Now().Year(), rand.Intn(1000)),
				ExpiryDate:       time.Now().AddDate(1, 0, 0), // 1 year from now
			}
			if err := tx.Create(&item).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to seed items"})
				return
			}

			// Create Log
			logAction := "入库登记"
			if item.Status == "已耗尽" {
				logAction = "空瓶核销"
			}
			if item.Status == "已到货" {
				logAction = "变更信息"
			}
			log := models.ReagentLog{
				ReagentItemID: item.UUID,
				UserID:        user.ID,
				Action:        logAction,
				Quantity:      capacity, // The full capacity was involved in the action
				Remarks:       "系统初始化演示数据生成",
			}
			tx.Create(&log)
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Reagent demo data seeded successfully"})
}

// SeedTeamInventory 为各研发团队注入在库试剂 seed 数据（真实写入数据库，非 Mock）
func SeedTeamInventory(c *gin.Context) {
	tx := database.DB.Begin()

	// 各团队负责人用户 ID（对应已有数据：分析团队=21, 研发A=26, 研发B=30, 研发C=34）
	type teamSeed struct {
		userID    uint
		teamLabel string
	}
	teams := []teamSeed{
		{userID: 21, teamLabel: "分析团队"},
		{userID: 26, teamLabel: "研发团队A"},
		{userID: 30, teamLabel: "研发团队B"},
		{userID: 34, teamLabel: "研发团队C"},
	}

	// 各团队的试剂清单（CAS 号 → 数量 → 库位）
	type reagentEntry struct {
		casNumber string
		quantity  int
		location  string
		remarks   string
	}
	teamReagents := map[uint][]reagentEntry{
		21: {
			{"64-17-5", 3, "E309", "分析用无水乙醇"},
			{"67-64-1", 2, "F103", "色谱流动相丙酮"},
			{"7647-01-0", 4, "E307", "消解用盐酸"},
		},
		26: {
			{"67-56-1", 3, "E309", "提取用甲醇"},
			{"108-88-3", 2, "F103", "合成用甲苯"},
			{"7664-93-9", 2, "E307", "酸消解硫酸"},
			{"1310-73-2", 2, "F309", "碱处理氢氧化钠"},
		},
		30: {
			{"64-17-5", 4, "E309", "洗涤用无水乙醇"},
			{"50-99-7", 2, "E309", "培养基葡萄糖"},
			{"7647-14-5", 3, "E309", "生理盐水氯化钠"},
			{"9002-93-1", 2, "F103", "细胞裂解 Triton X-100"},
		},
		34: {
			{"67-64-1", 3, "F103", "反应溶剂丙酮"},
			{"6381-92-6", 2, "F309", "螯合剂 EDTA 二钠"},
			{"1336-21-6", 3, "F309", "氨水碱化处理"},
		},
	}

	_ = teams // suppress unused warning
	for _, t := range teams {
		entries, ok := teamReagents[t.userID]
		if !ok {
			continue
		}
		for _, entry := range entries {
			// 找到或创建 catalog
			var cat models.ReagentCatalog
			if err := tx.Where("cas_number = ?", entry.casNumber).First(&cat).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "catalog not found for CAS " + entry.casNumber + ", please run seed_reagents first"})
				return
			}

			// 创建申购单
			req := models.ReagentRequest{
				RequestorID:      t.userID,
				ReagentCatalogID: cat.ID,
				Quantity:         entry.quantity,
				Status:           "已入库",
				Remarks:          entry.remarks,
			}
			if err := tx.Create(&req).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request: " + err.Error()})
				return
			}

			// 为每瓶创建 ReagentItem（状态直接为在库）
			for i := 0; i < entry.quantity; i++ {
				remaining := 500.0
				// 最后一瓶剩余量随机减少，模拟使用中
				if i == entry.quantity-1 && entry.quantity > 1 {
					remaining = float64(100 + rand.Intn(350))
				}
				item := models.ReagentItem{
					UUID:             uuid.New().String(),
					ReagentRequestID: req.ID,
					ReagentCatalogID: cat.ID,
					Status:           "在库",
					Location:         entry.location,
					Capacity:         500.0,
					RemainingVolume:  remaining,
					BatchNumber:      fmt.Sprintf("BATCH-%d-%d", time.Now().Year(), rand.Intn(9000)+1000),
					ExpiryDate:       time.Now().AddDate(1, rand.Intn(12), 0),
				}
				if err := tx.Create(&item).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create item: " + err.Error()})
					return
				}
				// 记录入库日志
				log := models.ReagentLog{
					ReagentItemID: item.UUID,
					UserID:        t.userID,
					Action:        "扫码入库",
					Quantity:      500.0,
					Remarks:       "团队库存 seed 数据初始化",
				}
				tx.Create(&log)
			}
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Team inventory seed data created successfully"})
}
