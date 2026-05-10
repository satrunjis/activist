package role

import (
	"strings"
	"time"

	"activist-base/src/domain/eventlog"
	"activist-base/src/domain/membership"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
)

var ErrPrivilegeEscalation = &shared.Error{
	Code:    "access.privilege_escalation",
	Message: "cannot assign a permission the actor does not hold",
}

// — Create ——————————————————————————————————————————————————————————————————

type CreateInput struct {
	ActorID     shared.UserID
	ID          shared.RoleID
	Name        string
	Permissions domainrole.PermissionSet
	Access      membership.EffectivePermissions
	Now         time.Time
}

type CreateResult struct {
	Role  domainrole.Role
	Event eventlog.Entry
}

func Create(input CreateInput) (CreateResult, error) {
	if !input.Access.Has(domainrole.SystemAdmin, false) {
		return CreateResult{}, shared.ErrForbidden
	}

	now := input.Now.UTC()
	if input.Now.IsZero() {
		now = time.Now().UTC()
	}
	r := domainrole.Role{
		ID:          input.ID,
		Name:        input.Name,
		Permissions: input.Permissions,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	r.Normalize()
	if err := r.Validate(); err != nil {
		return CreateResult{}, err
	}

	event, err := roleEvent(input.ActorID, eventlog.EventRoleCreated, r.ID, map[string]string{
		"role_id": string(r.ID),
		"name":    r.Name,
	}, input.Now)
	if err != nil {
		return CreateResult{}, err
	}
	return CreateResult{Role: r, Event: event}, nil
}

// — Edit ————————————————————————————————————————————————————————————————————

type EditPatch struct {
	Name        *string
	Permissions *domainrole.PermissionSet
}

type EditInput struct {
	ActorID  shared.UserID
	Existing domainrole.Role
	Patch    EditPatch
	Access   membership.EffectivePermissions
	Now      time.Time
}

type EditResult struct {
	Role  domainrole.Role
	Event eventlog.Entry
}

func Edit(input EditInput) (EditResult, error) {
	if !input.Access.Has(domainrole.SystemAdmin, false) {
		return EditResult{}, shared.ErrForbidden
	}

	updated := input.Existing
	changed := make([]string, 0, 2)

	if input.Patch.Name != nil {
		updated.Name = *input.Patch.Name
		changed = append(changed, "name")
	}
	if input.Patch.Permissions != nil {
		updated.Permissions = *input.Patch.Permissions
		changed = append(changed, "permissions")
	}
	updated.UpdatedAt = input.Now.UTC()
	if input.Now.IsZero() {
		updated.UpdatedAt = time.Now().UTC()
	}
	updated.Normalize()
	if err := updated.Validate(); err != nil {
		return EditResult{}, err
	}

	event, err := roleEvent(input.ActorID, eventlog.EventRoleEdited, updated.ID, map[string]string{
		"role_id":        string(updated.ID),
		"name":           updated.Name,
		"changed_fields": strings.Join(changed, ","),
	}, input.Now)
	if err != nil {
		return EditResult{}, err
	}
	return EditResult{Role: updated, Event: event}, nil
}

// — Delete ——————————————————————————————————————————————————————————————————

type DeleteInput struct {
	ActorID shared.UserID
	RoleID  shared.RoleID
	Access  membership.EffectivePermissions
}

// Delete checks whether the actor may delete a role.
// Only SystemAdmin scope is permitted to delete roles.
func Delete(input DeleteInput) error {
	if !input.Access.Has(domainrole.SystemAdmin, false) {
		return shared.ErrForbidden
	}
	return nil
}

// — shared event helper —————————————————————————————————————————————————————

func roleEvent(
	actorID shared.UserID,
	typ eventlog.EventType,
	roleID shared.RoleID,
	payload map[string]string,
	now time.Time,
) (eventlog.Entry, error) {
	p, err := eventlog.NewPayload(payload)
	if err != nil {
		return eventlog.Entry{}, err
	}
	return eventlog.NewEntry(actorID, typ, eventlog.SubjectRole, string(roleID), p, now)
}
