package division

import "activist-base/src/domain/shared"

var (
	ErrRootExists       = &shared.Error{Code: "division.root_already_exists", Message: "tree can have exactly one root division"}
	ErrSelfParent       = &shared.Error{Code: "division.self_parent_not_allowed", Message: "division cannot be its own parent"}
	ErrCycleOnMove      = &shared.Error{Code: "division.cycle_detected_on_move", Message: "moving division would create a cycle"}
	ErrArchived         = &shared.Error{Code: "division.archived", Message: "division is archived"}
	ErrArchivedParent   = &shared.Error{Code: "division.archived_parent_not_allowed", Message: "parent division is archived"}
	ErrInvalidTreeDepth = &shared.Error{Code: "division.invalid_tree_depth", Message: "tree depth must be positive"}
)

type Division struct {
	ID             shared.DivisionID
	ShortName      string
	FullName       string
	Description    string
	ParentID       *shared.DivisionID
	RegulationURL  string
	MediaLinks     []shared.Link
	IsArchived     bool
	PositionsCount int
	MembersCount   int
}

// Normalize trims whitespace from all string fields in place.
// Call before Validate.
func (d *Division) Normalize() {
	d.ShortName = shared.Trim(d.ShortName)
	d.FullName = shared.Trim(d.FullName)
	d.Description = shared.Trim(d.Description)
	d.RegulationURL = shared.Trim(d.RegulationURL)
	shared.NormalizeLinks(d.MediaLinks)
}

// Validate checks structural invariants. Assumes fields have been normalized.
func (d *Division) Validate() error {
	if err := shared.RequireID("division.id", string(d.ID)); err != nil {
		return err
	}
	if err := shared.RequireText("division.short_name", d.ShortName, shared.MaxName); err != nil {
		return err
	}
	if err := shared.AllowText("division.full_name", d.FullName, shared.MaxName); err != nil {
		return err
	}
	if err := shared.AllowText("division.description", d.Description, shared.MaxDescription); err != nil {
		return err
	}
	if err := shared.AllowText("division.regulation_url", d.RegulationURL, shared.MaxURL); err != nil {
		return err
	}
	if d.ParentID != nil {
		if err := shared.RequireID("division.parent_id", string(*d.ParentID)); err != nil {
			return err
		}
		if *d.ParentID == d.ID {
			return ErrSelfParent
		}
	}
	return shared.ValidateLinks(d.MediaLinks)
}

func (d *Division) IsRoot() bool { return d.ParentID == nil }

func (d *Division) CanAcceptChild() error {
	if d.IsArchived {
		return ErrArchivedParent
	}
	return nil
}

// ParentRef constructs a *DivisionID for use in Division.ParentID.
func ParentRef(id shared.DivisionID) *shared.DivisionID { return &id }
