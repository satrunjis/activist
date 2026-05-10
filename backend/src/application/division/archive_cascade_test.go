package division_test

import (
	"context"
	"testing"
	"time"

	appdivision "activist-base/src/application/division"
	appeventlog "activist-base/src/application/eventlog"
	domaineventlog "activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	"activist-base/src/repository/postgres"
	"activist-base/src/repository/postgres/sqlc"
	testpostgres "activist-base/src/testutil/postgres"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestDivisionArchiveCascade_ArchivesSubtreeAndKeepsSiblingUntouched(t *testing.T) {
	ctx := context.Background()
	store := newArchiveCascadeTestStore(t, ctx)
	seedArchiveCascadeFixture(t, ctx, store)

	divisionRepo := postgres.NewDivisionRepository(store.Queries)
	eventRepo := postgres.NewEventLogRepository(store.Queries)
	commands := appeventlog.NewCommandService(postgres.NewTxManager(store.Pool), eventRepo)

	archiveOnce := func() error {
		current, err := divisionRepo.GetByID(ctx, shared.DivisionID("target-1"))
		if err != nil {
			return err
		}
		if current.IsArchived {
			return nil
		}

		access := archiveAccess(shared.DivisionID("target-1"))
		return commands.ExecuteWithEvent(ctx, func(execCtx context.Context, tx appeventlog.Transaction) (domaineventlog.Entry, error) {
			_, removedMemberships, archivedPositions, didArchive, err := divisionRepo.ArchiveCascadeWithTx(execCtx, tx, shared.DivisionID("target-1"))
			if err != nil {
				return domaineventlog.Entry{}, err
			}
			if !didArchive {
				return domaineventlog.Entry{}, nil
			}
			result, err := appdivision.Archive(appdivision.ArchiveInput{
				ActorID:                 shared.UserID("actor-1"),
				Division:                current,
				ArchivedPositionsCount:  archivedPositions,
				RemovedMembershipsCount: removedMemberships,
				Access:                  access,
				Now:                     time.Now().UTC(),
			})
			if err != nil {
				return domaineventlog.Entry{}, err
			}
			return result.Event, nil
		})
	}

	if err := archiveOnce(); err != nil {
		t.Fatalf("first archive: %v", err)
	}

	target, err := store.Queries.GetDivisionByID(ctx, "target-1")
	if err != nil {
		t.Fatalf("get target division: %v", err)
	}
	grandchild, err := store.Queries.GetDivisionByID(ctx, "grandchild-1")
	if err != nil {
		t.Fatalf("get grandchild division: %v", err)
	}
	sibling, err := store.Queries.GetDivisionByID(ctx, "sibling-1")
	if err != nil {
		t.Fatalf("get sibling division: %v", err)
	}
	if !target.IsArchived {
		t.Fatal("target division must be archived")
	}
	if !grandchild.IsArchived {
		t.Fatal("grandchild division must be archived")
	}
	if sibling.IsArchived {
		t.Fatal("sibling division must stay active")
	}

	targetPosition, err := store.Queries.GetPositionByID(ctx, "pos-target")
	if err != nil {
		t.Fatalf("get target position: %v", err)
	}
	grandchildPosition, err := store.Queries.GetPositionByID(ctx, "pos-grandchild")
	if err != nil {
		t.Fatalf("get grandchild position: %v", err)
	}
	siblingPosition, err := store.Queries.GetPositionByID(ctx, "pos-sibling")
	if err != nil {
		t.Fatalf("get sibling position: %v", err)
	}
	if !targetPosition.IsArchived {
		t.Fatal("target position must be archived")
	}
	if !grandchildPosition.IsArchived {
		t.Fatal("grandchild position must be archived")
	}
	if siblingPosition.IsArchived {
		t.Fatal("sibling position must stay active")
	}

	targetMembershipExists, err := store.Queries.MembershipExists(ctx, sqlc.MembershipExistsParams{
		UserID:     "member-target",
		PositionID: "pos-target",
	})
	if err != nil {
		t.Fatalf("target membership exists check: %v", err)
	}
	grandchildMembershipExists, err := store.Queries.MembershipExists(ctx, sqlc.MembershipExistsParams{
		UserID:     "member-grandchild",
		PositionID: "pos-grandchild",
	})
	if err != nil {
		t.Fatalf("grandchild membership exists check: %v", err)
	}
	siblingMembershipExists, err := store.Queries.MembershipExists(ctx, sqlc.MembershipExistsParams{
		UserID:     "member-sibling",
		PositionID: "pos-sibling",
	})
	if err != nil {
		t.Fatalf("sibling membership exists check: %v", err)
	}
	if targetMembershipExists || grandchildMembershipExists {
		t.Fatal("memberships in archived subtree must be removed")
	}
	if !siblingMembershipExists {
		t.Fatal("sibling membership must stay intact")
	}

	divisionArchived := string(domaineventlog.EventDivisionArchived)
	subjectDivision := string(domaineventlog.SubjectDivision)
	subjectID := "target-1"
	firstEventCount, err := store.Queries.CountEventLog(ctx, sqlc.CountEventLogParams{
		EventType:   pgtype.Text{String: divisionArchived, Valid: true},
		SubjectType: pgtype.Text{String: subjectDivision, Valid: true},
		SubjectID:   pgtype.Text{String: subjectID, Valid: true},
	})
	if err != nil {
		t.Fatalf("count first archive events: %v", err)
	}
	if firstEventCount != 1 {
		t.Fatalf("archive event rows after first archive = %d, want 1", firstEventCount)
	}

	if err := archiveOnce(); err != nil {
		t.Fatalf("second archive: %v", err)
	}
	secondEventCount, err := store.Queries.CountEventLog(ctx, sqlc.CountEventLogParams{
		EventType:   pgtype.Text{String: divisionArchived, Valid: true},
		SubjectType: pgtype.Text{String: subjectDivision, Valid: true},
		SubjectID:   pgtype.Text{String: subjectID, Valid: true},
	})
	if err != nil {
		t.Fatalf("count second archive events: %v", err)
	}
	if secondEventCount != 1 {
		t.Fatalf("archive event rows after second archive = %d, want 1", secondEventCount)
	}

	siblingMembershipStillExists, err := store.Queries.MembershipExists(ctx, sqlc.MembershipExistsParams{
		UserID:     "member-sibling",
		PositionID: "pos-sibling",
	})
	if err != nil {
		t.Fatalf("sibling membership exists check after second archive: %v", err)
	}
	if !siblingMembershipStillExists {
		t.Fatal("sibling membership must remain intact after repeated archive")
	}

	siblingPositionAfterSecondArchive, err := store.Queries.GetPositionByID(ctx, "pos-sibling")
	if err != nil {
		t.Fatalf("get sibling position after second archive: %v", err)
	}
	if siblingPositionAfterSecondArchive.IsArchived {
		t.Fatal("sibling position must stay active after repeated archive")
	}
}

func newArchiveCascadeTestStore(t *testing.T, ctx context.Context) *postgres.Store {
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

func seedArchiveCascadeFixture(t *testing.T, ctx context.Context, store *postgres.Store) {
	t.Helper()

	createUser := func(id, login string) {
		t.Helper()
		_, err := store.Queries.CreateUser(ctx, sqlc.CreateUserParams{
			ID:           id,
			Login:        login,
			PasswordHash: "hash",
			FirstName:    "User",
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

	createUser("actor-1", "actor_one")
	createUser("member-target", "member_target")
	createUser("member-grandchild", "member_grandchild")
	createUser("member-sibling", "member_sibling")

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

	createDivision := func(id string, parentID *string, shortName string) {
		t.Helper()
		var parent pgtype.Text
		if parentID != nil {
			parent = pgtype.Text{String: *parentID, Valid: true}
		}
		_, err := store.Queries.CreateDivision(ctx, sqlc.CreateDivisionParams{
			ID:            id,
			ParentID:      parent,
			ShortName:     shortName,
			FullName:      shortName + " Division",
			Description:   "",
			RegulationUrl: "",
			MediaLinks:    []byte(`[]`),
			IsArchived:    false,
		})
		if err != nil {
			t.Fatalf("create division %s: %v", id, err)
		}
	}

	rootID := "root-1"
	targetID := "target-1"
	grandchildID := "grandchild-1"
	siblingID := "sibling-1"

	createDivision(rootID, nil, "Root")
	createDivision(targetID, &rootID, "Target")
	createDivision(grandchildID, &targetID, "Grandchild")
	createDivision(siblingID, &rootID, "Sibling")

	createPosition := func(id, title, divisionID string) {
		t.Helper()
		if _, err := store.Queries.CreatePosition(ctx, sqlc.CreatePositionParams{
			ID:         id,
			Title:      title,
			RoleID:     "role-1",
			DivisionID: divisionID,
			MaxCount:   pgtype.Int4{},
		}); err != nil {
			t.Fatalf("create position %s: %v", id, err)
		}
	}

	createPosition("pos-target", "Target Position", targetID)
	createPosition("pos-grandchild", "Grandchild Position", grandchildID)
	createPosition("pos-sibling", "Sibling Position", siblingID)

	if err := store.Queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		UserID:     "member-target",
		PositionID: "pos-target",
	}); err != nil {
		t.Fatalf("create target membership: %v", err)
	}
	if err := store.Queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		UserID:     "member-grandchild",
		PositionID: "pos-grandchild",
	}); err != nil {
		t.Fatalf("create grandchild membership: %v", err)
	}
	if err := store.Queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		UserID:     "member-sibling",
		PositionID: "pos-sibling",
	}); err != nil {
		t.Fatalf("create sibling membership: %v", err)
	}
}

func archiveAccess(divisionID shared.DivisionID) domainmembership.EffectivePermissions {
	permission, err := domainrole.NewPermission(domainrole.CanArchiveDivision, domainrole.ScopeCurrentAndDescendants)
	if err != nil {
		panic(err)
	}
	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  divisionID,
				Permissions: domainrole.NewPermissionSet([]domainrole.Permission{permission}),
			},
		},
		divisionID,
		nil,
	)
}
