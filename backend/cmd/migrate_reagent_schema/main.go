package main

import (
	"log"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"
)

func main() {
	database.InitDB()

	if err := database.DB.AutoMigrate(
		&models.Instrument{},
		&models.Reservation{},
		&models.User{},
		&models.Department{},
		&models.UserQualification{},
		&models.Supplier{},
		&models.ReagentCatalog{},
		&models.ReagentCabinet{},
		&models.ReagentRequest{},
		&models.ReagentItem{},
		&models.ReagentLog{},
		&models.ProcurementBatch{},
		&models.ProcurementBatchItem{},
		&models.ReagentDispenseRequest{},
	); err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}

	if err := database.CleanupLegacySchema(); err != nil {
		log.Fatalf("CleanupLegacySchema failed: %v", err)
	}

	log.Println("Schema migration completed: legacy coupling columns removed.")
}
