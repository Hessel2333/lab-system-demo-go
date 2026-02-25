package handlers

import (
	"fmt"
	"net/http"
	"time"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetReservations fetches reservations, optionally filtered by instrument_id and time range
func GetReservations(c *gin.Context) {
	instrumentID := c.Query("instrument_id")
	startStr := c.Query("start")
	endStr := c.Query("end")

	query := database.DB.Model(&models.Reservation{}).Where("status != ?", "cancelled")

	if instrumentID != "" {
		query = query.Where("instrument_id = ?", instrumentID)
	}

	if startStr != "" && endStr != "" {
		startTime, _ := time.Parse(time.RFC3339, startStr)
		endTime, _ := time.Parse(time.RFC3339, endStr)
		// Find overlaps: (StartA <= EndB) and (EndA >= StartB)
		query = query.Where("start_time < ? AND end_time > ?", endTime, startTime)
	}

	var reservations []models.Reservation
	if err := query.Find(&reservations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
		return
	}

	c.JSON(http.StatusOK, reservations)
}

// CreateReservation creates a new reservation with basic conflict checking
func CreateReservation(c *gin.Context) {
	var input models.Reservation
	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Received Reservation Request: %+v\n", input)

	// Basic validation
	if input.StartTime.After(input.EndTime) {
		fmt.Println("Error: StartTime after EndTime")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Start time must be before end time"})
		return
	}

	// Conflict Check
	var count int64
	database.DB.Model(&models.Reservation{}).
		Where("instrument_id = ? AND status != ? AND start_time < ? AND end_time > ?",
			input.InstrumentID, "cancelled", input.EndTime, input.StartTime).
		Count(&count)

	if count > 0 {
		fmt.Println("Error: Conflict detected")
		c.JSON(http.StatusConflict, gin.H{"error": "Time slot already reserved"})
		return
	}

	input.Status = "active"
	if input.UserID == "" {
		input.UserID = "current" // Default to current user for UI-created reservations without auth
		input.UserName = "我"
	}

	if err := database.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

// CancelReservation cancels a reservation (soft delete logic or status update)
func CancelReservation(c *gin.Context) {
	id := c.Param("id")

	// Verify ownership (mock: only allow 'current' user or admin, skipping for MVP)

	result := database.DB.Model(&models.Reservation{}).Where("id = ?", id).Update("status", "cancelled")
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel reservation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservation cancelled"})
}
