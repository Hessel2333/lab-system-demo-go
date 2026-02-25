package handlers

import (
	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserPermissions returns all instruments with the user's permission status
func GetUserPermissions(c *gin.Context) {
	userIDStr := c.Param("id")

	// 1. Fetch all instruments
	var instruments []models.Instrument
	if err := database.DB.Find(&instruments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch instruments"})
		return
	}

	// 2. Fetch existing qualifications for this user
	var qualifications []models.UserQualification
	if err := database.DB.Where("user_id = ?", userIDStr).Find(&qualifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch qualifications"})
		return
	}

	// Map qualification status by InstrumentID
	qualMap := make(map[uint]bool)
	for _, q := range qualifications {
		qualMap[q.InstrumentID] = q.Status
	}

	// 3. Build response
	type InstrumentPermission struct {
		InstrumentID   uint   `json:"instrument_id"`
		InstrumentName string `json:"instrument_name"`
		HasPermission  bool   `json:"has_permission"`
	}

	var response []InstrumentPermission
	for _, inst := range instruments {
		hasPerm := false
		if val, ok := qualMap[inst.ID]; ok {
			hasPerm = val
		}

		response = append(response, InstrumentPermission{
			InstrumentID:   inst.ID,
			InstrumentName: inst.Name,
			HasPermission:  hasPerm,
		})
	}

	c.JSON(http.StatusOK, response)
}

// UpdateUserPermission updates the permission status for a user and instrument
func UpdateUserPermission(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input struct {
		InstrumentID uint `json:"instrument_id" binding:"required"`
		Status       bool `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Upsert logic
	var qualification models.UserQualification
	err = database.DB.Where("user_id = ? AND instrument_id = ?", userID, input.InstrumentID).First(&qualification).Error

	if err == gorm.ErrRecordNotFound {
		// Create new
		qualification = models.UserQualification{
			UserID:       uint(userID),
			InstrumentID: input.InstrumentID,
			Status:       input.Status,
		}
		if err := database.DB.Create(&qualification).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create permission"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	} else {
		// Update existing
		qualification.Status = input.Status
		if err := database.DB.Save(&qualification).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update permission"})
			return
		}
	}

	c.JSON(http.StatusOK, qualification)
}
