package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	appeventlog "activist-base/src/application/eventlog"
	domaindivision "activist-base/src/domain/division"
	"activist-base/src/domain/shared"
	"activist-base/src/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type DivisionRepository struct {
	queries *sqlc.Queries
}

func NewDivisionRepository(queries *sqlc.Queries) *DivisionRepository {
	return &DivisionRepository{queries: queries}
}

func (r *DivisionRepository) GetRoot(ctx context.Context) (domaindivision.Division, error) {
	persisted, err := r.queries.GetRootDivision(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaindivision.Division{}, sql.ErrNoRows
		}
		return domaindivision.Division{}, err
	}
	return toDomainDivision(persisted)
}

func (r *DivisionRepository) GetByID(ctx context.Context, id shared.DivisionID) (domaindivision.Division, error) {
	persisted, err := r.queries.GetDivisionByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaindivision.Division{}, sql.ErrNoRows
		}
		return domaindivision.Division{}, err
	}
	return toDomainDivision(persisted)
}

func (r *DivisionRepository) Create(ctx context.Context, division domaindivision.Division) (domaindivision.Division, error) {
	mediaLinks, err := json.Marshal(division.MediaLinks)
	if err != nil {
		return domaindivision.Division{}, err
	}

	persisted, err := r.queries.CreateDivision(ctx, sqlc.CreateDivisionParams{
		ID:            string(division.ID),
		ParentID:      toText(division.ParentID),
		ShortName:     division.ShortName,
		FullName:      division.FullName,
		Description:   division.Description,
		RegulationUrl: division.RegulationURL,
		MediaLinks:    mediaLinks,
		IsArchived:    division.IsArchived,
	})
	if err != nil {
		if mapped := mapDivisionMutationError(err); mapped != nil {
			return domaindivision.Division{}, mapped
		}
		return domaindivision.Division{}, err
	}
	return toDomainDivision(persisted)
}

func (r *DivisionRepository) Update(ctx context.Context, division domaindivision.Division) (domaindivision.Division, error) {
	mediaLinks, err := json.Marshal(division.MediaLinks)
	if err != nil {
		return domaindivision.Division{}, err
	}

	persisted, err := r.queries.UpdateDivision(ctx, sqlc.UpdateDivisionParams{
		ID:            string(division.ID),
		ParentID:      toText(division.ParentID),
		ShortName:     division.ShortName,
		FullName:      division.FullName,
		Description:   division.Description,
		RegulationUrl: division.RegulationURL,
		MediaLinks:    mediaLinks,
		IsArchived:    division.IsArchived,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaindivision.Division{}, sql.ErrNoRows
		}
		if mapped := mapDivisionMutationError(err); mapped != nil {
			return domaindivision.Division{}, mapped
		}
		return domaindivision.Division{}, err
	}
	return toDomainDivision(persisted)
}

func (r *DivisionRepository) GetAncestorIDs(ctx context.Context, id shared.DivisionID) ([]shared.DivisionID, error) {
	rows, err := r.queries.GetDivisionAncestors(ctx, string(id))
	if err != nil {
		return nil, err
	}
	ancestorIDs := make([]shared.DivisionID, 0, len(rows))
	for _, row := range rows {
		ancestorIDs = append(ancestorIDs, shared.DivisionID(row.ID))
	}
	return ancestorIDs, nil
}

func (r *DivisionRepository) ListTree(ctx context.Context, includeArchived bool) ([]domaindivision.Division, error) {
	rows, err := r.queries.ListDivisionTree(ctx, includeArchived)
	if err != nil {
		return nil, err
	}

	divisions := make([]domaindivision.Division, 0, len(rows))
	for _, row := range rows {
		division, err := toDomainDivisionTreeRow(row)
		if err != nil {
			return nil, err
		}
		divisions = append(divisions, division)
	}
	return divisions, nil
}

// DivisionChildItem is a flat division row with eager child-count metadata.
type DivisionChildItem struct {
	Division      domaindivision.Division
	HasChildren   bool
	ChildrenCount int
}

// ListChildren returns direct non-archived children of the given parent division
// ordered deterministically by short_name, id. Each row includes has_children
// and children_count so the caller can render expand affordances without
// fetching the full subtree.
func (r *DivisionRepository) ListChildren(ctx context.Context, parentID shared.DivisionID) ([]DivisionChildItem, error) {
	rows, err := r.queries.ListDivisionChildrenByParent(ctx, pgtype.Text{
		String: string(parentID),
		Valid:  true,
	})
	if err != nil {
		return nil, err
	}
	items := make([]DivisionChildItem, 0, len(rows))
	for _, row := range rows {
		division, err := toDivisionChildrenByParentRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, division)
	}
	return items, nil
}

// ListRootChildren returns top-level (root) divisions with has_children metadata.
// Root divisions are those with no parent.
func (r *DivisionRepository) ListRootChildren(ctx context.Context) ([]DivisionChildItem, error) {
	rows, err := r.queries.ListRootDivisions(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]DivisionChildItem, 0, len(rows))
	for _, row := range rows {
		division, err := toDivisionRootRow(row)
		if err != nil {
			return nil, err
		}
		items = append(items, division)
	}
	return items, nil
}

func (r *DivisionRepository) ArchiveCascadeWithTx(
	ctx context.Context,
	tx appeventlog.Transaction,
	id shared.DivisionID,
) (domaindivision.Division, int, int, bool, error) {
	pgxTx, err := unwrapPGXTx(tx)
	if err != nil {
		return domaindivision.Division{}, 0, 0, false, err
	}

	queries := r.queries.WithTx(pgxTx)
	subtree, err := queries.ListDivisionSubtree(ctx, string(id))
	if err != nil {
		return domaindivision.Division{}, 0, 0, false, err
	}
	if len(subtree) == 0 {
		return domaindivision.Division{}, 0, 0, false, sql.ErrNoRows
	}

	var root sqlc.ListDivisionSubtreeRow
	foundRoot := false
	for _, row := range subtree {
		if row.ID == string(id) {
			root = row
			foundRoot = true
			break
		}
	}
	if !foundRoot {
		return domaindivision.Division{}, 0, 0, false, sql.ErrNoRows
	}

	rootDivision, err := toDomainDivisionTreeSubtreeRow(root)
	if err != nil {
		return domaindivision.Division{}, 0, 0, false, err
	}
	if rootDivision.IsArchived {
		return rootDivision, 0, 0, false, nil
	}

	removedMemberships, err := queries.DeleteMembershipsByDivisionSubtree(ctx, string(id))
	if err != nil {
		return domaindivision.Division{}, 0, 0, false, err
	}

	archivedPositions, err := queries.ArchivePositionsByDivisionSubtree(ctx, string(id))
	if err != nil {
		return domaindivision.Division{}, 0, 0, false, err
	}

	if _, err := queries.ArchiveChildDivisionsByRoot(ctx, string(id)); err != nil {
		return domaindivision.Division{}, 0, 0, false, err
	}

	archivedRootRows, err := queries.ArchiveDivisionByID(ctx, string(id))
	if err != nil {
		return domaindivision.Division{}, 0, 0, false, err
	}
	if archivedRootRows == 0 {
		latestRoot, getErr := queries.GetDivisionByID(ctx, string(id))
		if getErr != nil {
			if errors.Is(getErr, pgx.ErrNoRows) {
				return domaindivision.Division{}, 0, 0, false, sql.ErrNoRows
			}
			return domaindivision.Division{}, 0, 0, false, getErr
		}
		latestDivision, convErr := toDomainDivision(latestRoot)
		if convErr != nil {
			return domaindivision.Division{}, 0, 0, false, convErr
		}
		return latestDivision, 0, 0, false, nil
	}

	latestRoot, err := queries.GetDivisionByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaindivision.Division{}, 0, 0, false, sql.ErrNoRows
		}
		return domaindivision.Division{}, 0, 0, false, err
	}
	latestDivision, err := toDomainDivision(latestRoot)
	if err != nil {
		return domaindivision.Division{}, 0, 0, false, err
	}

	return latestDivision, int(removedMemberships), int(archivedPositions), true, nil
}

func mapDivisionMutationError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}

	switch pgErr.Code {
	case "23505":
		return domaindivision.ErrRootExists
	case "23514":
		return domaindivision.ErrSelfParent
	case "23503":
		return &shared.Error{
			Code:    "not_found.division",
			Message: "division not found",
		}
	default:
		return nil
	}
}

func toDomainDivision(persisted sqlc.Division) (domaindivision.Division, error) {
	var mediaLinks []shared.Link
	if len(persisted.MediaLinks) > 0 {
		if err := json.Unmarshal(persisted.MediaLinks, &mediaLinks); err != nil {
			return domaindivision.Division{}, err
		}
	}

	var parentID *shared.DivisionID
	if persisted.ParentID.Valid {
		id := shared.DivisionID(persisted.ParentID.String)
		parentID = &id
	}
	return domaindivision.Division{
		ID:            shared.DivisionID(persisted.ID),
		ParentID:      parentID,
		ShortName:     persisted.ShortName,
		FullName:      persisted.FullName,
		Description:   persisted.Description,
		RegulationURL: persisted.RegulationUrl,
		MediaLinks:    mediaLinks,
		IsArchived:    persisted.IsArchived,
	}, nil
}

func toDomainDivisionTreeRow(persisted sqlc.ListDivisionTreeRow) (domaindivision.Division, error) {
	var mediaLinks []shared.Link
	if len(persisted.MediaLinks) > 0 {
		if err := json.Unmarshal(persisted.MediaLinks, &mediaLinks); err != nil {
			return domaindivision.Division{}, err
		}
	}

	var parentID *shared.DivisionID
	if persisted.ParentID.Valid {
		id := shared.DivisionID(persisted.ParentID.String)
		parentID = &id
	}
	return domaindivision.Division{
		ID:             shared.DivisionID(persisted.ID),
		ParentID:       parentID,
		ShortName:      persisted.ShortName,
		FullName:       persisted.FullName,
		Description:    persisted.Description,
		RegulationURL:  persisted.RegulationUrl,
		MediaLinks:     mediaLinks,
		IsArchived:     persisted.IsArchived,
		PositionsCount: int(persisted.PositionsCount),
		MembersCount:   int(persisted.MembersCount),
	}, nil
}

func toText(id *shared.DivisionID) pgtype.Text {
	if id == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(*id), Valid: true}
}

func toDivisionChildrenByParentRow(persisted sqlc.ListDivisionChildrenByParentRow) (DivisionChildItem, error) {
	var mediaLinks []shared.Link
	if len(persisted.MediaLinks) > 0 {
		if err := json.Unmarshal(persisted.MediaLinks, &mediaLinks); err != nil {
			return DivisionChildItem{}, err
		}
	}

	var parentID *shared.DivisionID
	if persisted.ParentID.Valid {
		id := shared.DivisionID(persisted.ParentID.String)
		parentID = &id
	}
	return DivisionChildItem{
		Division: domaindivision.Division{
			ID:            shared.DivisionID(persisted.ID),
			ParentID:      parentID,
			ShortName:     persisted.ShortName,
			FullName:      persisted.FullName,
			Description:   persisted.Description,
			RegulationURL: persisted.RegulationUrl,
			MediaLinks:    mediaLinks,
			IsArchived:    persisted.IsArchived,
		},
		HasChildren:   persisted.HasChildren,
		ChildrenCount: int(persisted.ChildrenCount),
	}, nil
}

func toDivisionRootRow(persisted sqlc.ListRootDivisionsRow) (DivisionChildItem, error) {
	var mediaLinks []shared.Link
	if len(persisted.MediaLinks) > 0 {
		if err := json.Unmarshal(persisted.MediaLinks, &mediaLinks); err != nil {
			return DivisionChildItem{}, err
		}
	}

	var parentID *shared.DivisionID
	if persisted.ParentID.Valid {
		id := shared.DivisionID(persisted.ParentID.String)
		parentID = &id
	}
	return DivisionChildItem{
		Division: domaindivision.Division{
			ID:            shared.DivisionID(persisted.ID),
			ParentID:      parentID,
			ShortName:     persisted.ShortName,
			FullName:      persisted.FullName,
			Description:   persisted.Description,
			RegulationURL: persisted.RegulationUrl,
			MediaLinks:    mediaLinks,
			IsArchived:    persisted.IsArchived,
		},
		HasChildren:   persisted.HasChildren,
		ChildrenCount: int(persisted.ChildrenCount),
	}, nil
}

func toDomainDivisionTreeSubtreeRow(persisted sqlc.ListDivisionSubtreeRow) (domaindivision.Division, error) {
	var mediaLinks []shared.Link
	if len(persisted.MediaLinks) > 0 {
		if err := json.Unmarshal(persisted.MediaLinks, &mediaLinks); err != nil {
			return domaindivision.Division{}, err
		}
	}

	var parentID *shared.DivisionID
	if persisted.ParentID.Valid {
		id := shared.DivisionID(persisted.ParentID.String)
		parentID = &id
	}
	return domaindivision.Division{
		ID:            shared.DivisionID(persisted.ID),
		ParentID:      parentID,
		ShortName:     persisted.ShortName,
		FullName:      persisted.FullName,
		Description:   persisted.Description,
		RegulationURL: persisted.RegulationUrl,
		MediaLinks:    mediaLinks,
		IsArchived:    persisted.IsArchived,
	}, nil
}
