package search

import (
	"context"
	"errors"
	"testing"

	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
)

func TestSearchUsersMasksContactsWithoutPermission(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{
		result: SearchUsersResult{
			Items: []SearchUser{
				{
					ID:          "target-1",
					FirstName:   "Alice",
					Phone:       stringPtr("+79990000000"),
					SocialLinks: linksPtr([]shared.Link{{Platform: "tg", Value: "https://t.me/a"}}),
					About:       stringPtr("about"),
				},
			},
			Total: 1,
		},
	}
	authz := &stubAuthorizationService{
		byTarget: map[shared.UserID]domainmembership.EffectivePermissions{
			"target-1": {},
		},
	}

	svc := NewService(repo, authz)
	result, err := svc.SearchUsers(context.Background(), "actor-1", SearchUsersInput{})
	if err != nil {
		t.Fatalf("SearchUsers returned error: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].Phone != nil {
		t.Fatalf("expected phone to be masked")
	}
	if result.Items[0].SocialLinks != nil {
		t.Fatalf("expected social links to be masked")
	}
	if result.Items[0].About != nil {
		t.Fatalf("expected about to be masked")
	}
}

func TestSearchUsersKeepsContactsWithViewContacts(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{
		result: SearchUsersResult{
			Items: []SearchUser{
				{
					ID:          "target-1",
					FirstName:   "Alice",
					Phone:       stringPtr("+79990000000"),
					SocialLinks: linksPtr([]shared.Link{{Platform: "tg", Value: "https://t.me/a"}}),
					About:       stringPtr("about"),
				},
			},
			Total: 1,
		},
	}
	authz := &stubAuthorizationService{
		byTarget: map[shared.UserID]domainmembership.EffectivePermissions{
			"target-1": mustEffective(role.CanViewContacts),
		},
	}

	svc := NewService(repo, authz)
	result, err := svc.SearchUsers(context.Background(), "actor-1", SearchUsersInput{})
	if err != nil {
		t.Fatalf("SearchUsers returned error: %v", err)
	}

	if result.Items[0].Phone == nil {
		t.Fatalf("expected phone to be visible")
	}
	if result.Items[0].SocialLinks == nil {
		t.Fatalf("expected social links to be visible")
	}
	if result.Items[0].About == nil {
		t.Fatalf("expected about to be visible")
	}
}

func TestSearchUsersDefaultPagination(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{}
	svc := NewService(repo, &stubAuthorizationService{})

	_, err := svc.SearchUsers(context.Background(), "actor-1", SearchUsersInput{})
	if err != nil {
		t.Fatalf("SearchUsers returned error: %v", err)
	}

	if repo.params.Limit != 20 {
		t.Fatalf("expected default limit=20, got %d", repo.params.Limit)
	}
	if repo.params.Offset != 0 {
		t.Fatalf("expected default offset=0, got %d", repo.params.Offset)
	}
}

func TestSearchUsersLimitUpperBound(t *testing.T) {
	t.Parallel()

	svc := NewService(&stubRepository{}, &stubAuthorizationService{})
	_, err := svc.SearchUsers(context.Background(), "actor-1", SearchUsersInput{
		Limit: 51,
	})
	if err == nil {
		t.Fatalf("expected validation error")
	}

	var domainErr *shared.Error
	if !errors.As(err, &domainErr) {
		t.Fatalf("expected shared.Error, got %T", err)
	}
	if domainErr.Code != "validation.limit" {
		t.Fatalf("expected validation.limit, got %s", domainErr.Code)
	}
}

func TestSearchUsersPassesRolePositionFilters(t *testing.T) {
	t.Parallel()

	repo := &stubRepository{}
	svc := NewService(repo, &stubAuthorizationService{})

	_, err := svc.SearchUsers(context.Background(), "actor-1", SearchUsersInput{
		PositionTitle: "Lead",
		RoleName:      "Coordinator",
	})
	if err != nil {
		t.Fatalf("SearchUsers returned error: %v", err)
	}
	if repo.params.PositionTitle != "Lead" {
		t.Fatalf("expected position_title forwarded, got %q", repo.params.PositionTitle)
	}
	if repo.params.RoleName != "Coordinator" {
		t.Fatalf("expected role_name forwarded, got %q", repo.params.RoleName)
	}
}

type stubRepository struct {
	params SearchUsersParams
	result SearchUsersResult
	err    error
}

func (s *stubRepository) SearchUsers(_ context.Context, params SearchUsersParams) (SearchUsersResult, error) {
	s.params = params
	if s.err != nil {
		return SearchUsersResult{}, s.err
	}
	return s.result, nil
}

type stubAuthorizationService struct {
	byTarget map[shared.UserID]domainmembership.EffectivePermissions
	err      error
}

func (s *stubAuthorizationService) ResolveForUser(
	_ context.Context,
	_ shared.UserID,
	targetUserID shared.UserID,
) (domainmembership.EffectivePermissions, error) {
	if s.err != nil {
		return domainmembership.EffectivePermissions{}, s.err
	}
	if s.byTarget == nil {
		return domainmembership.EffectivePermissions{}, nil
	}
	return s.byTarget[targetUserID], nil
}

func mustEffective(code role.PermissionCode) domainmembership.EffectivePermissions {
	perm, err := role.NewPermission(code, role.ScopeCurrentDivision)
	if err != nil {
		panic(err)
	}
	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  "division-1",
				Permissions: role.NewPermissionSet([]role.Permission{perm}),
			},
		},
		"division-1",
		nil,
	)
}

func stringPtr(value string) *string {
	return &value
}

func linksPtr(values []shared.Link) *[]shared.Link {
	copied := append([]shared.Link(nil), values...)
	return &copied
}
