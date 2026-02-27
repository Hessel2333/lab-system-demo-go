package database

import (
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
