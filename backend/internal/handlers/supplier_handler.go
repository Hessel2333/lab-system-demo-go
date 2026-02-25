package handlers

import (
	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetSuppliers(c *gin.Context) {
	var suppliers []models.Supplier
	database.DB.Find(&suppliers)
	c.JSON(http.StatusOK, suppliers)
}

func CreateSupplier(c *gin.Context) {
	var input models.Supplier
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Create(&input)
	c.JSON(http.StatusOK, input)
}

func UpdateSupplier(c *gin.Context) {
	var supplier models.Supplier
	if err := database.DB.First(&supplier, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier not found"})
		return
	}
	var input models.Supplier
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	database.DB.Model(&supplier).Updates(input)
	c.JSON(http.StatusOK, supplier)
}

func DeleteSupplier(c *gin.Context) {
	database.DB.Delete(&models.Supplier{}, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"message": "Supplier deleted"})
}
