package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func resolveDefaultDBPath() string {
	// Prefer locating database under backend/data regardless of current working directory.
	if _, file, _, ok := runtime.Caller(0); ok {
		backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
		return filepath.Join(backendRoot, "data", "lab_system.db")
	}
	return "./data/lab_system.db"
}

func InitDB() {
	var err error

	dbPath := os.Getenv("LAB_DB_PATH")
	if dbPath == "" {
		dbPath = resolveDefaultDBPath()
	}
	if !filepath.IsAbs(dbPath) {
		dbPath = filepath.Clean(dbPath)
	}

	// Create database folder if not exists
	if mkErr := os.MkdirAll(filepath.Dir(dbPath), 0755); mkErr != nil {
		log.Fatal("Failed to create database folder:", mkErr)
	}

	// Configure Logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Info,
		},
	)

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connection established:", dbPath)
}

// CleanupLegacySchema physically drops legacy coupling columns kept from historical versions.
func CleanupLegacySchema() error {
	if DB == nil {
		return fmt.Errorf("database is not initialized")
	}

	// SQLite 下用重建表的方式清理旧列，避免 GORM DropColumn 在已移除字段时崩溃
	if DB.Dialector.Name() != "sqlite" {
		if DB.Migrator().HasColumn("reagent_items", "reagent_request_id") {
			if err := DB.Migrator().DropColumn("reagent_items", "reagent_request_id"); err != nil {
				return fmt.Errorf("drop reagent_items.reagent_request_id failed: %w", err)
			}
			log.Println("Schema cleanup: dropped reagent_items.reagent_request_id")
		}
		if DB.Migrator().HasColumn("procurement_batch_items", "matched_request_id") {
			if err := DB.Migrator().DropColumn("procurement_batch_items", "matched_request_id"); err != nil {
				return fmt.Errorf("drop procurement_batch_items.matched_request_id failed: %w", err)
			}
			log.Println("Schema cleanup: dropped procurement_batch_items.matched_request_id")
		}
		return nil
	}

	if DB.Migrator().HasColumn("reagent_items", "reagent_request_id") {
		if err := rebuildReagentItemsWithoutRequestID(); err != nil {
			return err
		}
		log.Println("Schema cleanup: dropped reagent_items.reagent_request_id")
	}

	hasMatchedRequest := DB.Migrator().HasColumn("procurement_batch_items", "matched_request_id")
	hasSuggestion := DB.Migrator().HasColumn("procurement_batch_items", "request_suggestion")
	if hasMatchedRequest || !hasSuggestion {
		if err := rebuildProcurementBatchItems(hasSuggestion); err != nil {
			return err
		}
		if hasMatchedRequest {
			log.Println("Schema cleanup: dropped procurement_batch_items.matched_request_id")
		}
		if !hasSuggestion {
			log.Println("Schema cleanup: added procurement_batch_items.request_suggestion")
		}
	}

	return nil
}

func rebuildReagentItemsWithoutRequestID() error {
	if err := DB.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		return fmt.Errorf("disable foreign keys failed: %w", err)
	}
	defer DB.Exec("PRAGMA foreign_keys = ON")

	tx := DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction failed: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	createSQL := `
CREATE TABLE reagent_items_new (
  uuid char(36) PRIMARY KEY,
  reagent_catalog_id integer,
  status text,
  location text,
  cabinet_id integer,
  capacity real,
  remaining_volume real,
  batch_number text,
  expiry_date datetime,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  CONSTRAINT fk_reagent_items_reagent_catalog FOREIGN KEY (reagent_catalog_id) REFERENCES reagent_catalogs(id),
  CONSTRAINT fk_reagent_items_cabinet FOREIGN KEY (cabinet_id) REFERENCES reagent_cabinets(id)
)`
	if err := tx.Exec(createSQL).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("create reagent_items_new failed: %w", err)
	}

	copySQL := `
INSERT INTO reagent_items_new (
  uuid, reagent_catalog_id, status, location, cabinet_id, capacity,
  remaining_volume, batch_number, expiry_date, created_at, updated_at, deleted_at
)
SELECT
  uuid, reagent_catalog_id, status, location, cabinet_id, capacity,
  remaining_volume, batch_number, expiry_date, created_at, updated_at, deleted_at
FROM reagent_items`
	if err := tx.Exec(copySQL).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("copy reagent_items data failed: %w", err)
	}

	if err := tx.Exec("DROP TABLE reagent_items").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("drop reagent_items failed: %w", err)
	}
	if err := tx.Exec("ALTER TABLE reagent_items_new RENAME TO reagent_items").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("rename reagent_items_new failed: %w", err)
	}
	if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_reagent_items_deleted_at ON reagent_items(deleted_at)").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("create reagent_items index failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit reagent_items rebuild failed: %w", err)
	}
	return nil
}

func rebuildProcurementBatchItems(hasSuggestion bool) error {
	if err := DB.Exec("PRAGMA foreign_keys = OFF").Error; err != nil {
		return fmt.Errorf("disable foreign keys failed: %w", err)
	}
	defer DB.Exec("PRAGMA foreign_keys = ON")

	tx := DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin transaction failed: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	createSQL := `
CREATE TABLE procurement_batch_items_new (
  id integer PRIMARY KEY AUTOINCREMENT,
  batch_id integer,
  reagent_name text,
  cas_number text,
  quantity integer,
  unit text,
  unit_price real,
  supplier text,
  material_category text,
  product_category text,
  matched_catalog_id integer,
  matched_user_id integer,
  request_suggestion text,
  match_status text,
  receive_status text DEFAULT "待收货",
  received_quantity integer DEFAULT 0,
  created_at datetime,
  updated_at datetime,
  deleted_at datetime,
  row_hash text,
  CONSTRAINT fk_procurement_batches_items FOREIGN KEY (batch_id) REFERENCES procurement_batches(id)
)`
	if err := tx.Exec(createSQL).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("create procurement_batch_items_new failed: %w", err)
	}

	suggestionExpr := "''"
	if hasSuggestion {
		suggestionExpr = "request_suggestion"
	}
	copySQL := fmt.Sprintf(`
INSERT INTO procurement_batch_items_new (
  id, batch_id, reagent_name, cas_number, quantity, unit, unit_price, supplier,
  material_category, product_category, matched_catalog_id, matched_user_id, request_suggestion,
  match_status, receive_status, received_quantity, created_at, updated_at, deleted_at, row_hash
)
SELECT
  id, batch_id, reagent_name, cas_number, quantity, unit, unit_price, supplier,
  material_category, product_category, matched_catalog_id, matched_user_id, %s,
  match_status, receive_status, received_quantity, created_at, updated_at, deleted_at, row_hash
FROM procurement_batch_items`, suggestionExpr)
	if err := tx.Exec(copySQL).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("copy procurement_batch_items data failed: %w", err)
	}

	if err := tx.Exec("DROP TABLE procurement_batch_items").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("drop procurement_batch_items failed: %w", err)
	}
	if err := tx.Exec("ALTER TABLE procurement_batch_items_new RENAME TO procurement_batch_items").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("rename procurement_batch_items_new failed: %w", err)
	}
	if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_procurement_batch_items_deleted_at ON procurement_batch_items(deleted_at)").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("create procurement_batch_items deleted index failed: %w", err)
	}
	if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_procurement_batch_items_row_hash ON procurement_batch_items(row_hash)").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("create procurement_batch_items row_hash index failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit procurement_batch_items rebuild failed: %w", err)
	}
	return nil
}
