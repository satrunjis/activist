package authorization

import (
	"context"
	"testing"

	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
)

func TestResolveForDivisionScopeMatrix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		memberships []domainmembership.MembershipContext
		target      shared.DivisionID
		ancestors   []shared.DivisionID
		expect      bool
	}{
		{
			name: "CURRENT_DIVISION allows exact division",
			memberships: []domainmembership.MembershipContext{
				membershipWithPermission("division-a", role.CanManagePositions, role.ScopeCurrentDivision),
			},
			target:    "division-a",
			ancestors: nil,
			expect:    true,
		},
		{
			name: "CURRENT_AND_DESCENDANTS allows child division",
			memberships: []domainmembership.MembershipContext{
				membershipWithPermission("division-parent", role.CanManagePositions, role.ScopeCurrentAndDescendants),
			},
			target:    "division-child",
			ancestors: []shared.DivisionID{"division-parent"},
			expect:    true,
		},
		{
			name: "CURRENT_DIVISION denies child even when parent is an ancestor",
			memberships: []domainmembership.MembershipContext{
				membershipWithPermission("division-parent", role.CanManagePositions, role.ScopeCurrentDivision),
			},
			target:    "division-child",
			ancestors: []shared.DivisionID{"division-parent"},
			expect:    false,
		},
		{
			name: "cross-division deny for CURRENT_DIVISION",
			memberships: []domainmembership.MembershipContext{
				membershipWithPermission("division-x", role.CanManagePositions, role.ScopeCurrentDivision),
			},
			target:    "division-y",
			ancestors: nil,
			expect:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &stubRepository{
				actorMemberships: tc.memberships,
				ancestorsByDivision: map[shared.DivisionID][]shared.DivisionID{
					tc.target: tc.ancestors,
				},
			}
			svc := NewService(repo)

			access, err := svc.ResolveForDivision(context.Background(), "actor-1", tc.target)
			if err != nil {
				t.Fatalf("ResolveForDivision: %v", err)
			}
			if got := access.Has(role.CanManagePositions, false); got != tc.expect {
				t.Fatalf("access mismatch: got=%v want=%v", got, tc.expect)
			}
		})
	}
}

func TestResolveForUserAndGlobal(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{
		actorMemberships: []domainmembership.MembershipContext{
			membershipWithPermission("division-a", role.CanViewContacts, role.ScopeCurrentDivision),
			membershipWithPermission("division-root", role.CanViewContacts, role.ScopeCurrentAndDescendants),
			membershipWithPermission("division-admin", role.SystemAdmin, role.ScopeCurrentAndDescendants),
		},
		targetDivisionsByUser: map[shared.UserID][]shared.DivisionID{
			"target-user": {"division-a", "division-b"},
		},
		ancestorsByDivision: map[shared.DivisionID][]shared.DivisionID{
			"division-a": {"division-root"},
			"division-b": {"division-root"},
		},
	}
	svc := NewService(repo)

	accessByUser, err := svc.ResolveForUser(context.Background(), "actor-1", "target-user")
	if err != nil {
		t.Fatalf("ResolveForUser: %v", err)
	}
	if !accessByUser.Has(role.CanViewContacts, false) {
		t.Fatal("expected CanViewContacts grant for target user context")
	}

	globalAccess, err := svc.ResolveGlobal(context.Background(), "actor-1")
	if err != nil {
		t.Fatalf("ResolveGlobal: %v", err)
	}
	if !globalAccess.Has(role.CanArchiveDivision, false) {
		t.Fatal("expected global access through system_admin")
	}
}

func TestResolveForUserDeniesOutOfScopeActor(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{
		actorMemberships: []domainmembership.MembershipContext{
			membershipWithPermission("division-a", role.CanViewContacts, role.ScopeCurrentDivision),
		},
		targetDivisionsByUser: map[shared.UserID][]shared.DivisionID{
			"target-user": {"division-b"},
		},
		ancestorsByDivision: map[shared.DivisionID][]shared.DivisionID{
			"division-b": {"division-root"},
		},
	}
	svc := NewService(repo)

	access, err := svc.ResolveForUser(context.Background(), "actor-1", "target-user")
	if err != nil {
		t.Fatalf("ResolveForUser: %v", err)
	}
	if access.Has(role.CanViewContacts, false) {
		t.Fatal("expected out-of-scope actor to be denied")
	}
}

func TestResolveForUserAllowsSystemAdminWithoutTargetDivisions(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{
		actorMemberships: []domainmembership.MembershipContext{
			membershipWithPermission("division-admin", role.SystemAdmin, role.ScopeCurrentAndDescendants),
		},
		targetDivisionsByUser: map[shared.UserID][]shared.DivisionID{
			"target-user": {},
		},
	}
	svc := NewService(repo)

	access, err := svc.ResolveForUser(context.Background(), "actor-1", "target-user")
	if err != nil {
		t.Fatalf("ResolveForUser: %v", err)
	}
	if !access.Has(role.SystemAdmin, false) {
		t.Fatal("expected system_admin grant for target user without active divisions")
	}
}

func membershipWithPermission(divisionID shared.DivisionID, code role.PermissionCode, scope role.ScopeMode) domainmembership.MembershipContext {
	permission, err := role.NewPermission(code, scope)
	if err != nil {
		panic(err)
	}
	return domainmembership.MembershipContext{
		DivisionID:  divisionID,
		Permissions: role.NewPermissionSet([]role.Permission{permission}),
	}
}

type stubRepository struct {
	actorMemberships      []domainmembership.MembershipContext
	ancestorsByDivision   map[shared.DivisionID][]shared.DivisionID
	targetDivisionsByUser map[shared.UserID][]shared.DivisionID
}

func (s *stubRepository) ListActorMembershipContexts(_ context.Context, _ shared.UserID) ([]domainmembership.MembershipContext, error) {
	return append([]domainmembership.MembershipContext(nil), s.actorMemberships...), nil
}

func (s *stubRepository) ListTargetDivisionAncestors(_ context.Context, targetDivisionID shared.DivisionID) ([]shared.DivisionID, error) {
	ancestors := s.ancestorsByDivision[targetDivisionID]
	return append([]shared.DivisionID(nil), ancestors...), nil
}

func (s *stubRepository) ListTargetUserDivisionIDs(_ context.Context, targetUserID shared.UserID) ([]shared.DivisionID, error) {
	ids := s.targetDivisionsByUser[targetUserID]
	return append([]shared.DivisionID(nil), ids...), nil
}
