package division

import (
	"strconv"
	"strings"
	"time"

	domaindivision "activist-base/src/domain/division"
	"activist-base/src/domain/eventlog"
	"activist-base/src/domain/membership"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
)

// — Create ——————————————————————————————————————————————————————————————————

type CreateInput struct {
	ActorID        shared.UserID
	ID             shared.DivisionID
	Parent         *domaindivision.Division
	ExistingRootID *shared.DivisionID
	ShortName      string
	FullName       string
	Description    string
	RegulationURL  string
	MediaLinks     []shared.Link
	Access         membership.EffectivePermissions
	Now            time.Time
}

type CreateResult struct {
	Division domaindivision.Division
	Event    eventlog.Entry
}

func Create(input CreateInput) (CreateResult, error) {
	var parentID *shared.DivisionID

	if input.Parent == nil {
		if !input.Access.Has(role.SystemAdmin, false) {
			return CreateResult{}, shared.ErrForbidden
		}
		if err := domaindivision.ValidateCreateRoot(input.ExistingRootID); err != nil {
			return CreateResult{}, err
		}
	} else {
		if !input.Access.Has(role.CanCreateSubdivision, false) {
			return CreateResult{}, shared.ErrForbidden
		}
		if err := domaindivision.ValidateCreateChild(input.ID, *input.Parent); err != nil {
			return CreateResult{}, err
		}
		parentID = domaindivision.ParentRef(input.Parent.ID)
	}

	d := domaindivision.Division{
		ID:            input.ID,
		ShortName:     input.ShortName,
		FullName:      input.FullName,
		Description:   input.Description,
		RegulationURL: input.RegulationURL,
		ParentID:      parentID,
		MediaLinks:    append([]shared.Link(nil), input.MediaLinks...),
	}
	d.Normalize()
	if err := d.Validate(); err != nil {
		return CreateResult{}, err
	}

	parentIDStr := ""
	if d.ParentID != nil {
		parentIDStr = string(*d.ParentID)
	}
	event, err := divisionEvent(input.ActorID, eventlog.EventDivisionCreated, d.ID, map[string]string{
		"division_id": string(d.ID),
		"parent_id":   parentIDStr,
		"short_name":  d.ShortName,
		"full_name":   d.FullName,
	}, input.Now)
	if err != nil {
		return CreateResult{}, err
	}

	return CreateResult{Division: d, Event: event}, nil
}

// — Edit ————————————————————————————————————————————————————————————————————

type EditPatch struct {
	ShortName     *string
	FullName      *string
	Description   *string
	RegulationURL *string
	MediaLinks    *[]shared.Link
}

type EditInput struct {
	ActorID            shared.UserID
	Existing           domaindivision.Division
	Patch              EditPatch
	ReassignParent     bool
	NewParent          *domaindivision.Division
	NewParentAncestors []shared.DivisionID
	Access             membership.EffectivePermissions
	Now                time.Time
}

type EditResult struct {
	Division domaindivision.Division
	Event    eventlog.Entry
}

func Edit(input EditInput) (EditResult, error) {
	if !input.Access.Has(role.CanEditDivision, false) {
		return EditResult{}, shared.ErrForbidden
	}
	if input.Existing.IsArchived {
		return EditResult{}, domaindivision.ErrArchived
	}

	updated := input.Existing
	changed := make([]string, 0, 5)

	if input.Patch.ShortName != nil {
		updated.ShortName = *input.Patch.ShortName
		changed = append(changed, "short_name")
	}
	if input.Patch.FullName != nil {
		updated.FullName = *input.Patch.FullName
		changed = append(changed, "full_name")
	}
	if input.Patch.Description != nil {
		updated.Description = *input.Patch.Description
		changed = append(changed, "description")
	}
	if input.Patch.RegulationURL != nil {
		updated.RegulationURL = *input.Patch.RegulationURL
		changed = append(changed, "regulation_url")
	}
	if input.Patch.MediaLinks != nil {
		updated.MediaLinks = append([]shared.Link(nil), *input.Patch.MediaLinks...)
		changed = append(changed, "media_links")
	}
	if input.ReassignParent {
		if input.NewParent == nil {
			if !input.Access.Has(role.SystemAdmin, false) {
				return EditResult{}, shared.ErrForbidden
			}
			updated.ParentID = nil
		} else {
			if err := domaindivision.ValidateMove(updated.ID, *input.NewParent, input.NewParentAncestors); err != nil {
				return EditResult{}, err
			}
			updated.ParentID = domaindivision.ParentRef(input.NewParent.ID)
		}
		changed = append(changed, "parent_id")
	}
	updated.Normalize()
	if err := updated.Validate(); err != nil {
		return EditResult{}, err
	}

	event, err := divisionEvent(input.ActorID, eventlog.EventDivisionEdited, updated.ID, map[string]string{
		"division_id":    string(updated.ID),
		"changed_fields": strings.Join(changed, ","),
	}, input.Now)
	if err != nil {
		return EditResult{}, err
	}
	return EditResult{Division: updated, Event: event}, nil
}

// — Archive —————————————————————————————————————————————————————————————————

type ArchiveInput struct {
	ActorID                 shared.UserID
	Division                domaindivision.Division
	ArchivedPositionsCount  int
	RemovedMembershipsCount int
	Access                  membership.EffectivePermissions
	Now                     time.Time
}

type ArchiveResult struct {
	Division domaindivision.Division
	Event    eventlog.Entry
}

func Archive(input ArchiveInput) (ArchiveResult, error) {
	if !input.Access.Has(role.CanArchiveDivision, false) {
		return ArchiveResult{}, shared.ErrForbidden
	}
	if input.Division.IsArchived {
		return ArchiveResult{}, domaindivision.ErrArchived
	}

	archived := input.Division
	archived.IsArchived = true

	event, err := divisionEvent(input.ActorID, eventlog.EventDivisionArchived, archived.ID, map[string]string{
		"division_id":               string(archived.ID),
		"short_name":                archived.ShortName,
		"archived_positions_count":  strconv.Itoa(input.ArchivedPositionsCount),
		"removed_memberships_count": strconv.Itoa(input.RemovedMembershipsCount),
	}, input.Now)
	if err != nil {
		return ArchiveResult{}, err
	}
	return ArchiveResult{Division: archived, Event: event}, nil
}

// — Tree ————————————————————————————————————————————————————————————————————

type TreeNode struct {
	Division       domaindivision.Division
	Children       []TreeNode
	HasChildren    bool
	ChildrenCount  int
	PositionsCount int
	MembersCount   int
}

type TreeInput struct {
	Root             TreeNode
	Depth            int
	CanViewCard      bool
	IncludeAncestors bool
	Ancestors        []domaindivision.Division
}

type TreeAncestor struct {
	ID        shared.DivisionID
	ParentID  *shared.DivisionID
	ShortName string
}

type TreeResult struct {
	ID             shared.DivisionID
	ParentID       *shared.DivisionID
	ShortName      string
	IsArchived     bool
	HasChildren    bool
	ChildrenCount  int
	PositionsCount int
	MembersCount   int
	Children       []TreeResult
	FullName       *string
	Description    *string
	RegulationURL  *string
	MediaLinks     []shared.Link
	Ancestors      []TreeAncestor
}

func GetTree(input TreeInput) (TreeResult, error) {
	if input.Depth <= 0 {
		return TreeResult{}, domaindivision.ErrInvalidTreeDepth
	}
	node := buildNode(input.Root, input.Depth, input.CanViewCard)
	if input.IncludeAncestors {
		node.Ancestors = make([]TreeAncestor, 0, len(input.Ancestors))
		for _, a := range input.Ancestors {
			node.Ancestors = append(node.Ancestors, TreeAncestor{
				ID:        a.ID,
				ParentID:  a.ParentID,
				ShortName: a.ShortName,
			})
		}
	}
	return node, nil
}

func buildNode(n TreeNode, depth int, canViewCard bool) TreeResult {
	childCount := n.ChildrenCount
	if childCount == 0 {
		childCount = len(n.Children)
	}
	result := TreeResult{
		ID:             n.Division.ID,
		ParentID:       n.Division.ParentID,
		ShortName:      n.Division.ShortName,
		IsArchived:     n.Division.IsArchived,
		HasChildren:    n.HasChildren || len(n.Children) > 0,
		ChildrenCount:  childCount,
		PositionsCount: n.Division.PositionsCount,
		MembersCount:   n.Division.MembersCount,
	}
	if canViewCard {
		result.FullName = strPtr(n.Division.FullName)
		result.Description = strPtr(n.Division.Description)
		result.RegulationURL = strPtr(n.Division.RegulationURL)
		result.MediaLinks = append([]shared.Link(nil), n.Division.MediaLinks...)
	}
	if depth > 1 {
		result.Children = make([]TreeResult, 0, len(n.Children))
		for _, child := range n.Children {
			result.Children = append(result.Children, buildNode(child, depth-1, canViewCard))
		}
	}
	return result
}

func strPtr(s string) *string { return &s }

// — shared event helper —————————————————————————————————————————————————————

func divisionEvent(
	actorID shared.UserID,
	typ eventlog.EventType,
	divisionID shared.DivisionID,
	payload map[string]string,
	now time.Time,
) (eventlog.Entry, error) {
	p, err := eventlog.NewPayload(payload)
	if err != nil {
		return eventlog.Entry{}, err
	}
	return eventlog.NewEntry(actorID, typ, eventlog.SubjectDivision, string(divisionID), p, now)
}
