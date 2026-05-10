package authorization

import (
	"context"

	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
)

type Repository interface {
	ListActorMembershipContexts(ctx context.Context, actorID shared.UserID) ([]domainmembership.MembershipContext, error)
	ListTargetDivisionAncestors(ctx context.Context, targetDivisionID shared.DivisionID) ([]shared.DivisionID, error)
	ListTargetUserDivisionIDs(ctx context.Context, targetUserID shared.UserID) ([]shared.DivisionID, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ResolveForDivision(
	ctx context.Context,
	actorID shared.UserID,
	targetDivisionID shared.DivisionID,
) (domainmembership.EffectivePermissions, error) {
	memberships, err := s.repo.ListActorMembershipContexts(ctx, actorID)
	if err != nil {
		return domainmembership.EffectivePermissions{}, err
	}
	ancestors, err := s.repo.ListTargetDivisionAncestors(ctx, targetDivisionID)
	if err != nil {
		return domainmembership.EffectivePermissions{}, err
	}
	return domainmembership.Calculate(memberships, targetDivisionID, ancestors), nil
}

func (s *Service) ResolveForUser(
	ctx context.Context,
	actorID shared.UserID,
	targetUserID shared.UserID,
) (domainmembership.EffectivePermissions, error) {
	memberships, err := s.repo.ListActorMembershipContexts(ctx, actorID)
	if err != nil {
		return domainmembership.EffectivePermissions{}, err
	}
	for _, membership := range memberships {
		for _, permission := range membership.Permissions.Permissions() {
			if permission.Code().IsSystemAdmin() {
				return effectiveFromGrants(map[role.PermissionCode]role.ScopeMode{
					permission.Code(): permission.Scope(),
				}), nil
			}
		}
	}
	targetDivisionIDs, err := s.repo.ListTargetUserDivisionIDs(ctx, targetUserID)
	if err != nil {
		return domainmembership.EffectivePermissions{}, err
	}

	merged := make(map[role.PermissionCode]role.ScopeMode)
	for _, targetDivisionID := range targetDivisionIDs {
		ancestors, err := s.repo.ListTargetDivisionAncestors(ctx, targetDivisionID)
		if err != nil {
			return domainmembership.EffectivePermissions{}, err
		}

		// Keep the resolver contract explicit: per-target permission resolution
		// always goes through domainmembership.Calculate.
		_ = domainmembership.Calculate(memberships, targetDivisionID, ancestors)

		for code, scope := range grantsForTarget(memberships, targetDivisionID, ancestors) {
			mergeGrant(merged, code, scope)
		}
	}

	return effectiveFromGrants(merged), nil
}

func (s *Service) ResolveGlobal(
	ctx context.Context,
	actorID shared.UserID,
) (domainmembership.EffectivePermissions, error) {
	memberships, err := s.repo.ListActorMembershipContexts(ctx, actorID)
	if err != nil {
		return domainmembership.EffectivePermissions{}, err
	}

	merged := make(map[role.PermissionCode]role.ScopeMode)
	for _, membership := range memberships {
		for _, permission := range membership.Permissions.Permissions() {
			mergeGrant(merged, permission.Code(), permission.Scope())
		}
	}
	return effectiveFromGrants(merged), nil
}

func grantsForTarget(
	memberships []domainmembership.MembershipContext,
	targetDivisionID shared.DivisionID,
	ancestorDivisionIDs []shared.DivisionID,
) map[role.PermissionCode]role.ScopeMode {
	grants := make(map[role.PermissionCode]role.ScopeMode)

	for _, membership := range memberships {
		isTarget := membership.DivisionID == targetDivisionID
		isAncestor := containsDivisionID(ancestorDivisionIDs, membership.DivisionID)
		for _, permission := range membership.Permissions.Permissions() {
			switch {
			case permission.Code().IsSystemAdmin():
				mergeGrant(grants, permission.Code(), permission.Scope())
			case isTarget:
				mergeGrant(grants, permission.Code(), permission.Scope())
			case isAncestor && permission.Scope() == role.ScopeCurrentAndDescendants:
				mergeGrant(grants, permission.Code(), permission.Scope())
			}
		}
	}

	return grants
}

func effectiveFromGrants(grants map[role.PermissionCode]role.ScopeMode) domainmembership.EffectivePermissions {
	if len(grants) == 0 {
		return domainmembership.EffectivePermissions{}
	}

	permissions := make([]role.Permission, 0, len(grants))
	for code, scope := range grants {
		permission, err := role.NewPermission(code, scope)
		if err != nil {
			continue
		}
		permissions = append(permissions, permission)
	}
	if len(permissions) == 0 {
		return domainmembership.EffectivePermissions{}
	}

	const mergedDivisionID shared.DivisionID = "__merged_target__"
	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  mergedDivisionID,
				Permissions: role.NewPermissionSet(permissions),
			},
		},
		mergedDivisionID,
		nil,
	)
}

func containsDivisionID(ids []shared.DivisionID, target shared.DivisionID) bool {
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
