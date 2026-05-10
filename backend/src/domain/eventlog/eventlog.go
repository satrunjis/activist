package eventlog

import (
	"time"

	"activist-base/src/domain/shared"
)

var (
	ErrInvalidEventType   = &shared.Error{Code: "eventlog.invalid_event_type", Message: "event type is not supported"}
	ErrInvalidSubjectType = &shared.Error{Code: "eventlog.invalid_subject_type", Message: "event subject type is not supported"}
)

// Payload version stamps stored events so future schema changes can be detected.
const (
	PayloadVersion    = "1"
	PayloadVersionKey = "payload_version"
)

type EventType string

const (
	EventDivisionCreated  EventType = "division_created"
	EventDivisionEdited   EventType = "division_edited"
	EventDivisionArchived EventType = "division_archived"
	EventPositionCreated  EventType = "position_created"
	EventPositionArchived EventType = "position_archived"
	EventPositionAssigned EventType = "position_assigned"
	EventPositionRemoved  EventType = "position_removed"
	EventRoleCreated      EventType = "role_created"
	EventRoleEdited       EventType = "role_edited"
)

func (t EventType) IsValid() bool {
	switch t {
	case EventDivisionCreated, EventDivisionEdited, EventDivisionArchived,
		EventPositionCreated, EventPositionArchived,
		EventPositionAssigned, EventPositionRemoved,
		EventRoleCreated, EventRoleEdited:
		return true
	}
	return false
}

type SubjectType string

const (
	SubjectUser       SubjectType = "user"
	SubjectDivision   SubjectType = "division"
	SubjectPosition   SubjectType = "position"
	SubjectMembership SubjectType = "membership"
	SubjectRole       SubjectType = "role"
)

func (s SubjectType) IsValid() bool {
	switch s {
	case SubjectUser, SubjectDivision, SubjectPosition, SubjectMembership, SubjectRole:
		return true
	}
	return false
}

// Payload is the key-value bag attached to each event.
// It always contains PayloadVersionKey.
type Payload map[string]string

// NewPayload builds a versioned event payload, trimming and validating each entry.
func NewPayload(pairs map[string]string) (Payload, error) {
	p := make(Payload, len(pairs)+1)
	p[PayloadVersionKey] = PayloadVersion
	for k, v := range pairs {
		k = shared.Trim(k)
		if err := shared.RequireText("eventlog.payload_key", k, shared.MaxPayloadKey); err != nil {
			return nil, err
		}
		v = shared.Trim(v)
		if err := shared.AllowText("eventlog.payload_value", v, shared.MaxPayloadValue); err != nil {
			return nil, err
		}
		p[k] = v
	}
	return p, nil
}

func (p Payload) Get(key string) (string, bool) {
	v, ok := p[key]
	return v, ok
}

// Entry is a single immutable audit log record.
type Entry struct {
	ID          shared.EventLogID
	EventType   EventType
	ActorID     shared.UserID
	SubjectType SubjectType
	SubjectID   string
	Payload     Payload
	Timestamp   time.Time
}

func (e Entry) validate() error {
	if !e.EventType.IsValid() {
		return ErrInvalidEventType
	}
	if !e.SubjectType.IsValid() {
		return ErrInvalidSubjectType
	}
	if err := shared.RequireID("eventlog.actor_id", string(e.ActorID)); err != nil {
		return err
	}
	return shared.RequireText("eventlog.subject_id", e.SubjectID, shared.MaxEventSubjectID)
}

// NewEntry is the only way to construct a valid Entry in use cases.
// If now is zero, the current UTC time is used.
func NewEntry(
	actorID shared.UserID,
	eventType EventType,
	subjectType SubjectType,
	subjectID string,
	payload Payload,
	now time.Time,
) (Entry, error) {
	ts := now.UTC()
	if now.IsZero() {
		ts = time.Now().UTC()
	}
	e := Entry{
		EventType:   eventType,
		ActorID:     actorID,
		SubjectType: subjectType,
		SubjectID:   subjectID,
		Payload:     payload,
		Timestamp:   ts,
	}
	return e, e.validate()
}
