package postgres_test

import (
	"context"
	"testing"
	"time"

	"activist-base/src/repository/postgres"
	"activist-base/src/repository/postgres/sqlc"
	testpostgres "activist-base/src/testutil/postgres"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestSearchUsersCombinationFilters(t *testing.T) {
	ctx := context.Background()
	store := newSearchTestStore(t, ctx)
	seedSearchFixtures(t, ctx, store)

	repo := postgres.NewSearchRepository(store.Queries)
	result, err := repo.SearchUsers(ctx, postgres.SearchUsersParams{
		Login:           "combo_login",
		Institute:       "Engineering",
		About:           "organizing",
		IncludeArchived: false,
		Limit:           10,
		Offset:          0,
	})
	if err != nil {
		t.Fatalf("search users: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total=1, got %d", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result.Items))
	}
	if result.Items[0].ID != "user-1" {
		t.Fatalf("expected user-1, got %s", result.Items[0].ID)
	}
}

func TestSearchUsersByRoleAndPosition(t *testing.T) {
	ctx := context.Background()
	store := newSearchTestStore(t, ctx)
	seedSearchFixtures(t, ctx, store)

	repo := postgres.NewSearchRepository(store.Queries)
	result, err := repo.SearchUsers(ctx, postgres.SearchUsersParams{
		PositionTitle:   "Team Lead",
		RoleName:        "Coordinator",
		IncludeArchived: false,
		Limit:           10,
		Offset:          0,
	})
	if err != nil {
		t.Fatalf("search users: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total=1, got %d", result.Total)
	}
	if len(result.Items) != 1 || result.Items[0].ID != "user-1" {
		t.Fatalf("expected active assignment user user-1, got %+v", result.Items)
	}
}

func TestSearchUsersExcludesArchivedAssignmentsByDefault(t *testing.T) {
	ctx := context.Background()
	store := newSearchTestStore(t, ctx)
	seedSearchFixtures(t, ctx, store)

	repo := postgres.NewSearchRepository(store.Queries)
	result, err := repo.SearchUsers(ctx, postgres.SearchUsersParams{
		Login:           "archived_login",
		IncludeArchived: false,
		Limit:           10,
		Offset:          0,
	})
	if err != nil {
		t.Fatalf("search users: %v", err)
	}

	if result.Total != 0 {
		t.Fatalf("expected total=0, got %d", result.Total)
	}
	if len(result.Items) != 0 {
		t.Fatalf("expected no items, got %d", len(result.Items))
	}
}

func TestSearchUsersIncludeArchived(t *testing.T) {
	ctx := context.Background()
	store := newSearchTestStore(t, ctx)
	seedSearchFixtures(t, ctx, store)

	repo := postgres.NewSearchRepository(store.Queries)
	result, err := repo.SearchUsers(ctx, postgres.SearchUsersParams{
		Login:           "archived_login",
		IncludeArchived: true,
		Limit:           10,
		Offset:          0,
	})
	if err != nil {
		t.Fatalf("search users: %v", err)
	}

	if result.Total != 1 {
		t.Fatalf("expected total=1, got %d", result.Total)
	}
	if len(result.Items) != 1 || result.Items[0].ID != "user-3" {
		t.Fatalf("expected archived assignment user user-3, got %+v", result.Items)
	}
}

func TestSearchUsersPagination(t *testing.T) {
	ctx := context.Background()
	store := newSearchTestStore(t, ctx)
	seedSearchFixtures(t, ctx, store)

	repo := postgres.NewSearchRepository(store.Queries)
	result, err := repo.SearchUsers(ctx, postgres.SearchUsersParams{
		IncludeArchived: false,
		Limit:           1,
		Offset:          1,
	})
	if err != nil {
		t.Fatalf("search users: %v", err)
	}

	if result.Total != 3 {
		t.Fatalf("expected total=3, got %d", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("expected one paginated item, got %d", len(result.Items))
	}
	if result.Items[0].ID != "user-2" {
		t.Fatalf("expected second ordered user user-2, got %s", result.Items[0].ID)
	}
}

func newSearchTestStore(t *testing.T, ctx context.Context) *postgres.Store {
	t.Helper()

	tdb, cleanupDB, err := testpostgres.NewTestDatabase(ctx)
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}
	t.Cleanup(cleanupDB)

	store, err := postgres.New(ctx, tdb.DSN)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(store.Close)

	return store
}

func seedSearchFixtures(t *testing.T, ctx context.Context, store *postgres.Store) {
	t.Helper()

	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}

	if _, err := store.Queries.CreateRole(ctx, sqlc.CreateRoleParams{
		ID:          "role-active",
		Name:        "Coordinator",
		Permissions: []byte{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("create active role: %v", err)
	}
	if _, err := store.Queries.CreateRole(ctx, sqlc.CreateRoleParams{
		ID:          "role-archived",
		Name:        "Former Coordinator",
		Permissions: []byte{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("create archived role: %v", err)
	}

	if _, err := store.Queries.CreateDivision(ctx, sqlc.CreateDivisionParams{
		ID:            "div-active",
		ParentID:      pgtype.Text{},
		ShortName:     "ACT",
		FullName:      "Active Division",
		Description:   "",
		RegulationUrl: "",
		MediaLinks:    []byte(`[]`),
		IsArchived:    false,
	}); err != nil {
		t.Fatalf("create active division: %v", err)
	}
	if _, err := store.Queries.CreateDivision(ctx, sqlc.CreateDivisionParams{
		ID:            "div-archived",
		ParentID:      toText("div-active"),
		ShortName:     "ARC",
		FullName:      "Archived Division",
		Description:   "",
		RegulationUrl: "",
		MediaLinks:    []byte(`[]`),
		IsArchived:    true,
	}); err != nil {
		t.Fatalf("create archived division: %v", err)
	}

	if _, err := store.Queries.CreatePosition(ctx, sqlc.CreatePositionParams{
		ID:         "pos-active",
		Title:      "Team Lead",
		RoleID:     "role-active",
		DivisionID: "div-active",
		MaxCount:   pgtype.Int4{Int32: 10, Valid: true},
	}); err != nil {
		t.Fatalf("create active position: %v", err)
	}
	if _, err := store.Queries.CreatePosition(ctx, sqlc.CreatePositionParams{
		ID:         "pos-archived",
		Title:      "Team Lead",
		RoleID:     "role-archived",
		DivisionID: "div-archived",
		MaxCount:   pgtype.Int4{Int32: 10, Valid: true},
	}); err != nil {
		t.Fatalf("create archived position: %v", err)
	}
	if _, err := store.Queries.ArchivePosition(ctx, "pos-archived"); err != nil {
		t.Fatalf("archive position: %v", err)
	}

	createUser := func(
		id, login, firstName, lastName, institute, about, groupNumber, phone, gradebook string,
	) {
		t.Helper()
		_, err := store.Queries.CreateUser(ctx, sqlc.CreateUserParams{
			ID:           id,
			Login:        login,
			PasswordHash: "hash",
			FirstName:    firstName,
			LastName:     toText(lastName),
			MiddleName:   pgtype.Text{},
			GradebookNumber: pgtype.Text{
				String: gradebook,
				Valid:  gradebook != "",
			},
			GroupNumber: toText(groupNumber),
			Institute:   toText(institute),
			BirthDate:   pgtype.Date{},
			Phone:       toText(phone),
			SocialLinks: []byte(`["https://example.org/profile"]`),
			About:       toText(about),
		})
		if err != nil {
			t.Fatalf("create user %s: %v", id, err)
		}
	}

	createUser("user-1", "combo_login", "Alice", "Able", "Engineering", "organizing committee", "E-01", "+70000000001", "GB-001")
	createUser("user-2", "beta_login", "Bob", "Baker", "Math", "field work", "M-02", "+70000000002", "GB-002")
	createUser("user-3", "archived_login", "Cara", "Zulu", "History", "archived staff", "H-03", "+70000000003", "GB-003")
	createUser("user-4", "nomember_login", "Dan", "Clark", "Physics", "no memberships", "P-04", "+70000000004", "GB-004")

	if err := store.Queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		UserID:     "user-1",
		PositionID: "pos-active",
	}); err != nil {
		t.Fatalf("assign active membership: %v", err)
	}
	if err := store.Queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		UserID:     "user-3",
		PositionID: "pos-archived",
	}); err != nil {
		t.Fatalf("assign archived membership: %v", err)
	}
}

func toText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}
