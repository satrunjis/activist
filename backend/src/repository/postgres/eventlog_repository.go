package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	appeventlog "activist-base/src/application/eventlog"
	domaineventlog "activist-base/src/domain/eventlog"
	"activist-base/src/domain/shared"
	"activist-base/src/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5/pgtype"
)

type EventLogRepository struct {
	queries *sqlc.Queries
}

func NewEventLogRepository(queries *sqlc.Queries) *EventLogRepository {
	return &EventLogRepository{queries: queries}
}

func (r *EventLogRepository) Insert(ctx context.Context, entry domaineventlog.Entry) (domaineventlog.Entry, error) {
	return r.insertWithQueries(ctx, r.queries, entry)
}

func (r *EventLogRepository) InsertWithTx(ctx context.Context, tx appeventlog.Transaction, entry domaineventlog.Entry) (domaineventlog.Entry, error) {
	pgxTx, err := unwrapPGXTx(tx)
	if err != nil {
		return domaineventlog.Entry{}, err
	}
	return r.insertWithQueries(ctx, r.queries.WithTx(pgxTx), entry)
}

func (r *EventLogRepository) insertWithQueries(ctx context.Context, queries *sqlc.Queries, entry domaineventlog.Entry) (domaineventlog.Entry, error) {
	id := string(entry.ID)
	if id == "" {
		var err error
		id, err = newEventLogID()
		if err != nil {
			return domaineventlog.Entry{}, fmt.Errorf("generate event log id: %w", err)
		}
	}

	payload, err := json.Marshal(entry.Payload)
	if err != nil {
		return domaineventlog.Entry{}, fmt.Errorf("marshal payload: %w", err)
	}

	persisted, err := queries.InsertEventLog(ctx, sqlc.InsertEventLogParams{
		ID:          id,
		EventType:   string(entry.EventType),
		ActorID:     string(entry.ActorID),
		SubjectType: string(entry.SubjectType),
		SubjectID:   entry.SubjectID,
		Payload:     payload,
		Timestamp:   timestamptz(entry.Timestamp),
	})
	if err != nil {
		return domaineventlog.Entry{}, err
	}
	return toDomainEventLog(persisted)
}

func (r *EventLogRepository) List(ctx context.Context, input appeventlog.ListInput) (appeventlog.ListResult, error) {
	normalized, err := appeventlog.NormalizeListInput(input)
	if err != nil {
		return appeventlog.ListResult{}, err
	}

	count, err := r.queries.CountEventLog(ctx, sqlc.CountEventLogParams{
		EventType:   eventTypeFilter(normalized.EventType),
		SubjectType: subjectTypeFilter(normalized.SubjectType),
		SubjectID:   textFilter(normalized.SubjectID),
	})
	if err != nil {
		return appeventlog.ListResult{}, err
	}

	rows, err := r.queries.ListEventLog(ctx, sqlc.ListEventLogParams{
		EventType:   eventTypeFilter(normalized.EventType),
		SubjectType: subjectTypeFilter(normalized.SubjectType),
		SubjectID:   textFilter(normalized.SubjectID),
		PageOffset:  int32(normalized.Offset),
		PageLimit:   int32(normalized.Limit),
	})
	if err != nil {
		return appeventlog.ListResult{}, err
	}

	items := make([]domaineventlog.Entry, 0, len(rows))
	for _, row := range rows {
		item, err := toDomainEventLog(row)
		if err != nil {
			return appeventlog.ListResult{}, err
		}
		items = append(items, item)
	}

	return appeventlog.ListResult{
		Items:  items,
		Total:  int(count),
		Limit:  normalized.Limit,
		Offset: normalized.Offset,
	}, nil
}

func toDomainEventLog(persisted sqlc.EventLog) (domaineventlog.Entry, error) {
	ts, err := timeFromTimestamp(persisted.Timestamp)
	if err != nil {
		return domaineventlog.Entry{}, err
	}

	var payload domaineventlog.Payload
	if err := json.Unmarshal(persisted.Payload, &payload); err != nil {
		return domaineventlog.Entry{}, fmt.Errorf("unmarshal payload: %w", err)
	}

	return domaineventlog.Entry{
		ID:          shared.EventLogID(persisted.ID),
		EventType:   domaineventlog.EventType(persisted.EventType),
		ActorID:     shared.UserID(persisted.ActorID),
		SubjectType: domaineventlog.SubjectType(persisted.SubjectType),
		SubjectID:   persisted.SubjectID,
		Payload:     payload,
		Timestamp:   ts,
	}, nil
}

func eventTypeFilter(value *domaineventlog.EventType) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return textOrNull(string(*value))
}

func subjectTypeFilter(value *domaineventlog.SubjectType) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return textOrNull(string(*value))
}

func textFilter(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return textOrNull(*value)
}

func newEventLogID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
