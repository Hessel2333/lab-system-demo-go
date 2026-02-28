package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"

	"github.com/gin-gonic/gin"
)

var allowedUserRoles = map[string]struct{}{
	"admin":                  {},
	"director":               {},
	"team_leader":            {},
	"member":                 {},
	"researcher":             {},
	"procurement":            {},
	"procurement_specialist": {},
	"measurement_specialist": {},
	"safety_specialist":      {},
}

func normalizeUserRole(role string) string {
	return strings.TrimSpace(strings.ToLower(role))
}

func validateUserRole(role string) error {
	normalized := normalizeUserRole(role)
	if normalized == "" {
		return fmt.Errorf("role is required")
	}
	if _, ok := allowedUserRoles[normalized]; !ok {
		return fmt.Errorf("invalid role: %s", role)
	}
	return nil
}

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
	includeChildrenRaw := strings.TrimSpace(strings.ToLower(c.DefaultQuery("include_children", "true")))
	includeChildren := includeChildrenRaw != "false" && includeChildrenRaw != "0" && includeChildrenRaw != "no"

	query := database.DB.Model(&models.User{}).Preload("Department")

	if deptID != "" {
		if includeChildren {
			deptIDUint, err := strconv.ParseUint(deptID, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department_id"})
				return
			}
			var allDepts []models.Department
			if err := database.DB.Find(&allDepts).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch departments"})
				return
			}
			scopeIDs := collectDeptScopeIDs(allDepts, uint(deptIDUint))
			query = query.Where("department_id IN ?", scopeIDs)
		} else {
			query = query.Where("department_id = ?", deptID)
		}
	}

	var users []models.User
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

func collectDeptScopeIDs(all []models.Department, rootID uint) []uint {
	childrenMap := make(map[uint][]uint)
	for _, dept := range all {
		if dept.ParentID == nil {
			continue
		}
		parent := *dept.ParentID
		childrenMap[parent] = append(childrenMap[parent], dept.ID)
	}

	visited := make(map[uint]bool)
	queue := []uint{rootID}
	visited[rootID] = true
	result := []uint{rootID}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, child := range childrenMap[current] {
			if visited[child] {
				continue
			}
			visited[child] = true
			result = append(result, child)
			queue = append(queue, child)
		}
	}

	return result
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
		Username             string `json:"username" binding:"required"`
		RealName             string `json:"real_name" binding:"required"`
		DepartmentID         uint   `json:"department_id" binding:"required"`
		Role                 string `json:"role" binding:"required"`
		IsDispenseKeyHolderA bool   `json:"is_dispense_key_holder_a"`
		IsDispenseKeyHolderB bool   `json:"is_dispense_key_holder_b"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.IsDispenseKeyHolderA && input.IsDispenseKeyHolderB {
		c.JSON(http.StatusBadRequest, gin.H{"error": "同一用户不能同时担任A/B双签持有人"})
		return
	}
	if err := validateUserRole(input.Role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := models.User{
		Username:             input.Username,
		RealName:             input.RealName,
		DepartmentID:         input.DepartmentID,
		Role:                 normalizeUserRole(input.Role),
		IsDispenseKeyHolderA: input.IsDispenseKeyHolderA,
		IsDispenseKeyHolderB: input.IsDispenseKeyHolderB,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	ensureUniqueDispenseKeyHolders(user.ID, input.IsDispenseKeyHolderA, input.IsDispenseKeyHolderB)

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
		Username             *string `json:"username"`
		RealName             *string `json:"real_name"`
		DepartmentID         *uint   `json:"department_id"`
		Role                 *string `json:"role"`
		IsDispenseKeyHolderA *bool   `json:"is_dispense_key_holder_a"`
		IsDispenseKeyHolderB *bool   `json:"is_dispense_key_holder_b"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	finalA := user.IsDispenseKeyHolderA
	finalB := user.IsDispenseKeyHolderB
	if input.IsDispenseKeyHolderA != nil {
		finalA = *input.IsDispenseKeyHolderA
	}
	if input.IsDispenseKeyHolderB != nil {
		finalB = *input.IsDispenseKeyHolderB
	}
	if finalA && finalB {
		c.JSON(http.StatusBadRequest, gin.H{"error": "同一用户不能同时担任A/B双签持有人"})
		return
	}

	updates := map[string]interface{}{}
	if input.Username != nil {
		updates["username"] = *input.Username
	}
	if input.RealName != nil {
		updates["real_name"] = *input.RealName
	}
	if input.DepartmentID != nil {
		updates["department_id"] = *input.DepartmentID
	}
	if input.Role != nil {
		if err := validateUserRole(*input.Role); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		updates["role"] = normalizeUserRole(*input.Role)
	}
	if input.IsDispenseKeyHolderA != nil {
		updates["is_dispense_key_holder_a"] = *input.IsDispenseKeyHolderA
	}
	if input.IsDispenseKeyHolderB != nil {
		updates["is_dispense_key_holder_b"] = *input.IsDispenseKeyHolderB
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	ensureUniqueDispenseKeyHolders(user.ID, finalA, finalB)

	if err := database.DB.Preload("Department").First(&user, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated user"})
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

func ensureUniqueDispenseKeyHolders(userID uint, asA bool, asB bool) {
	if asA {
		database.DB.Model(&models.User{}).
			Where("id <> ? AND is_dispense_key_holder_a = ?", userID, true).
			Update("is_dispense_key_holder_a", false)
	}
	if asB {
		database.DB.Model(&models.User{}).
			Where("id <> ? AND is_dispense_key_holder_b = ?", userID, true).
			Update("is_dispense_key_holder_b", false)
	}
}

func GetUserReagentPermissions(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_dispense_key_holder_a": user.IsDispenseKeyHolderA,
		"is_dispense_key_holder_b": user.IsDispenseKeyHolderB,
	})
}

func UpdateUserReagentPermissions(c *gin.Context) {
	id := c.Param("id")
	userID64, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var input struct {
		IsDispenseKeyHolderA bool `json:"is_dispense_key_holder_a"`
		IsDispenseKeyHolderB bool `json:"is_dispense_key_holder_b"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if input.IsDispenseKeyHolderA && input.IsDispenseKeyHolderB {
		c.JSON(http.StatusBadRequest, gin.H{"error": "同一用户不能同时担任A/B双签持有人"})
		return
	}

	if err := database.DB.Model(&user).Updates(map[string]interface{}{
		"is_dispense_key_holder_a": input.IsDispenseKeyHolderA,
		"is_dispense_key_holder_b": input.IsDispenseKeyHolderB,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reagent permissions"})
		return
	}

	ensureUniqueDispenseKeyHolders(uint(userID64), input.IsDispenseKeyHolderA, input.IsDispenseKeyHolderB)
	c.JSON(http.StatusOK, gin.H{
		"is_dispense_key_holder_a": input.IsDispenseKeyHolderA,
		"is_dispense_key_holder_b": input.IsDispenseKeyHolderB,
	})
}
