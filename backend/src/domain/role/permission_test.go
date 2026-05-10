package role

import "testing"

func TestCanManageRolesIsValid(t *testing.T) {
	t.Parallel()

	if !CanManageRoles.IsValid() {
		t.Fatal("CanManageRoles should be valid")
	}
}

func TestCanManageRolesString(t *testing.T) {
	t.Parallel()

	if got := CanManageRoles.String(); got != "can_manage_roles" {
		t.Fatalf("CanManageRoles.String() = %q, want %q", got, "can_manage_roles")
	}
}

func TestCanViewAuditLogIsValid(t *testing.T) {
	t.Parallel()

	if !CanViewAuditLog.IsValid() {
		t.Fatal("CanViewAuditLog should be valid")
	}
}

func TestCanViewAuditLogString(t *testing.T) {
	t.Parallel()

	if got := CanViewAuditLog.String(); got != "can_view_audit_log" {
		t.Fatalf("CanViewAuditLog.String() = %q, want %q", got, "can_view_audit_log")
	}
}
