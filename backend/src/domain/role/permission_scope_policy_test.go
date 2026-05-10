package role

import (
	"errors"
	"testing"

	"activist-base/src/domain/shared"
)

func TestCanonicalScopeForPermission(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		code PermissionCode
		want ScopeMode
	}{
		{name: "can_add_member uses CURRENT_DIVISION", code: CanAddMember, want: ScopeCurrentDivision},
		{name: "can_manage_positions uses CURRENT_DIVISION", code: CanManagePositions, want: ScopeCurrentDivision},
		{name: "can_create_subdivision uses CURRENT_AND_DESCENDANTS", code: CanCreateSubdivision, want: ScopeCurrentAndDescendants},
		{name: "can_archive_division uses CURRENT_AND_DESCENDANTS", code: CanArchiveDivision, want: ScopeCurrentAndDescendants},
		{name: "can_edit_self_profile uses SELF", code: CanEditSelfProfile, want: ScopeSelf},
		{name: "can_manage_roles uses CURRENT_AND_DESCENDANTS", code: CanManageRoles, want: ScopeCurrentAndDescendants},
		{name: "can_view_audit_log uses CURRENT_AND_DESCENDANTS", code: CanViewAuditLog, want: ScopeCurrentAndDescendants},
		{name: "system_admin uses CURRENT_AND_DESCENDANTS", code: SystemAdmin, want: ScopeCurrentAndDescendants},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := CanonicalScopeForPermission(tc.code)
			if !ok {
				t.Fatalf("CanonicalScopeForPermission(%v): not found", tc.code)
			}
			if got != tc.want {
				t.Fatalf("CanonicalScopeForPermission(%v) = %v, want %v", tc.code, got, tc.want)
			}
		})
	}
}

func TestCanonicalPermissionSetFromCodes(t *testing.T) {
	t.Parallel()

	set, err := CanonicalPermissionSetFromCodes([]PermissionCode{
		CanManagePositions,
		CanCreateSubdivision,
	})
	if err != nil {
		t.Fatalf("CanonicalPermissionSetFromCodes() error = %v", err)
	}

	if !set.Contains(CanManagePositions, ScopeCurrentDivision) {
		t.Fatal("expected CanManagePositions with ScopeCurrentDivision")
	}
	if !set.Contains(CanCreateSubdivision, ScopeCurrentAndDescendants) {
		t.Fatal("expected CanCreateSubdivision with ScopeCurrentAndDescendants")
	}

	rolesSet, err := CanonicalPermissionSetFromCodes([]PermissionCode{
		CanManageRoles,
	})
	if err != nil {
		t.Fatalf("CanonicalPermissionSetFromCodes(CanManageRoles) error = %v", err)
	}
	if !rolesSet.Contains(CanManageRoles, ScopeCurrentAndDescendants) {
		t.Fatal("expected CanManageRoles with ScopeCurrentAndDescendants")
	}
}

func TestValidatePermissionSetCanonicalRejectsScopeDrift(t *testing.T) {
	t.Parallel()

	drifted, err := NewPermission(CanManagePositions, ScopeCurrentAndDescendants)
	if err != nil {
		t.Fatalf("NewPermission() error = %v", err)
	}
	err = ValidatePermissionSetCanonical(NewPermissionSet([]Permission{drifted}))
	if err == nil {
		t.Fatal("expected validation error for scope drift")
	}

	var validationErr *shared.Error
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected *shared.Error, got %T", err)
	}
	if validationErr.Code != "validation.permissions" {
		t.Fatalf("expected validation.permissions, got %s", validationErr.Code)
	}
}
