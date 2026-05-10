package user

import (
	"errors"
	"testing"
	"time"

	domaindivision "activist-base/src/domain/division"
	domainmembership "activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
)

func TestEditProfile(t *testing.T) {
	t.Parallel()

	existing := domainuser.User{
		ID:           shared.UserID("user-1"),
		Login:        "activist_user",
		PasswordHash: "hash",
		FirstName:    "Anna",
	}

	firstName := "Ann"

	t.Run("forbids foreign actor even when permission is granted", func(t *testing.T) {
		t.Parallel()

		_, err := EditProfile(EditProfileInput{
			ActorID:  shared.UserID("user-2"),
			Existing: existing,
			Patch: ProfilePatch{
				FirstName: &firstName,
			},
			Access: permissionAccess(
				t,
				role.CanEditSelfProfile,
				role.ScopeSelf,
				"self",
				"self",
			),
		})

		if !errors.Is(err, shared.ErrForbidden) {
			t.Fatalf("expected forbidden for foreign actor, got %v", err)
		}
	})

	t.Run("forbids self edit without can_edit_self_profile", func(t *testing.T) {
		t.Parallel()

		_, err := EditProfile(EditProfileInput{
			ActorID:  existing.ID,
			Existing: existing,
			Patch: ProfilePatch{
				FirstName: &firstName,
			},
			Access: domainmembership.EffectivePermissions{},
		})

		if !errors.Is(err, shared.ErrForbidden) {
			t.Fatalf("expected forbidden without self-edit permission, got %v", err)
		}
	})

	t.Run("allows self edit when can_edit_self_profile is granted", func(t *testing.T) {
		t.Parallel()

		result, err := EditProfile(EditProfileInput{
			ActorID:  existing.ID,
			Existing: existing,
			Patch: ProfilePatch{
				FirstName: &firstName,
			},
			Access: permissionAccess(
				t,
				role.CanEditSelfProfile,
				role.ScopeSelf,
				"self",
				"self",
			),
		})
		if err != nil {
			t.Fatalf("expected self edit success, got %v", err)
		}
		if result.User.FirstName != firstName {
			t.Fatalf("expected first_name %q, got %q", firstName, result.User.FirstName)
		}
	})
}

func TestReadProfile(t *testing.T) {
	t.Parallel()

	birthDate := time.Date(2004, 3, 14, 0, 0, 0, 0, time.UTC)
	target := domainuser.User{
		ID:              shared.UserID("target-user"),
		Login:           "target_login",
		PasswordHash:    "hash",
		FirstName:       "Ivan",
		GradebookNumber: "GB-123",
		GroupNumber:     "A-11",
		Institute:       "Engineering",
		BirthDate:       &birthDate,
		Phone:           "+79990001122",
		SocialLinks: []shared.Link{
			{Platform: "tg", Value: "t.me/activist"},
		},
		About: "Active member",
	}

	t.Run("hides contact and birth_date for unrelated actor without permissions", func(t *testing.T) {
		t.Parallel()

		result, err := ReadProfile(ReadProfileInput{
			ActorID:    shared.UserID("viewer-1"),
			TargetUser: target,
			Access:     domainmembership.EffectivePermissions{},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		visible := domainuser.VisibleFields(result.Access)
		if visible[domainuser.FieldPhone] {
			t.Fatal("phone must be hidden without can_view_contacts")
		}
		if visible[domainuser.FieldSocialLinks] {
			t.Fatal("social_links must be hidden without can_view_contacts")
		}
		if visible[domainuser.FieldBirthDate] {
			t.Fatal("birth_date must be hidden for unrelated actor")
		}
	})

	t.Run("shows contact fields with can_view_contacts but keeps birth_date hidden", func(t *testing.T) {
		t.Parallel()

		result, err := ReadProfile(ReadProfileInput{
			ActorID:    shared.UserID("viewer-2"),
			TargetUser: target,
			Access: permissionAccess(
				t,
				role.CanViewContacts,
				role.ScopeCurrentDivision,
				"div-1",
				"div-1",
			),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		visible := domainuser.VisibleFields(result.Access)
		if !visible[domainuser.FieldPhone] || !visible[domainuser.FieldSocialLinks] {
			t.Fatal("contact fields must be visible with can_view_contacts")
		}
		if visible[domainuser.FieldBirthDate] {
			t.Fatal("birth_date must stay hidden without self/system_admin")
		}
	})

	t.Run("shows birth_date for self actor", func(t *testing.T) {
		t.Parallel()

		result, err := ReadProfile(ReadProfileInput{
			ActorID:    target.ID,
			TargetUser: target,
			Access:     domainmembership.EffectivePermissions{},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		visible := domainuser.VisibleFields(result.Access)
		if !visible[domainuser.FieldBirthDate] {
			t.Fatal("birth_date must be visible for self")
		}
	})

	t.Run("shows birth_date for system_admin actor", func(t *testing.T) {
		t.Parallel()

		result, err := ReadProfile(ReadProfileInput{
			ActorID:    shared.UserID("admin-1"),
			TargetUser: target,
			Access: permissionAccess(
				t,
				role.SystemAdmin,
				role.ScopeSelf,
				"admin",
				"admin",
			),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		visible := domainuser.VisibleFields(result.Access)
		if !visible[domainuser.FieldBirthDate] {
			t.Fatal("birth_date must be visible for system_admin")
		}
	})

	t.Run("does not infer can_view_contacts from scope_self grant", func(t *testing.T) {
		t.Parallel()

		result, err := ReadProfile(ReadProfileInput{
			ActorID:    target.ID,
			TargetUser: target,
			Access: permissionAccess(
				t,
				role.CanViewContacts,
				role.ScopeSelf,
				"self",
				"self",
			),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Access.CanViewContacts {
			t.Fatal("can_view_contacts must come from resolved ACL grants, not self shortcut")
		}
	})
}

func TestGetMemberships(t *testing.T) {
	t.Parallel()

	target := domainuser.User{
		ID:           shared.UserID("user-1"),
		Login:        "activist_user",
		PasswordHash: "hash",
		FirstName:    "Anna",
	}

	result, err := GetMemberships(GetMembershipsInput{
		ActorID:    shared.UserID("viewer-1"),
		TargetUser: target,
		Items: []MembershipView{
			{
				Membership: domainmembership.Membership{
					UserID:     target.ID,
					PositionID: shared.PositionID("pos-active"),
				},
				Position: domainposition.Position{
					ID:         shared.PositionID("pos-active"),
					DivisionID: shared.DivisionID("div-active"),
				},
				Division: domaindivision.Division{
					ID: shared.DivisionID("div-active"),
				},
			},
			{
				Membership: domainmembership.Membership{
					UserID:     target.ID,
					PositionID: shared.PositionID("pos-archived"),
				},
				Position: domainposition.Position{
					ID:         shared.PositionID("pos-archived"),
					DivisionID: shared.DivisionID("div-active"),
					IsArchived: true,
				},
				Division: domaindivision.Division{
					ID: shared.DivisionID("div-active"),
				},
			},
			{
				Membership: domainmembership.Membership{
					UserID:     target.ID,
					PositionID: shared.PositionID("pos-in-archived-division"),
				},
				Position: domainposition.Position{
					ID:         shared.PositionID("pos-in-archived-division"),
					DivisionID: shared.DivisionID("div-archived"),
				},
				Division: domaindivision.Division{
					ID:         shared.DivisionID("div-archived"),
					IsArchived: true,
				},
			},
			{
				Membership: domainmembership.Membership{
					UserID:     shared.UserID("other-user"),
					PositionID: shared.PositionID("pos-other"),
				},
				Position: domainposition.Position{
					ID:         shared.PositionID("pos-other"),
					DivisionID: shared.DivisionID("div-active"),
				},
				Division: domaindivision.Division{
					ID: shared.DivisionID("div-active"),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 active membership, got %d", len(result.Items))
	}
	if result.Items[0].Membership.PositionID != shared.PositionID("pos-active") {
		t.Fatalf("expected active position only, got %q", result.Items[0].Membership.PositionID)
	}
}

func permissionAccess(
	t *testing.T,
	code role.PermissionCode,
	scope role.ScopeMode,
	sourceDivision shared.DivisionID,
	targetDivision shared.DivisionID,
) domainmembership.EffectivePermissions {
	t.Helper()

	permission, err := role.NewPermission(code, scope)
	if err != nil {
		t.Fatalf("new permission: %v", err)
	}
	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  sourceDivision,
				Permissions: role.NewPermissionSet([]role.Permission{permission}),
			},
		},
		targetDivision,
		nil,
	)
}
