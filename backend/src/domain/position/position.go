package position

import "activist-base/src/domain/shared"

var (
	ErrInvalidMaxCount = &shared.Error{Code: "position.invalid_max_count", Message: "position max_count must be positive when set"}
	ErrArchived        = &shared.Error{Code: "position.archived", Message: "position is archived"}
	ErrLimitReached    = &shared.Error{Code: "position.limit_reached", Message: "position max_count has been reached"}
)

type Position struct {
	ID         shared.PositionID
	Title      string
	RoleID     shared.RoleID
	DivisionID shared.DivisionID
	MaxCount   *int
	IsArchived bool
}

// Normalize trims whitespace from all string fields in place.
// Call before Validate.
func (p *Position) Normalize() {
	p.Title = shared.Trim(p.Title)
}

// Validate checks structural invariants. Assumes fields have been normalized.
func (p Position) Validate() error {
	if err := shared.RequireID("position.id", string(p.ID)); err != nil {
		return err
	}
	if err := shared.RequireID("position.role_id", string(p.RoleID)); err != nil {
		return err
	}
	if err := shared.RequireID("position.division_id", string(p.DivisionID)); err != nil {
		return err
	}
	if err := shared.RequireText("position.title", p.Title, shared.MaxName); err != nil {
		return err
	}
	if p.MaxCount != nil && *p.MaxCount <= 0 {
		return ErrInvalidMaxCount
	}
	return nil
}

// CanAssign returns an error if the position cannot accept a new assignment.
func (p Position) CanAssign(currentCount int) error {
	if p.IsArchived {
		return ErrArchived
	}
	if p.MaxCount != nil && currentCount >= *p.MaxCount {
		return ErrLimitReached
	}
	return nil
}

// MaxCountOf constructs a *int for use in Position.MaxCount.
func MaxCountOf(n int) *int { return &n }
