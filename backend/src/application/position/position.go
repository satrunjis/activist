package position

import (
	"strconv"
	"time"

	domaindivision "activist-base/src/domain/division"
	"activist-base/src/domain/eventlog"
	"activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
)

// — Create ——————————————————————————————————————————————————————————————————

type CreateInput struct {
	ActorID  shared.UserID
	ID       shared.PositionID
	Division domaindivision.Division
	Role     domainrole.Role
	Title    string
	MaxCount *int
	Access   membership.EffectivePermissions
	Now      time.Time
}

type CreateResult struct {
	Position domainposition.Position
	Event    eventlog.Entry
}

func Create(input CreateInput) (CreateResult, error) {
	if !input.Access.Has(domainrole.CanManagePositions, false) {
		return CreateResult{}, shared.ErrForbidden
	}
	if input.Division.IsArchived {
		return CreateResult{}, domaindivision.ErrArchived
	}
	if err := input.Role.Validate(); err != nil {
		return CreateResult{}, err
	}

	p := domainposition.Position{
		ID:         input.ID,
		Title:      input.Title,
		RoleID:     input.Role.ID,
		DivisionID: input.Division.ID,
		MaxCount:   input.MaxCount,
	}
	p.Normalize()
	if err := p.Validate(); err != nil {
		return CreateResult{}, err
	}

	event, err := positionEvent(input.ActorID, eventlog.EventPositionCreated, p.ID, map[string]string{
		"position_id": string(p.ID),
		"division_id": string(p.DivisionID),
		"title":       p.Title,
		"role_id":     string(p.RoleID),
	}, input.Now)
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{Position: p, Event: event}, nil
}

// — Archive —————————————————————————————————————————————————————————————————

type ArchiveInput struct {
	ActorID                 shared.UserID
	Position                domainposition.Position
	RemovedMembershipsCount int
	Access                  membership.EffectivePermissions
	Now                     time.Time
}

type ArchiveResult struct {
	Position domainposition.Position
	Event    eventlog.Entry
}

func Archive(input ArchiveInput) (ArchiveResult, error) {
	if !input.Access.Has(domainrole.CanManagePositions, false) {
		return ArchiveResult{}, shared.ErrForbidden
	}
	if input.Position.IsArchived {
		return ArchiveResult{}, domainposition.ErrArchived
	}

	archived := input.Position
	archived.IsArchived = true

	event, err := positionEvent(input.ActorID, eventlog.EventPositionArchived, archived.ID, map[string]string{
		"position_id":               string(archived.ID),
		"division_id":               string(archived.DivisionID),
		"title":                     archived.Title,
		"removed_memberships_count": strconv.Itoa(input.RemovedMembershipsCount),
	}, input.Now)
	if err != nil {
		return ArchiveResult{}, err
	}
	return ArchiveResult{Position: archived, Event: event}, nil
}

// — shared event helper —————————————————————————————————————————————————————

func positionEvent(
	actorID shared.UserID,
	typ eventlog.EventType,
	positionID shared.PositionID,
	payload map[string]string,
	now time.Time,
) (eventlog.Entry, error) {
	p, err := eventlog.NewPayload(payload)
	if err != nil {
		return eventlog.Entry{}, err
	}
	return eventlog.NewEntry(actorID, typ, eventlog.SubjectPosition, string(positionID), p, now)
}
