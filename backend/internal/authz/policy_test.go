package authz

import "testing"

func TestNormalizeRole(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"team_leader", "leader"},
		{"member", "researcher"},
		{"procurement_specialist", "procurement"},
		{"admin", "admin"},
	}

	for _, tc := range tests {
		if got := NormalizeRole(tc.in); got != tc.want {
			t.Fatalf("NormalizeRole(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestEvaluateDispenseApply(t *testing.T) {
	researcher := Actor{UserID: 10, DepartmentID: 2, Role: "researcher"}
	leader := Actor{UserID: 11, DepartmentID: 2, Role: "leader"}

	allowReq := PermissionRequest{
		Resource: ResourceDispenseRequest,
		Action:   ActionApply,
		Scope:    ScopeSelf,
		OwnerID:  10,
	}
	if d := Evaluate(researcher, allowReq); !d.Allowed {
		t.Fatalf("expected researcher apply allow, got deny: %s", d.Reason)
	}

	denyReq := PermissionRequest{
		Resource: ResourceDispenseRequest,
		Action:   ActionApply,
		Scope:    ScopeSelf,
		OwnerID:  99,
	}
	if d := Evaluate(researcher, denyReq); d.Allowed {
		t.Fatal("expected researcher apply deny for self scope mismatch")
	}

	if d := Evaluate(leader, allowReq); d.Allowed {
		t.Fatal("expected leader apply deny")
	}
}

func TestEvaluateLeaderApproveTeamScope(t *testing.T) {
	leader := Actor{UserID: 11, DepartmentID: 2, Role: "leader"}

	allowReq := PermissionRequest{
		Resource:     ResourceDispenseRequest,
		Action:       ActionLeaderApprove,
		Scope:        ScopeTeam,
		DepartmentID: 2,
	}
	if d := Evaluate(leader, allowReq); !d.Allowed {
		t.Fatalf("expected leader team approve allow, got deny: %s", d.Reason)
	}

	denyReq := PermissionRequest{
		Resource:     ResourceDispenseRequest,
		Action:       ActionLeaderApprove,
		Scope:        ScopeTeam,
		DepartmentID: 3,
	}
	if d := Evaluate(leader, denyReq); d.Allowed {
		t.Fatal("expected leader team approve deny on department mismatch")
	}
}

func TestEvaluateAssignedKeyHolder(t *testing.T) {
	actorA := Actor{UserID: 21, Role: "procurement"}
	actorOther := Actor{UserID: 99, Role: "researcher"}

	req := PermissionRequest{
		Resource:     ResourceDispenseRequest,
		Action:       ActionKeyHolderConfirm,
		Scope:        ScopeAssigned,
		KeyHolderAID: 21,
		KeyHolderBID: 22,
	}

	if d := Evaluate(actorA, req); !d.Allowed {
		t.Fatalf("expected assigned key holder allow, got deny: %s", d.Reason)
	}
	if d := Evaluate(actorOther, req); d.Allowed {
		t.Fatal("expected unassigned actor deny")
	}
}

func TestEvaluateProcurementBatch(t *testing.T) {
	procurement := Actor{UserID: 31, Role: "procurement"}
	researcher := Actor{UserID: 32, Role: "researcher"}

	req := PermissionRequest{
		Resource: ResourceProcurementBatch,
		Action:   ActionConfirmBatch,
		Scope:    ScopeGlobal,
	}

	if d := Evaluate(procurement, req); !d.Allowed {
		t.Fatalf("expected procurement allow, got deny: %s", d.Reason)
	}
	if d := Evaluate(researcher, req); d.Allowed {
		t.Fatal("expected non-procurement deny")
	}
}
