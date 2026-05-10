package postgres_test

import (
	"context"
	"testing"

	"activist-base/src/domain/shared"
	"activist-base/src/repository/postgres"
	"activist-base/src/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

// seedDivisionTree creates a root -> child-A -> grandchild hierarchy plus an
// archived sibling of child-A. The archived node must NOT appear in has_children
// computations.
func seedDivisionTree(t *testing.T, ctx context.Context, store *postgres.Store) {
	t.Helper()

	createDivision := func(id, parentID, shortName string, archived bool) {
		t.Helper()
		parent := pgtype.Text{}
		if parentID != "" {
			parent = pgtype.Text{String: parentID, Valid: true}
		}
		if _, err := store.Queries.CreateDivision(ctx, sqlc.CreateDivisionParams{
			ID:            id,
			ParentID:      parent,
			ShortName:     shortName,
			FullName:      shortName + " Full",
			Description:   "",
			RegulationUrl: "",
			MediaLinks:    []byte(`[]`),
			IsArchived:    archived,
		}); err != nil {
			t.Fatalf("create division %q: %v", id, err)
		}
	}

	// root (no parent)
	createDivision("root-1", "", "Root", false)
	// two active children of root
	createDivision("child-a", "root-1", "Alpha", false)
	createDivision("child-b", "root-1", "Beta", false)
	// archived child — must not count toward has_children of root
	createDivision("child-archived", "root-1", "Archived", true)
	// grandchild of child-a
	createDivision("grand-1", "child-a", "GrandOne", false)
}

// TestDivisionRepository_ListChildren_DirectChildrenOnly verifies that only
// direct children of the requested parent are returned, ordered by short_name.
func TestDivisionRepository_ListChildren_DirectChildrenOnly(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t, ctx)
	seedDivisionTree(t, ctx, store)

	repo := postgres.NewDivisionRepository(store.Queries)

	items, err := repo.ListChildren(ctx, shared.DivisionID("root-1"))
	if err != nil {
		t.Fatalf("ListChildren: %v", err)
	}

	// archived child must not appear
	for _, item := range items {
		if item.Division.IsArchived {
			t.Errorf("ListChildren returned archived division %q", item.Division.ID)
		}
	}

	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, string(item.Division.ID))
	}

	// Expect exactly alpha and beta (not archived, not grandchild)
	if len(items) != 2 {
		t.Fatalf("expected 2 direct children, got %d: %v", len(items), ids)
	}
	// Ordered by short_name: Alpha before Beta
	if string(items[0].Division.ID) != "child-a" {
		t.Errorf("expected child-a first (Alpha), got %q", items[0].Division.ID)
	}
	if string(items[1].Division.ID) != "child-b" {
		t.Errorf("expected child-b second (Beta), got %q", items[1].Division.ID)
	}
}

// TestDivisionRepository_ListChildren_HasChildrenMetadata verifies that
// has_children is true only for nodes that have at least one non-archived child.
func TestDivisionRepository_ListChildren_HasChildrenMetadata(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t, ctx)
	seedDivisionTree(t, ctx, store)

	repo := postgres.NewDivisionRepository(store.Queries)

	items, err := repo.ListChildren(ctx, shared.DivisionID("root-1"))
	if err != nil {
		t.Fatalf("ListChildren: %v", err)
	}

	byID := make(map[string]postgres.DivisionChildItem, len(items))
	for _, item := range items {
		byID[string(item.Division.ID)] = item
	}

	// child-a has grand-1 => has_children true
	if !byID["child-a"].HasChildren {
		t.Error("expected child-a has_children=true")
	}
	if byID["child-a"].ChildrenCount != 1 {
		t.Errorf("expected child-a children_count=1, got %d", byID["child-a"].ChildrenCount)
	}

	// child-b has no children
	if byID["child-b"].HasChildren {
		t.Error("expected child-b has_children=false")
	}
	if byID["child-b"].ChildrenCount != 0 {
		t.Errorf("expected child-b children_count=0, got %d", byID["child-b"].ChildrenCount)
	}
}

// TestDivisionRepository_ListRootChildren_RootMode verifies that ListRootChildren
// returns root-level nodes with correct metadata.
func TestDivisionRepository_ListRootChildren_RootMode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := newTestStore(t, ctx)
	seedDivisionTree(t, ctx, store)

	repo := postgres.NewDivisionRepository(store.Queries)

	roots, err := repo.ListRootChildren(ctx)
	if err != nil {
		t.Fatalf("ListRootChildren: %v", err)
	}

	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	root := roots[0]
	if string(root.Division.ID) != "root-1" {
		t.Errorf("expected root-1, got %q", root.Division.ID)
	}
	// root has 2 non-archived children (child-a, child-b)
	if !root.HasChildren {
		t.Error("expected root has_children=true")
	}
	if root.ChildrenCount != 2 {
		t.Errorf("expected root children_count=2, got %d", root.ChildrenCount)
	}
}
