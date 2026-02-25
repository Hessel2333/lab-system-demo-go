package handlers

import (
	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetInstrument authorized users returns users who have permission for this instrument
func GetAuthorizedUsers(c *gin.Context) {
	instrumentIDStr := c.Param("id")

	// Fetch qualifications with status = true
	var qualifications []models.UserQualification
	if err := database.DB.Preload("User").Where("instrument_id = ? AND status = ?", instrumentIDStr, true).Find(&qualifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch authorized users"})
		return
	}

	var users []models.User
	for _, q := range qualifications {
		users = append(users, q.User)
	}

	c.JSON(http.StatusOK, users)
}

// UpdateInstrumentAdmin updates the admin of an instrument
func UpdateInstrumentAdmin(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Admin string `json:"admin" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&models.Instrument{}).Where("id = ?", id).Update("admin", input.Admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin updated"})
}
