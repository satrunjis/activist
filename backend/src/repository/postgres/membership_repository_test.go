package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domainmembership "activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	"activist-base/src/repository/postgres"
	"activist-base/src/repository/postgres/sqlc"
	testpostgres "activist-base/src/testutil/postgres"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestAssignWithCountCheckUsesLockedPositionForMaxCount(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	seedMembershipContext(t, ctx, store)

	stalePosition, err := store.Queries.GetPositionByID(ctx, "pos-1")
	if err != nil {
		t.Fatalf("get stale position: %v", err)
	}

	// Simulate concurrent admin update after the caller already has a stale snapshot.
	_, err = store.Queries.UpdatePosition(ctx, sqlc.UpdatePositionParams{
		ID:       "pos-1",
		Title:    stalePosition.Title,
		RoleID:   stalePosition.RoleID,
		MaxCount: pgtype.Int4{Int32: 1, Valid: true},
	})
	if err != nil {
		t.Fatalf("tighten max_count: %v", err)
	}

	repo := postgres.NewMembershipRepository(store.Queries, store.Pool)
	_, err = repo.AssignWithCountCheck(
		ctx,
		shared.UserID("user-2"),
		toDomainPosition(stalePosition),
		assignAccess(shared.DivisionID("div-1")),
		shared.UserID("actor-1"),
	)
	if !errors.Is(err, domainposition.ErrLimitReached) {
		t.Fatalf("expected ErrLimitReached, got %v", err)
	}

	exists, err := store.Queries.MembershipExists(ctx, sqlc.MembershipExistsParams{
		UserID:     "user-2",
		PositionID: "pos-1",
	})
	if err != nil {
		t.Fatalf("membership exists check: %v", err)
	}
	if exists {
		t.Fatal("membership should not be created when max_count is reached")
	}
}

func TestAssignWithCountCheckUsesLockedPositionForArchivedState(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	seedMembershipContext(t, ctx, store)

	stalePosition, err := store.Queries.GetPositionByID(ctx, "pos-1")
	if err != nil {
		t.Fatalf("get stale position: %v", err)
	}

	// Simulate concurrent archive after stale read.
	if _, err := store.Queries.ArchivePosition(ctx, "pos-1"); err != nil {
		t.Fatalf("archive position: %v", err)
	}

	repo := postgres.NewMembershipRepository(store.Queries, store.Pool)
	_, err = repo.AssignWithCountCheck(
		ctx,
		shared.UserID("user-2"),
		toDomainPosition(stalePosition),
		assignAccess(shared.DivisionID("div-1")),
		shared.UserID("actor-1"),
	)
	if !errors.Is(err, domainposition.ErrArchived) {
		t.Fatalf("expected ErrArchived, got %v", err)
	}

	exists, err := store.Queries.MembershipExists(ctx, sqlc.MembershipExistsParams{
		UserID:     "user-2",
		PositionID: "pos-1",
	})
	if err != nil {
		t.Fatalf("membership exists check: %v", err)
	}
	if exists {
		t.Fatal("membership should not be created for archived position")
	}
}

func TestAssignWithCountCheckRejectsDuplicateMembership(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	seedMembershipContext(t, ctx, store)

	position, err := store.Queries.GetPositionByID(ctx, "pos-1")
	if err != nil {
		t.Fatalf("get position: %v", err)
	}

	repo := postgres.NewMembershipRepository(store.Queries, store.Pool)
	_, err = repo.AssignWithCountCheck(
		ctx,
		shared.UserID("user-1"),
		toDomainPosition(position),
		assignAccess(shared.DivisionID("div-1")),
		shared.UserID("actor-1"),
	)
	if !errors.Is(err, domainmembership.ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}

	count, err := store.Queries.CountMembersByPosition(ctx, "pos-1")
	if err != nil {
		t.Fatalf("count memberships: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected membership count to remain 1, got %d", count)
	}
}

func TestListMembersByPositionReturnsDeterministicOrder(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	seedMembershipContext(t, ctx, store)

	if err := store.Queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		UserID:     "user-2",
		PositionID: "pos-1",
	}); err != nil {
		t.Fatalf("create membership: %v", err)
	}

	repo := postgres.NewMembershipRepository(store.Queries, store.Pool)
	items, err := repo.ListMembersByPosition(ctx, shared.PositionID("pos-1"))
	if err != nil {
		t.Fatalf("list members by position: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].Membership.UserID != shared.UserID("user-2") || items[1].Membership.UserID != shared.UserID("user-1") {
		t.Fatalf(
			"expected deterministic order [user-2,user-1], got [%s,%s]",
			items[0].Membership.UserID,
			items[1].Membership.UserID,
		)
	}
	if items[0].Membership.PositionID != shared.PositionID("pos-1") || items[0].Position.RoleID != shared.RoleID("role-1") {
		t.Fatalf("unexpected payload mapping: %+v", items[0])
	}
}

func TestListMembersByPositionExcludesArchivedAndNonMatchingRows(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	seedMembershipContext(t, ctx, store)

	if _, err := store.Queries.CreatePosition(ctx, sqlc.CreatePositionParams{
		ID:         "pos-2",
		Title:      "Deputy",
		RoleID:     "role-1",
		DivisionID: "div-1",
		MaxCount:   pgtype.Int4{Int32: 3, Valid: true},
	}); err != nil {
		t.Fatalf("create position: %v", err)
	}
	if err := store.Queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		UserID:     "user-2",
		PositionID: "pos-2",
	}); err != nil {
		t.Fatalf("create second membership: %v", err)
	}

	repo := postgres.NewMembershipRepository(store.Queries, store.Pool)
	activeItems, err := repo.ListMembersByPosition(ctx, shared.PositionID("pos-1"))
	if err != nil {
		t.Fatalf("list active members: %v", err)
	}
	if len(activeItems) != 1 || activeItems[0].Membership.UserID != shared.UserID("user-1") {
		t.Fatalf("expected only user-1 for pos-1, got %+v", activeItems)
	}

	if _, err := store.Queries.ArchivePosition(ctx, "pos-1"); err != nil {
		t.Fatalf("archive position: %v", err)
	}
	archivedItems, err := repo.ListMembersByPosition(ctx, shared.PositionID("pos-1"))
	if err != nil {
		t.Fatalf("list archived members: %v", err)
	}
	if len(archivedItems) != 0 {
		t.Fatalf("expected archived position members excluded, got %d", len(archivedItems))
	}
}

func TestListMembersByPositionUnknownPositionReturnsEmpty(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t, ctx)
	seedMembershipContext(t, ctx, store)

	beforeCount, err := store.Queries.CountMembersByPosition(ctx, "pos-1")
	if err != nil {
		t.Fatalf("count memberships before read: %v", err)
	}

	repo := postgres.NewMembershipRepository(store.Queries, store.Pool)
	items, err := repo.ListMembersByPosition(ctx, shared.PositionID("missing-position"))
	if err != nil {
		t.Fatalf("list members by position: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected empty list for unknown position, got %d", len(items))
	}

	afterCount, err := store.Queries.CountMembersByPosition(ctx, "pos-1")
	if err != nil {
		t.Fatalf("count memberships after read: %v", err)
	}
	if afterCount != beforeCount {
		t.Fatalf("read path must not mutate memberships: before=%d after=%d", beforeCount, afterCount)
	}
}

func newTestStore(t *testing.T, ctx context.Context) *postgres.Store {
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

func seedMembershipContext(t *testing.T, ctx context.Context, store *postgres.Store) {
	t.Helper()

	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	if _, err := store.Queries.CreateRole(ctx, sqlc.CreateRoleParams{
		ID:          "role-1",
		Name:        "Organizer",
		Permissions: []byte{},
		CreatedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		t.Fatalf("create role: %v", err)
	}

	if _, err := store.Queries.CreateDivision(ctx, sqlc.CreateDivisionParams{
		ID:            "div-1",
		ParentID:      pgtype.Text{},
		ShortName:     "Ops",
		FullName:      "Operations",
		Description:   "",
		RegulationUrl: "",
		MediaLinks:    []byte(`[]`),
		IsArchived:    false,
	}); err != nil {
		t.Fatalf("create division: %v", err)
	}

	if _, err := store.Queries.CreatePosition(ctx, sqlc.CreatePositionParams{
		ID:         "pos-1",
		Title:      "Chair",
		RoleID:     "role-1",
		DivisionID: "div-1",
		MaxCount:   pgtype.Int4{Int32: 2, Valid: true},
	}); err != nil {
		t.Fatalf("create position: %v", err)
	}

	createUser := func(id, login, firstName string) {
		t.Helper()
		_, err := store.Queries.CreateUser(ctx, sqlc.CreateUserParams{
			ID:           id,
			Login:        login,
			PasswordHash: "hash",
			FirstName:    firstName,
			LastName:     pgtype.Text{},
			MiddleName:   pgtype.Text{},
			GradebookNumber: pgtype.Text{
				String: id + "-gb",
				Valid:  true,
			},
			GroupNumber: pgtype.Text{},
			Institute:   pgtype.Text{},
			BirthDate:   pgtype.Date{},
			Phone:       pgtype.Text{},
			SocialLinks: []byte(`[]`),
			About:       pgtype.Text{},
		})
		if err != nil {
			t.Fatalf("create user %s: %v", id, err)
		}
	}
	createUser("user-1", "user_one", "Zed")
	createUser("user-2", "user_two", "Amy")

	if err := store.Queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		UserID:     "user-1",
		PositionID: "pos-1",
	}); err != nil {
		t.Fatalf("create existing membership: %v", err)
	}
}

func assignAccess(divisionID shared.DivisionID) domainmembership.EffectivePermissions {
	perm, err := domainrole.NewPermission(domainrole.CanAssignPosition, domainrole.ScopeCurrentDivision)
	if err != nil {
		panic(err)
	}
	return domainmembership.Calculate([]domainmembership.MembershipContext{
		{
			DivisionID:  divisionID,
			Permissions: domainrole.NewPermissionSet([]domainrole.Permission{perm}),
		},
	}, divisionID, nil)
}

func toDomainPosition(p sqlc.Position) domainposition.Position {
	var maxCount *int
	if p.MaxCount.Valid {
		n := int(p.MaxCount.Int32)
		maxCount = &n
	}
	return domainposition.Position{
		ID:         shared.PositionID(p.ID),
		Title:      p.Title,
		RoleID:     shared.RoleID(p.RoleID),
		DivisionID: shared.DivisionID(p.DivisionID),
		MaxCount:   maxCount,
		IsArchived: p.IsArchived,
	}
}
