package division

import (
	"errors"
	"testing"
	"time"

	domaindivision "activist-base/src/domain/division"
	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
)

func TestCreateRejectsSecondRoot(t *testing.T) {
	t.Parallel()

	existingRootID := shared.DivisionID("root-1")
	_, err := Create(CreateInput{
		ActorID:        shared.UserID("actor-1"),
		ID:             shared.DivisionID("division-2"),
		ExistingRootID: &existingRootID,
		ShortName:      "Second Root",
		Access:         permissionAccess(t, role.SystemAdmin),
		Now:            time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, domaindivision.ErrRootExists) {
		t.Fatalf("expected ErrRootExists, got %v", err)
	}
}

func TestEditRejectsMoveIntoOwnSubtree(t *testing.T) {
	t.Parallel()

	newParent := domaindivision.Division{
		ID:        shared.DivisionID("division-b"),
		ShortName: "Division B",
	}
	_, err := Edit(EditInput{
		ActorID: shared.UserID("actor-1"),
		Existing: domaindivision.Division{
			ID:        shared.DivisionID("division-a"),
			ShortName: "Division A",
		},
		ReassignParent:     true,
		NewParent:          &newParent,
		NewParentAncestors: []shared.DivisionID{shared.DivisionID("division-a")},
		Access:             permissionAccess(t, role.CanEditDivision),
		Now:                time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, domaindivision.ErrCycleOnMove) {
		t.Fatalf("expected ErrCycleOnMove, got %v", err)
	}
}

func TestCreateRejectsArchivedParent(t *testing.T) {
	t.Parallel()

	archivedParent := domaindivision.Division{
		ID:         shared.DivisionID("parent-1"),
		ShortName:  "Archived",
		IsArchived: true,
	}
	_, err := Create(CreateInput{
		ActorID:   shared.UserID("actor-1"),
		ID:        shared.DivisionID("child-1"),
		Parent:    &archivedParent,
		ShortName: "Child",
		Access:    permissionAccess(t, role.CanCreateSubdivision),
		Now:       time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, domaindivision.ErrArchivedParent) {
		t.Fatalf("expected ErrArchivedParent, got %v", err)
	}
}

func TestCreateAndEditEnforcePermissions(t *testing.T) {
	t.Parallel()

	parent := domaindivision.Division{
		ID:        shared.DivisionID("parent-1"),
		ShortName: "Parent",
	}

	t.Run("create rejects actor without can_create_subdivision", func(t *testing.T) {
		t.Parallel()

		_, err := Create(CreateInput{
			ActorID:   shared.UserID("actor-1"),
			ID:        shared.DivisionID("child-1"),
			Parent:    &parent,
			ShortName: "Child",
			Access:    domainmembership.EffectivePermissions{},
			Now:       time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC),
		})
		if !errors.Is(err, shared.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("edit rejects actor without can_edit_division", func(t *testing.T) {
		t.Parallel()

		_, err := Edit(EditInput{
			ActorID: shared.UserID("actor-1"),
			Existing: domaindivision.Division{
				ID:        shared.DivisionID("division-1"),
				ShortName: "Division 1",
			},
			Patch: EditPatch{
				ShortName: strPtr("Updated"),
			},
			Access: domainmembership.EffectivePermissions{},
			Now:    time.Date(2026, 4, 12, 12, 0, 0, 0, time.UTC),
		})
		if !errors.Is(err, shared.ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestGetTreeAppliesFieldVisibility(t *testing.T) {
	t.Parallel()

	rootID := shared.DivisionID("root-1")
	childID := shared.DivisionID("child-1")
	root := TreeNode{
		Division: domaindivision.Division{
			ID:            rootID,
			ShortName:     "Root",
			FullName:      "Root Full",
			Description:   "Root Description",
			RegulationURL: "https://example.com",
			MediaLinks: []shared.Link{
				{Platform: "site", Value: "https://example.com/media"},
			},
		},
		Children: []TreeNode{
			{
				Division: domaindivision.Division{
					ID:        childID,
					ParentID:  &rootID,
					ShortName: "Child",
					FullName:  "Child Full",
				},
			},
		},
	}

	hidden, err := GetTree(TreeInput{Root: root, Depth: 2, CanViewCard: false})
	if err != nil {
		t.Fatalf("get tree hidden: %v", err)
	}
	if hidden.FullName != nil || hidden.Description != nil || hidden.RegulationURL != nil || hidden.MediaLinks != nil {
		t.Fatal("expected full card fields to be hidden")
	}
	if len(hidden.Children) != 1 {
		t.Fatalf("expected one child, got %d", len(hidden.Children))
	}
	if hidden.Children[0].FullName != nil {
		t.Fatal("expected child full_name to be hidden")
	}

	visible, err := GetTree(TreeInput{Root: root, Depth: 2, CanViewCard: true})
	if err != nil {
		t.Fatalf("get tree visible: %v", err)
	}
	if visible.FullName == nil || *visible.FullName != "Root Full" {
		t.Fatalf("expected visible full_name, got %#v", visible.FullName)
	}
	if visible.Description == nil || *visible.Description != "Root Description" {
		t.Fatalf("expected visible description, got %#v", visible.Description)
	}
	if visible.RegulationURL == nil || *visible.RegulationURL != "https://example.com" {
		t.Fatalf("expected visible regulation_url, got %#v", visible.RegulationURL)
	}
	if len(visible.MediaLinks) != 1 {
		t.Fatalf("expected media_links to be visible, got %d", len(visible.MediaLinks))
	}
}

func permissionAccess(t *testing.T, code role.PermissionCode) domainmembership.EffectivePermissions {
	t.Helper()

	permission, err := role.NewPermission(code, role.ScopeCurrentAndDescendants)
	if err != nil {
		t.Fatalf("new permission: %v", err)
	}
	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  shared.DivisionID("division-1"),
				Permissions: role.NewPermissionSet([]role.Permission{permission}),
			},
		},
		shared.DivisionID("division-1"),
		nil,
	)
}
