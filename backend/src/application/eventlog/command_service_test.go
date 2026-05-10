package eventlog

import (
	"context"
	"errors"
	"testing"
	"time"

	domaineventlog "activist-base/src/domain/eventlog"
	"activist-base/src/domain/shared"
)

func TestCommandServiceExecuteWithEvent_CommitsOnSuccess(t *testing.T) {
	t.Parallel()

	tx := &fakeTransaction{}
	manager := &fakeTxManager{tx: tx}
	writer := &fakeEventWriter{}

	svc := NewCommandService(manager, writer)
	entry := mustNewEventEntry(t)

	mutationCalls := 0
	err := svc.ExecuteWithEvent(context.Background(), func(context.Context, Transaction) (domaineventlog.Entry, error) {
		mutationCalls++
		return entry, nil
	})
	if err != nil {
		t.Fatalf("ExecuteWithEvent() error = %v", err)
	}

	if mutationCalls != 1 {
		t.Fatalf("mutation calls = %d, want 1", mutationCalls)
	}
	if writer.calls != 1 {
		t.Fatalf("event writer calls = %d, want 1", writer.calls)
	}
	if tx.commits != 1 {
		t.Fatalf("tx commits = %d, want 1", tx.commits)
	}
	if tx.rollbacks != 0 {
		t.Fatalf("tx rollbacks = %d, want 0", tx.rollbacks)
	}
}

func TestCommandServiceExecuteWithEvent_RollsBackWhenMutationFails(t *testing.T) {
	t.Parallel()

	tx := &fakeTransaction{}
	manager := &fakeTxManager{tx: tx}
	writer := &fakeEventWriter{}

	svc := NewCommandService(manager, writer)
	expected := errors.New("mutation failed")

	err := svc.ExecuteWithEvent(context.Background(), func(context.Context, Transaction) (domaineventlog.Entry, error) {
		return domaineventlog.Entry{}, expected
	})
	if !errors.Is(err, expected) {
		t.Fatalf("ExecuteWithEvent() error = %v, want %v", err, expected)
	}

	if writer.calls != 0 {
		t.Fatalf("event writer calls = %d, want 0", writer.calls)
	}
	if tx.commits != 0 {
		t.Fatalf("tx commits = %d, want 0", tx.commits)
	}
	if tx.rollbacks != 1 {
		t.Fatalf("tx rollbacks = %d, want 1", tx.rollbacks)
	}
}

func TestCommandServiceExecuteWithEvent_RollsBackWhenEventInsertFails(t *testing.T) {
	t.Parallel()

	tx := &fakeTransaction{}
	manager := &fakeTxManager{tx: tx}
	writer := &fakeEventWriter{err: errors.New("insert failed")}

	svc := NewCommandService(manager, writer)
	entry := mustNewEventEntry(t)

	err := svc.ExecuteWithEvent(context.Background(), func(context.Context, Transaction) (domaineventlog.Entry, error) {
		return entry, nil
	})
	if err == nil {
		t.Fatal("ExecuteWithEvent() error = nil, want non-nil")
	}

	if writer.calls != 1 {
		t.Fatalf("event writer calls = %d, want 1", writer.calls)
	}
	if tx.commits != 0 {
		t.Fatalf("tx commits = %d, want 0", tx.commits)
	}
	if tx.rollbacks != 1 {
		t.Fatalf("tx rollbacks = %d, want 1", tx.rollbacks)
	}
}

func mustNewEventEntry(t *testing.T) domaineventlog.Entry {
	t.Helper()

	payload, err := domaineventlog.NewPayload(map[string]string{
		"division_id": "div-1",
	})
	if err != nil {
		t.Fatalf("NewPayload() error = %v", err)
	}
	entry, err := domaineventlog.NewEntry(
		shared.UserID("actor-1"),
		domaineventlog.EventPositionCreated,
		domaineventlog.SubjectPosition,
		"pos-1",
		payload,
		time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewEntry() error = %v", err)
	}
	return entry
}

type fakeTxManager struct {
	tx  Transaction
	err error
}

func (f *fakeTxManager) Begin(context.Context) (Transaction, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.tx, nil
}

type fakeTransaction struct {
	commits   int
	rollbacks int
	commitErr error
}

func (f *fakeTransaction) Commit(context.Context) error {
	f.commits++
	return f.commitErr
}

func (f *fakeTransaction) Rollback(context.Context) error {
	f.rollbacks++
	return nil
}

type fakeEventWriter struct {
	calls int
	err   error
}

func (f *fakeEventWriter) InsertWithTx(context.Context, Transaction, domaineventlog.Entry) (domaineventlog.Entry, error) {
	f.calls++
	if f.err != nil {
		return domaineventlog.Entry{}, f.err
	}
	return domaineventlog.Entry{}, nil
}
