package role

import (
	"errors"
	"testing"
	"time"

	domainmembership "activist-base/src/domain/membership"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
)

func TestCreateAllowsCustomScope(t *testing.T) {
	t.Parallel()

	driftedPermission, err := domainrole.NewPermission(domainrole.CanManagePositions, domainrole.ScopeCurrentAndDescendants)
	if err != nil {
		t.Fatalf("NewPermission() error = %v", err)
	}

	result, err := Create(CreateInput{
		ActorID:     "actor-1",
		ID:          "role-1",
		Name:        "Deputy role",
		Permissions: domainrole.NewPermissionSet([]domainrole.Permission{driftedPermission}),
		Access:      systemAdminAccess(t),
		Now:         time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !result.Role.Permissions.Contains(domainrole.CanManagePositions, domainrole.ScopeCurrentAndDescendants) {
		t.Fatal("expected custom scope to be preserved")
	}
}

func TestCreateKeepsCURRENT_DIVISIONAndCURRENT_AND_DESCENDANTSDistinct(t *testing.T) {
	t.Parallel()

	perms, err := domainrole.CanonicalPermissionSetFromCodes([]domainrole.PermissionCode{
		domainrole.CanManagePositions,
		domainrole.CanCreateSubdivision,
	})
	if err != nil {
		t.Fatalf("CanonicalPermissionSetFromCodes() error = %v", err)
	}

	result, err := Create(CreateInput{
		ActorID:     "actor-1",
		ID:          "role-1",
		Name:        "Leader role",
		Permissions: perms,
		Access:      systemAdminAccess(t),
		Now:         time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	effective := domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  "division-parent",
				Permissions: result.Role.Permissions,
			},
		},
		"division-child",
		[]shared.DivisionID{"division-parent"},
	)

	if effective.Has(domainrole.CanManagePositions, false) {
		t.Fatal("expected CURRENT_DIVISION permission to deny child division")
	}
	if !effective.Has(domainrole.CanCreateSubdivision, false) {
		t.Fatal("expected CURRENT_AND_DESCENDANTS permission to allow child division")
	}
}

func TestCreateForbiddenWithoutCanManageRoles(t *testing.T) {
	t.Parallel()

	perms := permissionSetFromCodes(t, domainrole.CanAddMember)
	_, err := Create(CreateInput{
		ActorID:     "actor-1",
		ID:          "role-1",
		Name:        "Role",
		Permissions: perms,
		Access:      domainmembership.EffectivePermissions{},
		Now:         time.Now().UTC(),
	})
	if !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("Create() error = %v, want %v", err, shared.ErrForbidden)
	}
}

func TestCreateForbiddenForNonSystemAdminEvenWithCanManageRoles(t *testing.T) {
	t.Parallel()

	perms := permissionSetFromCodes(t, domainrole.CanManageRoles)
	_, err := Create(CreateInput{
		ActorID:     "actor-1",
		ID:          "role-1",
		Name:        "Role",
		Permissions: perms,
		Access:      canManageRolesAccess(t),
		Now:         time.Now().UTC(),
	})
	if !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("Create() error = %v, want %v", err, shared.ErrForbidden)
	}
}

func TestCreateForbiddenForNonSystemAdmin(t *testing.T) {
	t.Parallel()

	perms := permissionSetFromCodes(t, domainrole.CanViewContacts)
	_, err := Create(CreateInput{
		ActorID:     "actor-1",
		ID:          "role-1",
		Name:        "Role",
		Permissions: perms,
		Access:      canManageRolesAccess(t, domainrole.CanAddMember),
		Now:         time.Now().UTC(),
	})
	if !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("Create() error = %v, want %v", err, shared.ErrForbidden)
	}
}

func TestCreateAllowsSystemAdminPermissionForSystemAdminActor(t *testing.T) {
	t.Parallel()

	perms, err := domainrole.CanonicalPermissionSetFromCodes([]domainrole.PermissionCode{domainrole.SystemAdmin})
	if err != nil {
		t.Fatalf("CanonicalPermissionSetFromCodes() error = %v", err)
	}
	result, err := Create(CreateInput{
		ActorID:     "actor-1",
		ID:          "role-1",
		Name:        "Role",
		Permissions: perms,
		Access:      systemAdminAccess(t),
		Now:         time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !result.Role.Permissions.HasCode(domainrole.SystemAdmin) {
		t.Fatal("expected created role to contain system_admin permission")
	}
}

func TestEditForbiddenWithoutCanManageRoles(t *testing.T) {
	t.Parallel()

	perms := permissionSetFromCodes(t, domainrole.CanAddMember)
	existing := domainrole.Role{
		ID:          "role-1",
		Name:        "Original",
		Permissions: perms,
	}
	_, err := Edit(EditInput{
		ActorID:  "actor-1",
		Existing: existing,
		Patch: EditPatch{
			Permissions: &perms,
		},
		Access: domainmembership.EffectivePermissions{},
		Now:    time.Now().UTC(),
	})
	if !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("Edit() error = %v, want %v", err, shared.ErrForbidden)
	}
}

func TestEditForbiddenForNonSystemAdmin(t *testing.T) {
	t.Parallel()

	basePerms := permissionSetFromCodes(t, domainrole.CanAddMember)
	patchPerms := permissionSetFromCodes(t, domainrole.CanAddMember, domainrole.CanViewContacts)
	existing := domainrole.Role{
		ID:          "role-1",
		Name:        "Original",
		Permissions: basePerms,
	}
	_, err := Edit(EditInput{
		ActorID:  "actor-1",
		Existing: existing,
		Patch: EditPatch{
			Permissions: &patchPerms,
		},
		Access: canManageRolesAccess(t, domainrole.CanAddMember),
		Now:    time.Now().UTC(),
	})
	if !errors.Is(err, shared.ErrForbidden) {
		t.Fatalf("Edit() error = %v, want %v", err, shared.ErrForbidden)
	}
}

func TestEditAllowsSystemAdminToAssignAnyPermissions(t *testing.T) {
	t.Parallel()

	basePerms := permissionSetFromCodes(t, domainrole.CanAddMember)
	patchPerms := permissionSetFromCodes(t, domainrole.CanAddMember, domainrole.CanViewContacts)
	existing := domainrole.Role{
		ID:          "role-1",
		Name:        "Original",
		Permissions: basePerms,
	}
	_, err := Edit(EditInput{
		ActorID:  "actor-1",
		Existing: existing,
		Patch: EditPatch{
			Permissions: &patchPerms,
		},
		Access: systemAdminAccess(t),
		Now:    time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Edit() error = %v", err)
	}
}

func TestSystemAdminBypassesAntiEscalation(t *testing.T) {
	t.Parallel()

	perms := permissionSetFromCodes(t, domainrole.CanManageRoles, domainrole.CanViewContacts)
	_, err := Create(CreateInput{
		ActorID:     "actor-1",
		ID:          "role-1",
		Name:        "Role",
		Permissions: perms,
		Access:      systemAdminAccess(t),
		Now:         time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func permissionSetFromCodes(t *testing.T, codes ...domainrole.PermissionCode) domainrole.PermissionSet {
	t.Helper()

	set, err := domainrole.CanonicalPermissionSetFromCodes(codes)
	if err != nil {
		t.Fatalf("CanonicalPermissionSetFromCodes() error = %v", err)
	}
	return set
}

func canManageRolesAccess(t *testing.T, extraCodes ...domainrole.PermissionCode) domainmembership.EffectivePermissions {
	t.Helper()

	perms := make([]domainrole.Permission, 0, 1+len(extraCodes))
	canManageRolesPermission, err := domainrole.NewPermission(domainrole.CanManageRoles, domainrole.ScopeCurrentAndDescendants)
	if err != nil {
		t.Fatalf("NewPermission(CanManageRoles): %v", err)
	}
	perms = append(perms, canManageRolesPermission)

	for _, code := range extraCodes {
		scope, ok := domainrole.CanonicalScopeForPermission(code)
		if !ok {
			t.Fatalf("no canonical scope for %v", code)
		}
		p, err := domainrole.NewPermission(code, scope)
		if err != nil {
			t.Fatalf("NewPermission(%v): %v", code, err)
		}
		perms = append(perms, p)
	}

	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  "division-root",
				Permissions: domainrole.NewPermissionSet(perms),
			},
		},
		"division-root",
		nil,
	)
}

func systemAdminAccess(t *testing.T) domainmembership.EffectivePermissions {
	t.Helper()

	adminPermission, err := domainrole.NewPermission(domainrole.SystemAdmin, domainrole.ScopeCurrentAndDescendants)
	if err != nil {
		t.Fatalf("NewPermission() error = %v", err)
	}

	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  "division-root",
				Permissions: domainrole.NewPermissionSet([]domainrole.Permission{adminPermission}),
			},
		},
		"division-root",
		nil,
	)
}
