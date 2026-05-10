package role

import (
	"errors"
	"testing"

	"activist-base/src/domain/shared"
)

func TestRoleValidateUserManagedRejectsSystemAdminPermission(t *testing.T) {
	t.Parallel()

	adminPerm, err := NewPermission(SystemAdmin, ScopeCurrentDivision)
	if err != nil {
		t.Fatalf("NewPermission() error = %v", err)
	}

	r := Role{
		ID:          shared.RoleID("role_1"),
		Name:        "Role",
		Permissions: NewPermissionSet([]Permission{adminPerm}),
	}

	err = r.ValidateUserManaged()
	if !errors.Is(err, ErrSystemAdminInRole) {
		t.Fatalf("ValidateUserManaged() error = %v, want %v", err, ErrSystemAdminInRole)
	}
}
