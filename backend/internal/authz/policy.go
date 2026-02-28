package authz

import (
	"errors"
	"strings"

	"lab-system-backend/internal/database"
	"lab-system-backend/internal/models"
)

type Resource string
type Action string
type Scope string

const (
	ResourceDispenseRequest  Resource = "dispense_request"
	ResourceProcurementBatch Resource = "procurement_batch"
	ResourceReagentItem      Resource = "reagent_item"
)

const (
	ActionApply            Action = "apply"
	ActionLeaderApprove    Action = "leader_approve"
	ActionKeyHolderConfirm Action = "keyholder_confirm"
	ActionConfirmBatch     Action = "confirm_batch"
	ActionReceive          Action = "receive"
	ActionCheckIn          Action = "checkin"
)

const (
	ScopeGlobal   Scope = "global"
	ScopeSelf     Scope = "self"
	ScopeTeam     Scope = "team"
	ScopeAssigned Scope = "assigned"
)

type Actor struct {
	UserID       uint   `json:"user_id"`
	DepartmentID uint   `json:"department_id"`
	RawRole      string `json:"raw_role"`
	Role         string `json:"role"`
}

type PermissionRequest struct {
	Resource     Resource `json:"resource"`
	Action       Action   `json:"action"`
	Scope        Scope    `json:"scope"`
	OwnerID      uint     `json:"owner_id"`
	DepartmentID uint     `json:"department_id"`
	AssigneeID   uint     `json:"assignee_id"`
	KeyHolderAID uint     `json:"key_holder_a_id"`
	KeyHolderBID uint     `json:"key_holder_b_id"`
}

type Decision struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

func allow() Decision {
	return Decision{Allowed: true}
}

func deny(reason string) Decision {
	return Decision{Allowed: false, Reason: reason}
}

func NormalizeRole(role string) string {
	r := strings.ToLower(strings.TrimSpace(role))
	switch r {
	case "admin":
		return "admin"
	case "leader", "team_leader":
		return "leader"
	case "researcher", "member":
		return "researcher"
	case "procurement", "procurement_specialist":
		return "procurement"
	case "director":
		return "director"
	default:
		return r
	}
}

func ResolveActor(userID uint) (Actor, error) {
	if userID == 0 {
		return Actor{}, errors.New("invalid actor user id")
	}
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return Actor{}, err
	}
	return Actor{
		UserID:       user.ID,
		DepartmentID: user.DepartmentID,
		RawRole:      user.Role,
		Role:         NormalizeRole(user.Role),
	}, nil
}

func Evaluate(actor Actor, req PermissionRequest) Decision {
	if actor.Role == "admin" {
		return allow()
	}

	switch req.Resource {
	case ResourceDispenseRequest:
		return evaluateDispense(actor, req)
	case ResourceProcurementBatch:
		return evaluateProcurementBatch(actor, req)
	case ResourceReagentItem:
		return evaluateReagentItem(actor, req)
	default:
		return deny("unknown resource")
	}
}

func evaluateDispense(actor Actor, req PermissionRequest) Decision {
	switch req.Action {
	case ActionApply:
		if actor.Role != "researcher" {
			return deny("only researcher can apply dispense request")
		}
		if req.Scope == ScopeSelf && req.OwnerID > 0 && req.OwnerID != actor.UserID {
			return deny("self scope mismatch")
		}
		return allow()

	case ActionLeaderApprove:
		if actor.Role != "leader" {
			return deny("only leader can approve")
		}
		if req.Scope == ScopeTeam && req.DepartmentID > 0 && actor.DepartmentID > 0 && req.DepartmentID != actor.DepartmentID {
			return deny("team scope mismatch")
		}
		return allow()

	case ActionKeyHolderConfirm:
		if req.Scope == ScopeAssigned {
			if req.KeyHolderAID == 0 && req.KeyHolderBID == 0 {
				return deny("no assigned key holder")
			}
			if actor.UserID != req.KeyHolderAID && actor.UserID != req.KeyHolderBID {
				return deny("actor is not assigned key holder")
			}
		}
		return allow()
	default:
		return deny("unsupported dispense action")
	}
}

func evaluateProcurementBatch(actor Actor, req PermissionRequest) Decision {
	switch req.Action {
	case ActionConfirmBatch, ActionReceive:
		if actor.Role != "procurement" {
			return deny("only procurement can operate procurement batch")
		}
		return allow()
	default:
		return deny("unsupported procurement action")
	}
}

func evaluateReagentItem(actor Actor, req PermissionRequest) Decision {
	switch req.Action {
	case ActionCheckIn:
		if actor.Role != "researcher" && actor.Role != "leader" {
			return deny("only researcher/leader can check in item")
		}
		if req.Scope == ScopeSelf && req.OwnerID > 0 && req.OwnerID != actor.UserID {
			return deny("self scope mismatch")
		}
		if req.Scope == ScopeAssigned && req.AssigneeID > 0 && req.AssigneeID != actor.UserID {
			return deny("assigned scope mismatch")
		}
		return allow()
	default:
		return deny("unsupported reagent item action")
	}
}
