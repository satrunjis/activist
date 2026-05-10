package membership

import (
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
)

// MembershipContext carries the permissions a user holds in one division.
type MembershipContext struct {
	DivisionID  shared.DivisionID
	Permissions role.PermissionSet
}

// EffectivePermissions is the resolved permission set for a user acting on a target division.
type EffectivePermissions struct {
	grants map[role.PermissionCode]role.ScopeMode
}

// Calculate resolves what the actor can actually do in targetDivisionID,
// given all their memberships and the ancestor chain of the target division.
func Calculate(
	memberships []MembershipContext,
	targetDivisionID shared.DivisionID,
	ancestorDivisionIDs []shared.DivisionID,
) EffectivePermissions {
	grants := make(map[role.PermissionCode]role.ScopeMode)
	for _, m := range memberships {
		isTarget := m.DivisionID == targetDivisionID
		isAncestor := containsID(ancestorDivisionIDs, m.DivisionID)

		for _, p := range m.Permissions.Permissions() {
			switch {
			case p.Code().IsSystemAdmin():
				mergeGrant(grants, p.Code(), p.Scope())
			case isTarget:
				mergeGrant(grants, p.Code(), p.Scope())
			case isAncestor && p.Scope() == role.ScopeCurrentAndDescendants:
				mergeGrant(grants, p.Code(), p.Scope())
			}
		}
	}
	return EffectivePermissions{grants: grants}
}

// Has reports whether the actor holds code in this context.
// passSelfCheck must be true when the target object belongs to the actor themselves
// (used for ScopeSelf permissions like CanEditSelfProfile).
func (ep EffectivePermissions) Has(code role.PermissionCode, passSelfCheck bool) bool {
	_, ok := ep.GrantScope(code, passSelfCheck)
	return ok
}

// GrantScope returns the strongest scope the actor has for a permission code in
// the current target context. The boolean is false when the permission is not
// effectively granted in this context.
func (ep EffectivePermissions) GrantScope(code role.PermissionCode, passSelfCheck bool) (role.ScopeMode, bool) {
	if _, ok := ep.grants[role.SystemAdmin]; ok {
		return role.ScopeCurrentAndDescendants, true
	}
	scope, ok := ep.grants[code]
	if !ok {
		return 0, false
	}
	if scope == role.ScopeSelf && !passSelfCheck {
		return 0, false
	}
	return scope, true
}

func containsID(ids []shared.DivisionID, target shared.DivisionID) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}

func mergeGrant(grants map[role.PermissionCode]role.ScopeMode, code role.PermissionCode, scope role.ScopeMode) {
	if existing, ok := grants[code]; !ok || scope > existing {
		grants[code] = scope
	}
}
