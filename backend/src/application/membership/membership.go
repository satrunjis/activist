package membership

import (
	"time"

	"activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
)

// — Assign ——————————————————————————————————————————————————————————————————

type AssignInput struct {
	ActorID              shared.UserID
	User                 domainuser.User
	Position             domainposition.Position
	CurrentPositionCount int
	AlreadyAssigned      bool
	Access               domainmembership.EffectivePermissions
	Now                  time.Time
}

type AssignResult struct {
	Membership domainmembership.Membership
	Event      eventlog.Entry
}

func Assign(input AssignInput) (AssignResult, error) {
	if !input.Access.Has(role.CanAssignPosition, false) {
		return AssignResult{}, shared.ErrForbidden
	}
	if input.User.ID == "" {
		return AssignResult{}, shared.ErrEmptyID
	}
	if err := input.Position.Validate(); err != nil {
		return AssignResult{}, err
	}
	if input.AlreadyAssigned {
		return AssignResult{}, domainmembership.ErrAlreadyExists
	}
	if err := input.Position.CanAssign(input.CurrentPositionCount); err != nil {
		return AssignResult{}, err
	}

	m := domainmembership.Membership{
		UserID:     input.User.ID,
		PositionID: input.Position.ID,
	}
	if err := m.Validate(); err != nil {
		return AssignResult{}, err
	}

	event, err := membershipEvent(input.ActorID, eventlog.EventPositionAssigned, m, map[string]string{
		"user_id":        string(m.UserID),
		"position_id":    string(m.PositionID),
		"position_title": input.Position.Title,
		"division_id":    string(input.Position.DivisionID),
	}, input.Now)
	if err != nil {
		return AssignResult{}, err
	}
	return AssignResult{Membership: m, Event: event}, nil
}

// — Remove ——————————————————————————————————————————————————————————————————

type RemoveInput struct {
	ActorID    shared.UserID
	Membership domainmembership.Membership
	Position   domainposition.Position
	Exists     bool
	Access     domainmembership.EffectivePermissions
	Now        time.Time
}

type RemoveResult struct {
	Event eventlog.Entry
}

func Remove(input RemoveInput) (RemoveResult, error) {
	if !input.Access.Has(role.CanRemoveMember, false) {
		return RemoveResult{}, shared.ErrForbidden
	}
	if !input.Exists {
		return RemoveResult{}, domainmembership.ErrNotFound
	}
	if err := input.Membership.Validate(); err != nil {
		return RemoveResult{}, err
	}
	if err := input.Position.Validate(); err != nil {
		return RemoveResult{}, err
	}

	event, err := membershipEvent(input.ActorID, eventlog.EventPositionRemoved, input.Membership, map[string]string{
		"user_id":        string(input.Membership.UserID),
		"position_id":    string(input.Membership.PositionID),
		"position_title": input.Position.Title,
		"division_id":    string(input.Position.DivisionID),
	}, input.Now)
	if err != nil {
		return RemoveResult{}, err
	}
	return RemoveResult{Event: event}, nil
}

// — shared event helper —————————————————————————————————————————————————————

func membershipEvent(
	actorID shared.UserID,
	typ eventlog.EventType,
	m domainmembership.Membership,
	payload map[string]string,
	now time.Time,
) (eventlog.Entry, error) {
	p, err := eventlog.NewPayload(payload)
	if err != nil {
		return eventlog.Entry{}, err
	}
	subjectID := string(m.UserID) + ":" + string(m.PositionID)
	return eventlog.NewEntry(actorID, typ, eventlog.SubjectMembership, subjectID, p, now)
}
