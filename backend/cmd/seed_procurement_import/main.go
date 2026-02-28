package main

import (
	"fmt"
	"log"
	"time"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"
)

func uintPtr(v uint) *uint {
	return &v
}

func ensureBaseCatalogs() ([]models.ReagentCatalog, error) {
	var catalogs []models.ReagentCatalog
	if err := database.DB.Order("id asc").Find(&catalogs).Error; err != nil {
		return nil, err
	}
	if len(catalogs) > 0 {
		return catalogs, nil
	}

	seeds := []models.ReagentCatalog{
		{CASNumber: "64-17-5", Name: "无水乙醇", Alias: "Ethanol", Category: "有机溶剂", Unit: "500ml", AlertThreshold: 10},
		{CASNumber: "67-64-1", Name: "丙酮", Alias: "Acetone", Category: "有机溶剂", Unit: "500ml", AlertThreshold: 8, IsControlled: true},
		{CASNumber: "67-56-1", Name: "甲醇", Alias: "Methanol", Category: "有机溶剂", Unit: "500ml", AlertThreshold: 8, IsControlled: true},
	}
	for i := range seeds {
		if err := database.DB.Create(&seeds[i]).Error; err != nil {
			return nil, err
		}
	}

	catalogs = nil
	if err := database.DB.Order("id asc").Find(&catalogs).Error; err != nil {
		return nil, err
	}
	return catalogs, nil
}

func ensureBaseUsers() (models.User, []models.User, error) {
	var procurement models.User
	if err := database.DB.Where("role = ?", "procurement").Order("id asc").First(&procurement).Error; err != nil {
		procurement = models.User{Username: "procurement_demo", RealName: "采购演示员", Role: "procurement"}
		if createErr := database.DB.Create(&procurement).Error; createErr != nil {
			return models.User{}, nil, createErr
		}
	}

	var assignees []models.User
	if err := database.DB.Where("role IN ?", []string{"researcher", "team_leader"}).Order("id asc").Limit(3).Find(&assignees).Error; err != nil {
		return models.User{}, nil, err
	}
	if len(assignees) == 0 {
		assignees = append(assignees, procurement)
	}
	return procurement, assignees, nil
}

func main() {
	database.InitDB()

	if err := database.DB.AutoMigrate(
		&models.ReagentCatalog{},
		&models.User{},
		&models.ProcurementBatch{},
		&models.ProcurementBatchItem{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
	if err := database.CleanupLegacySchema(); err != nil {
		log.Fatalf("CleanupLegacySchema failed: %v", err)
	}

	catalogs, err := ensureBaseCatalogs()
	if err != nil {
		log.Fatalf("prepare catalogs failed: %v", err)
	}
	procurement, assignees, err := ensureBaseUsers()
	if err != nil {
		log.Fatalf("prepare users failed: %v", err)
	}

	if err := database.DB.Exec("DELETE FROM procurement_batch_items").Error; err != nil {
		log.Fatalf("clear procurement_batch_items failed: %v", err)
	}
	if err := database.DB.Exec("DELETE FROM procurement_batches").Error; err != nil {
		log.Fatalf("clear procurement_batches failed: %v", err)
	}

	periodCurrent := time.Now().Format("2006-01")
	periodPrev := time.Now().AddDate(0, -1, 0).Format("2006-01")

	pendingBatch := models.ProcurementBatch{
		UploaderID:  procurement.ID,
		Period:      periodCurrent,
		OrderNumber: fmt.Sprintf("DEMO-IMPORT-%s-A", time.Now().Format("20060102")),
		Status:      "待确认",
	}
	if err := database.DB.Create(&pendingBatch).Error; err != nil {
		log.Fatalf("create pending batch failed: %v", err)
	}

	pendingItems := []models.ProcurementBatchItem{
		{
			BatchID:           pendingBatch.ID,
			RowHash:           fmt.Sprintf("demo-a-1-%d", time.Now().UnixNano()),
			ReagentName:       catalogs[0].Name,
			CASNumber:         catalogs[0].CASNumber,
			Quantity:          6,
			Unit:              "瓶",
			UnitPrice:         88.5,
			Supplier:          "国药试剂",
			MaterialCategory:  "化工",
			ProductCategory:   "试剂",
			MatchedCatalogID:  uintPtr(catalogs[0].ID),
			MatchedUserID:     uintPtr(assignees[0].ID),
			RequestSuggestion: fmt.Sprintf("建议需求 · %s · %s ×6瓶", assignees[0].RealName, catalogs[0].Name),
			MatchStatus:       "自动匹配",
			ReceiveStatus:     "待收货",
			ReceivedQuantity:  0,
		},
		{
			BatchID:           pendingBatch.ID,
			RowHash:           fmt.Sprintf("demo-a-2-%d", time.Now().UnixNano()),
			ReagentName:       catalogs[1].Name,
			CASNumber:         catalogs[1].CASNumber,
			Quantity:          3,
			Unit:              "瓶",
			UnitPrice:         128.0,
			Supplier:          "阿拉丁",
			MaterialCategory:  "化工",
			ProductCategory:   "试剂",
			MatchedCatalogID:  uintPtr(catalogs[1].ID),
			MatchedUserID:     uintPtr(assignees[len(assignees)-1].ID),
			RequestSuggestion: fmt.Sprintf("建议需求 · %s · %s ×3瓶", assignees[len(assignees)-1].RealName, catalogs[1].Name),
			MatchStatus:       "手动匹配",
			ReceiveStatus:     "待收货",
			ReceivedQuantity:  0,
		},
		{
			BatchID:          pendingBatch.ID,
			RowHash:          fmt.Sprintf("demo-a-3-%d", time.Now().UnixNano()),
			ReagentName:      "特殊耗材（待人工归类）",
			CASNumber:        "",
			Quantity:         2,
			Unit:             "件",
			UnitPrice:        35.0,
			Supplier:         "综合供应商",
			MaterialCategory: "化工",
			ProductCategory:  "辅料",
			MatchStatus:      "未匹配",
			ReceiveStatus:    "待收货",
			ReceivedQuantity: 0,
		},
		{
			BatchID:          pendingBatch.ID,
			RowHash:          fmt.Sprintf("demo-a-4-%d", time.Now().UnixNano()),
			ReagentName:      "防护手套",
			CASNumber:        "",
			Quantity:         20,
			Unit:             "副",
			UnitPrice:        4.6,
			Supplier:         "劳保商城",
			MaterialCategory: "劳保用品",
			ProductCategory:  "耗材",
			MatchStatus:      "已忽略",
			ReceiveStatus:    "待收货",
			ReceivedQuantity: 0,
		},
	}
	if err := database.DB.Create(&pendingItems).Error; err != nil {
		log.Fatalf("create pending items failed: %v", err)
	}

	confirmedBatch := models.ProcurementBatch{
		UploaderID:  procurement.ID,
		Period:      periodPrev,
		OrderNumber: fmt.Sprintf("DEMO-IMPORT-%s-B", time.Now().Format("20060102")),
		Status:      "已确认",
	}
	if err := database.DB.Create(&confirmedBatch).Error; err != nil {
		log.Fatalf("create confirmed batch failed: %v", err)
	}

	confirmedItems := []models.ProcurementBatchItem{
		{
			BatchID:           confirmedBatch.ID,
			RowHash:           fmt.Sprintf("demo-b-1-%d", time.Now().UnixNano()),
			ReagentName:       catalogs[2%len(catalogs)].Name,
			CASNumber:         catalogs[2%len(catalogs)].CASNumber,
			Quantity:          5,
			Unit:              "瓶",
			UnitPrice:         96.0,
			Supplier:          "默克",
			MaterialCategory:  "化工",
			ProductCategory:   "试剂",
			MatchedCatalogID:  uintPtr(catalogs[2%len(catalogs)].ID),
			MatchedUserID:     uintPtr(assignees[0].ID),
			RequestSuggestion: fmt.Sprintf("建议需求 · %s · %s ×5瓶", assignees[0].RealName, catalogs[2%len(catalogs)].Name),
			MatchStatus:       "自动匹配",
			ReceiveStatus:     "部分收货",
			ReceivedQuantity:  2,
		},
		{
			BatchID:           confirmedBatch.ID,
			RowHash:           fmt.Sprintf("demo-b-2-%d", time.Now().UnixNano()),
			ReagentName:       catalogs[0].Name,
			CASNumber:         catalogs[0].CASNumber,
			Quantity:          4,
			Unit:              "瓶",
			UnitPrice:         90.0,
			Supplier:          "国药试剂",
			MaterialCategory:  "化工",
			ProductCategory:   "试剂",
			MatchedCatalogID:  uintPtr(catalogs[0].ID),
			MatchedUserID:     uintPtr(assignees[len(assignees)-1].ID),
			RequestSuggestion: fmt.Sprintf("建议需求 · %s · %s ×4瓶", assignees[len(assignees)-1].RealName, catalogs[0].Name),
			MatchStatus:       "手动匹配",
			ReceiveStatus:     "待收货",
			ReceivedQuantity:  0,
		},
	}
	if err := database.DB.Create(&confirmedItems).Error; err != nil {
		log.Fatalf("create confirmed items failed: %v", err)
	}

	var batchCount int64
	var itemCount int64
	database.DB.Model(&models.ProcurementBatch{}).Count(&batchCount)
	database.DB.Model(&models.ProcurementBatchItem{}).Count(&itemCount)

	log.Printf("Procurement import demo seeded. batches=%d items=%d", batchCount, itemCount)
}
