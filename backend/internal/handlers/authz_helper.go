package handlers

import (
	"net/http"

	"lab-system-backend/internal/authz"

	"github.com/gin-gonic/gin"
)

func authorizeAction(c *gin.Context, req authz.PermissionRequest, fallbackUser int) (authz.Actor, bool) {
	userID := getUserIDOrDefault(c, fallbackUser)
	actor, err := authz.ResolveActor(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "invalid actor",
			"user_id": userID,
		})
		return authz.Actor{}, false
	}

	decision := authz.Evaluate(actor, req)
	if !decision.Allowed {
		c.JSON(http.StatusForbidden, gin.H{
			"error":         "permission denied",
			"reason":        decision.Reason,
			"resource":      req.Resource,
			"action":        req.Action,
			"scope":         req.Scope,
			"actor_user_id": actor.UserID,
			"actor_role":    actor.Role,
		})
		return actor, false
	}
	return actor, true
}
