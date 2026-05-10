package role

import (
	"fmt"

	"activist-base/src/domain/shared"
)

var canonicalPermissionScopes = map[PermissionCode]ScopeMode{
	CanAddMember:         ScopeCurrentDivision,       // CURRENT_DIVISION
	CanRemoveMember:      ScopeCurrentDivision,       // CURRENT_DIVISION
	CanAssignPosition:    ScopeCurrentDivision,       // CURRENT_DIVISION
	CanEditDivision:      ScopeCurrentDivision,       // CURRENT_DIVISION
	CanManagePositions:   ScopeCurrentDivision,       // CURRENT_DIVISION
	CanCreateSubdivision: ScopeCurrentAndDescendants, // CURRENT_AND_DESCENDANTS
	CanArchiveDivision:   ScopeCurrentAndDescendants, // CURRENT_AND_DESCENDANTS
	CanViewContacts:      ScopeCurrentDivision,       // CURRENT_DIVISION
	CanEditSelfProfile:   ScopeSelf,                  // SELF
	CanManageRoles:       ScopeCurrentAndDescendants, // global — roles have no division scope
	CanViewAuditLog:      ScopeCurrentAndDescendants, // global — audit log has no division scope
	SystemAdmin:          ScopeCurrentAndDescendants, // CURRENT_AND_DESCENDANTS
}

func CanonicalScopeForPermission(code PermissionCode) (ScopeMode, bool) {
	scope, ok := canonicalPermissionScopes[code]
	return scope, ok
}

func CanonicalPermissionSetFromCodes(codes []PermissionCode) (PermissionSet, error) {
	if len(codes) == 0 {
		return PermissionSet{}, nil
	}

	perms := make([]Permission, 0, len(codes))
	for _, code := range codes {
		scope, ok := CanonicalScopeForPermission(code)
		if !ok {
			return PermissionSet{}, validationPermissionError("unknown permission code")
		}
		perm, err := NewPermission(code, scope)
		if err != nil {
			return PermissionSet{}, err
		}
		perms = append(perms, perm)
	}
	return NewPermissionSet(perms), nil
}

func ValidatePermissionSetCanonical(set PermissionSet) error {
	for _, perm := range set.Permissions() {
		expected, ok := CanonicalScopeForPermission(perm.Code())
		if !ok {
			return validationPermissionError("unknown permission code")
		}
		if perm.Scope() != expected {
			return validationPermissionError(fmt.Sprintf(
				"scope drift for %s: expected %s, got %s",
				perm.Code().String(),
				scopeLabel(expected),
				scopeLabel(perm.Scope()),
			))
		}
	}
	return nil
}

func validationPermissionError(message string) error {
	return &shared.Error{
		Code:    "validation.permissions",
		Message: message,
	}
}

func scopeLabel(scope ScopeMode) string {
	switch scope {
	case ScopeSelf:
		return "SELF"
	case ScopeCurrentDivision:
		return "CURRENT_DIVISION"
	case ScopeCurrentAndDescendants:
		return "CURRENT_AND_DESCENDANTS"
	default:
		return "UNKNOWN"
	}
}
