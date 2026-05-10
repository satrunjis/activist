package eventlog

import (
	domaineventlog "activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
)

type ListInput struct {
	EventType   *domaineventlog.EventType
	SubjectType *domaineventlog.SubjectType
	SubjectID   *string
	Limit       int
	Offset      int
}

type ListResult struct {
	Items  []domaineventlog.Entry
	Total  int
	Limit  int
	Offset int
}

// CheckListAdmissibility returns an error if the actor's access does not permit
// viewing the event log. CanViewAuditLog permission is required.
func CheckListAdmissibility(access domainmembership.EffectivePermissions) error {
	if !access.Has(domainrole.CanViewAuditLog, false) {
		return shared.ErrForbidden
	}
	return nil
}

func NormalizeListInput(input ListInput) (ListInput, error) {
	if input.Offset < 0 {
		return ListInput{}, &shared.Error{
			Code:    "validation.offset",
			Message: "offset must be zero or positive",
		}
	}

	switch {
	case input.Limit == 0:
		input.Limit = 50
	case input.Limit < 0:
		return ListInput{}, &shared.Error{
			Code:    "validation.limit",
			Message: "limit must be positive",
		}
	case input.Limit > 200:
		input.Limit = 200
	}

	return input, nil
}
