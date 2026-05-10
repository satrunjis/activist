package postgres_test

import (
	"context"
	"strings"
	"testing"
	"time"

	appeventlog "activist-base/src/application/eventlog"
	domaineventlog "activist-base/src/domain/eventlog"
	"activist-base/src/domain/shared"
	"activist-base/src/repository/postgres"
	"activist-base/src/repository/postgres/sqlc"
	testpostgres "activist-base/src/testutil/postgres"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestEventLogRepositoryInsertPersistsLosslessly(t *testing.T) {
	ctx := context.Background()
	store := newEventLogTestStore(t, ctx)
	seedEventLogActor(t, ctx, store, "actor-1", "actor_one")

	repo := postgres.NewEventLogRepository(store.Queries)
	ts := time.Date(2026, 4, 18, 7, 0, 0, 0, time.UTC)
	payload, err := domaineventlog.NewPayload(map[string]string{
		"division_id": "div-1",
		"changed_by":  "actor-1",
	})
	if err != nil {
		t.Fatalf("new payload: %v", err)
	}

	entry, err := domaineventlog.NewEntry(
		shared.UserID("actor-1"),
		domaineventlog.EventDivisionArchived,
		domaineventlog.SubjectDivision,
		"div-1",
		payload,
		ts,
	)
	if err != nil {
		t.Fatalf("new entry: %v", err)
	}

	inserted, err := repo.Insert(ctx, entry)
	if err != nil {
		t.Fatalf("insert event log: %v", err)
	}

	if inserted.ID == "" {
		t.Fatal("expected generated id")
	}
	if inserted.EventType != entry.EventType {
		t.Fatalf("event_type mismatch: got %q want %q", inserted.EventType, entry.EventType)
	}
	if inserted.ActorID != entry.ActorID {
		t.Fatalf("actor_id mismatch: got %q want %q", inserted.ActorID, entry.ActorID)
	}
	if inserted.SubjectType != entry.SubjectType {
		t.Fatalf("subject_type mismatch: got %q want %q", inserted.SubjectType, entry.SubjectType)
	}
	if inserted.SubjectID != entry.SubjectID {
		t.Fatalf("subject_id mismatch: got %q want %q", inserted.SubjectID, entry.SubjectID)
	}
	if !inserted.Timestamp.Equal(entry.Timestamp) {
		t.Fatalf("timestamp mismatch: got %v want %v", inserted.Timestamp, entry.Timestamp)
	}
	if inserted.Payload[domaineventlog.PayloadVersionKey] != domaineventlog.PayloadVersion {
		t.Fatalf("payload version mismatch: got %q want %q", inserted.Payload[domaineventlog.PayloadVersionKey], domaineventlog.PayloadVersion)
	}
	if inserted.Payload["division_id"] != "div-1" {
		t.Fatalf("payload division_id mismatch: got %q", inserted.Payload["division_id"])
	}
}

func TestEventLogRepositoryListAppliesFiltersPaginationAndStableSort(t *testing.T) {
	ctx := context.Background()
	store := newEventLogTestStore(t, ctx)
	seedEventLogActor(t, ctx, store, "actor-1", "actor_one")

	repo := postgres.NewEventLogRepository(store.Queries)
	base := time.Date(2026, 4, 18, 7, 0, 0, 0, time.UTC)

	mustInsertEvent(t, ctx, repo, newEvent(t, "actor-1", domaineventlog.EventPositionAssigned, domaineventlog.SubjectMembership, "m-1", base.Add(1*time.Minute)))
	mustInsertEvent(t, ctx, repo, newEvent(t, "actor-1", domaineventlog.EventPositionRemoved, domaineventlog.SubjectMembership, "m-2", base.Add(2*time.Minute)))
	mustInsertEvent(t, ctx, repo, newEvent(t, "actor-1", domaineventlog.EventDivisionEdited, domaineventlog.SubjectDivision, "d-1", base.Add(3*time.Minute)))
	mustInsertEvent(t, ctx, repo, newEvent(t, "actor-1", domaineventlog.EventDivisionArchived, domaineventlog.SubjectDivision, "d-1", base.Add(4*time.Minute)))

	filterEvent := domaineventlog.EventDivisionArchived
	filterSubject := domaineventlog.SubjectDivision
	filterSubjectID := "d-1"
	list, err := repo.List(ctx, appeventlog.ListInput{
		EventType:   &filterEvent,
		SubjectType: &filterSubject,
		SubjectID:   &filterSubjectID,
		Limit:       10,
		Offset:      0,
	})
	if err != nil {
		t.Fatalf("list filtered events: %v", err)
	}
	if list.Total != 1 {
		t.Fatalf("filtered total mismatch: got %d want 1", list.Total)
	}
	if len(list.Items) != 1 {
		t.Fatalf("filtered items mismatch: got %d want 1", len(list.Items))
	}
	if list.Items[0].EventType != domaineventlog.EventDivisionArchived {
		t.Fatalf("unexpected filtered event type: %q", list.Items[0].EventType)
	}

	paged, err := repo.List(ctx, appeventlog.ListInput{
		Limit:  2,
		Offset: 1,
	})
	if err != nil {
		t.Fatalf("list paged events: %v", err)
	}
	if paged.Total != 4 {
		t.Fatalf("paged total mismatch: got %d want 4", paged.Total)
	}
	if len(paged.Items) != 2 {
		t.Fatalf("paged items mismatch: got %d want 2", len(paged.Items))
	}

	for i := 1; i < len(paged.Items); i++ {
		prev := paged.Items[i-1]
		curr := paged.Items[i]
		if prev.Timestamp.Before(curr.Timestamp) {
			t.Fatalf("ordering mismatch: %v should be >= %v", prev.Timestamp, curr.Timestamp)
		}
		if prev.Timestamp.Equal(curr.Timestamp) && string(prev.ID) < string(curr.ID) {
			t.Fatalf("id tie-break ordering mismatch: %s should be >= %s", prev.ID, curr.ID)
		}
	}
}

func TestEventLogStorageIsAppendOnly(t *testing.T) {
	ctx := context.Background()
	store := newEventLogTestStore(t, ctx)
	seedEventLogActor(t, ctx, store, "actor-1", "actor_one")

	repo := postgres.NewEventLogRepository(store.Queries)
	inserted := mustInsertEvent(t, ctx, repo, newEvent(t, "actor-1", domaineventlog.EventDivisionCreated, domaineventlog.SubjectDivision, "d-1", time.Now().UTC()))

	if _, err := store.Pool.Exec(ctx, "UPDATE event_log SET subject_id = 'mutated' WHERE id = $1", string(inserted.ID)); err == nil {
		t.Fatal("expected update to fail for append-only table")
	} else if !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("unexpected update error: %v", err)
	}

	if _, err := store.Pool.Exec(ctx, "DELETE FROM event_log WHERE id = $1", string(inserted.ID)); err == nil {
		t.Fatal("expected delete to fail for append-only table")
	} else if !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("unexpected delete error: %v", err)
	}
}

func newEventLogTestStore(t *testing.T, ctx context.Context) *postgres.Store {
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

func seedEventLogActor(t *testing.T, ctx context.Context, store *postgres.Store, id, login string) {
	t.Helper()

	_, err := store.Queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           id,
		Login:        login,
		PasswordHash: "hash",
		FirstName:    "Actor",
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
		t.Fatalf("create actor user: %v", err)
	}
}

func newEvent(
	t *testing.T,
	actorID string,
	eventType domaineventlog.EventType,
	subjectType domaineventlog.SubjectType,
	subjectID string,
	ts time.Time,
) domaineventlog.Entry {
	t.Helper()

	payload, err := domaineventlog.NewPayload(map[string]string{"subject_id": subjectID})
	if err != nil {
		t.Fatalf("new payload: %v", err)
	}
	entry, err := domaineventlog.NewEntry(
		shared.UserID(actorID),
		eventType,
		subjectType,
		subjectID,
		payload,
		ts,
	)
	if err != nil {
		t.Fatalf("new entry: %v", err)
	}
	return entry
}

func mustInsertEvent(t *testing.T, ctx context.Context, repo *postgres.EventLogRepository, entry domaineventlog.Entry) domaineventlog.Entry {
	t.Helper()
	inserted, err := repo.Insert(ctx, entry)
	if err != nil {
		t.Fatalf("insert event: %v", err)
	}
	return inserted
}
