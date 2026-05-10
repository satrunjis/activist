package eventlog

import (
	"context"
	"fmt"

	domaineventlog "activist-base/src/domain/eventlog"
)

type Transaction interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type TxManager interface {
	Begin(ctx context.Context) (Transaction, error)
}

type EventWriter interface {
	InsertWithTx(ctx context.Context, tx Transaction, entry domaineventlog.Entry) (domaineventlog.Entry, error)
}

type MutationWithEventFunc func(ctx context.Context, tx Transaction) (domaineventlog.Entry, error)

type CommandService struct {
	txManager   TxManager
	eventWriter EventWriter
}

func NewCommandService(txManager TxManager, eventWriter EventWriter) *CommandService {
	return &CommandService{
		txManager:   txManager,
		eventWriter: eventWriter,
	}
}

func (s *CommandService) ExecuteWithEvent(ctx context.Context, mutate MutationWithEventFunc) error {
	if s == nil || s.txManager == nil || s.eventWriter == nil {
		return fmt.Errorf("eventlog command service is not configured")
	}
	if mutate == nil {
		return fmt.Errorf("mutation function is required")
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	rolledBack := false
	rollback := func() {
		if rolledBack {
			return
		}
		_ = tx.Rollback(ctx)
		rolledBack = true
	}

	entry, err := mutate(ctx, tx)
	if err != nil {
		rollback()
		return err
	}

	if _, err := s.eventWriter.InsertWithTx(ctx, tx, entry); err != nil {
		rollback()
		return fmt.Errorf("insert event log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		rollback()
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
