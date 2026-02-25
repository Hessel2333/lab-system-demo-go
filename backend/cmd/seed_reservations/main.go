package main

import (
	"fmt"
	"math/rand"
	"time"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"
)

const API_BASE = "http://localhost:8080/api"

// Local structs removed, using internal/models

var projects = []string{
	"细胞凋亡实验",
	"蛋白纯化分析",
	"药物筛选初筛",
}

func main() {
	database.InitDB()
	rand.Seed(time.Now().UnixNano())

	// 1. Get Instruments (via DB directly for simplicity in seeder)
	var instruments []models.Instrument
	if err := database.DB.Find(&instruments).Error; err != nil {
		fmt.Println("Error fetching instruments:", err)
		return
	}

	if len(instruments) == 0 {
		fmt.Println("No instruments found. Please seed instruments first.")
		return
	}

	// 2. Get Real Users
	var dbUsers []models.User
	if err := database.DB.Preload("Department").Find(&dbUsers).Error; err != nil {
		panic(err)
	}

	// Filter out potential "current" user if we add auth later, but for now use all
	// Create a list of potential booking users
	var bookingUsers []models.User
	for _, u := range dbUsers {
		// Maybe exclude admin/specialists from "usage" booking?
		// Let's keep everyone for variety
		bookingUsers = append(bookingUsers, u)
	}

	targetInstrument := instruments[0]
	fmt.Printf("Seeding reservations for instrument: %s (ID: %d)\n", targetInstrument.Name, targetInstrument.ID)

	// 3. Clean existing reservations for this instrument
	database.DB.Where("instrument_id = ?", targetInstrument.ID).Delete(&models.Reservation{})
	fmt.Println("Cleared existing reservations.")

	// 2. Determine Date Range (Last 2 weeks)
	now := time.Now()
	offsetToMonday := int(now.Weekday())
	if offsetToMonday == 0 {
		offsetToMonday = 7
	}
	offsetToMonday = offsetToMonday - 1

	thisMonday := now.AddDate(0, 0, -offsetToMonday)
	lastMonday := thisMonday.AddDate(0, 0, -7)

	// Generate exactly 2-3 reservations total
	totalReservations := 2 + rand.Intn(2) // 2 or 3

	for i := 0; i < totalReservations; i++ {
		// Pick a random day in the 14-day window
		dayOffset := rand.Intn(14)
		date := lastMonday.AddDate(0, 0, dayOffset)

		// Pick a random hour 9-17
		startHour := 9 + rand.Intn(8)
		duration := 1 + rand.Intn(2) // 1-2 hours

		// Pick random user
		user := bookingUsers[rand.Intn(len(bookingUsers))]

		startTime := time.Date(date.Year(), date.Month(), date.Day(), startHour, 0, 0, 0, time.Local)
		endTime := startTime.Add(time.Duration(duration) * time.Hour)

		// Project name based on lab context
		desc := projects[rand.Intn(len(projects))]

		// If specialist, maybe maintenance
		typeStr := "usage"
		if user.Role == "safety_specialist" || user.Role == "measurement_specialist" {
			typeStr = "maintenance"
			desc = "设备维护/校准"
		}

		res := models.Reservation{
			InstrumentID: targetInstrument.ID,
			UserID:       user.Username,                                               // Use Username as ID
			UserName:     fmt.Sprintf("%s (%s)", user.RealName, user.Department.Name), // Display: Name (Dept)
			StartTime:    startTime,
			EndTime:      endTime,
			Type:         typeStr,
			Description:  desc,
			Status:       "active",
		}

		if err := database.DB.Create(&res).Error; err == nil {
			fmt.Printf("Created: %s %s - %s\n", res.StartTime.Format("2006-01-02 15:04"), res.UserName, res.Description)
		}
	}
}

func seedDay(instrumentID uint, date time.Time, users []models.User) {
	// Deprecated for sparse generation
}
