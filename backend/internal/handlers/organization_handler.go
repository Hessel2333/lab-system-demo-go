package handlers

import (
	"net/http"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetDepartments returns the hierarchical organization tree
func GetDepartments(c *gin.Context) {
	var depts []models.Department
	// Preload nothing for now, we will build tree manually or fetch all and build
	if err := database.DB.Find(&depts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch departments"})
		return
	}

	// Build Tree
	tree := buildDeptTree(depts, nil)
	c.JSON(http.StatusOK, tree)
}

// GetUsers returns users, optionally filtered by department_id
func GetUsers(c *gin.Context) {
	deptID := c.Query("department_id")

	query := database.DB.Model(&models.User{}).Preload("Department")

	if deptID != "" {
		// Include children departments? For simple MVP, just exact match
		query = query.Where("department_id = ?", deptID)
	}

	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// Helper: Build Tree directly (assuming small dataset)
func buildDeptTree(all []models.Department, parentID *uint) []models.Department {
	var nodes []models.Department
	for _, dept := range all {
		// Check if this dept is a child of the current parentID
		isChild := false
		if parentID == nil {
			if dept.ParentID == nil {
				isChild = true
			}
		} else {
			if dept.ParentID != nil && *dept.ParentID == *parentID {
				isChild = true
			}
		}

		if isChild {
			dept.Children = buildDeptTree(all, &dept.ID)
			nodes = append(nodes, dept)
		}
	}
	return nodes
}

// CreateUser creates a new user
func CreateUser(c *gin.Context) {
	var input struct {
		Username     string `json:"username" binding:"required"`
		RealName     string `json:"real_name" binding:"required"`
		DepartmentID uint   `json:"department_id" binding:"required"`
		Role         string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Username:     input.Username,
		RealName:     input.RealName,
		DepartmentID: input.DepartmentID,
		Role:         input.Role,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates an existing user
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input struct {
		Username     string `json:"username"`
		RealName     string `json:"real_name"`
		DepartmentID uint   `json:"department_id"`
		Role         string `json:"role"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&user).Updates(models.User{
		Username:     input.Username,
		RealName:     input.RealName,
		DepartmentID: input.DepartmentID,
		Role:         input.Role,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user (soft delete)
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}
