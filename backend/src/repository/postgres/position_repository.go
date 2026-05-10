package postgres

import (
	"context"
	"database/sql"
	"errors"

	appeventlog "activist-base/src/application/eventlog"
	appuser "activist-base/src/application/user"
	domaindivision "activist-base/src/domain/division"
	domainmembership "activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	"activist-base/src/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type PositionRepository struct {
	queries *sqlc.Queries
}

func NewPositionRepository(queries *sqlc.Queries, _ ...any) *PositionRepository {
	return &PositionRepository{queries: queries}
}

func (r *PositionRepository) Create(ctx context.Context, position domainposition.Position) (domainposition.Position, error) {
	return r.createWithQueries(ctx, r.queries, position)
}

func (r *PositionRepository) CreateWithTx(ctx context.Context, tx appeventlog.Transaction, position domainposition.Position) (domainposition.Position, error) {
	pgxTx, err := unwrapPGXTx(tx)
	if err != nil {
		return domainposition.Position{}, err
	}
	return r.createWithQueries(ctx, r.queries.WithTx(pgxTx), position)
}

func (r *PositionRepository) createWithQueries(ctx context.Context, queries *sqlc.Queries, position domainposition.Position) (domainposition.Position, error) {
	persisted, err := queries.CreatePosition(ctx, sqlc.CreatePositionParams{
		ID:         string(position.ID),
		Title:      position.Title,
		RoleID:     string(position.RoleID),
		DivisionID: string(position.DivisionID),
		MaxCount:   toInt4(position.MaxCount),
	})
	if err != nil {
		return domainposition.Position{}, err
	}
	return toDomainPosition(persisted), nil
}

func (r *PositionRepository) GetByID(ctx context.Context, id shared.PositionID) (domainposition.Position, error) {
	persisted, err := r.queries.GetPositionByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainposition.Position{}, sql.ErrNoRows
		}
		return domainposition.Position{}, err
	}
	return toDomainPosition(persisted), nil
}

func (r *PositionRepository) GetDivisionByID(ctx context.Context, id shared.DivisionID) (domaindivision.Division, error) {
	persisted, err := r.queries.GetDivisionByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domaindivision.Division{}, sql.ErrNoRows
		}
		return domaindivision.Division{}, err
	}
	return toDomainDivision(persisted)
}

func (r *PositionRepository) GetRoleByID(ctx context.Context, id shared.RoleID) (domainrole.Role, error) {
	persisted, err := r.queries.GetRoleByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrole.Role{}, sql.ErrNoRows
		}
		return domainrole.Role{}, err
	}
	return toDomainRole(persisted)
}

func (r *PositionRepository) ListByDivision(ctx context.Context, divisionID shared.DivisionID) ([]domainposition.Position, error) {
	rows, err := r.queries.ListPositionsByDivision(ctx, string(divisionID))
	if err != nil {
		return nil, err
	}

	items := make([]domainposition.Position, 0, len(rows))
	for _, row := range rows {
		items = append(items, toDomainPosition(row))
	}
	return items, nil
}

func (r *PositionRepository) CountMembers(ctx context.Context, id shared.PositionID) (int, error) {
	count, err := r.queries.CountMembersByPosition(ctx, string(id))
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *PositionRepository) ListMembersByPosition(ctx context.Context, id shared.PositionID) ([]appuser.MembershipView, error) {
	rows, err := r.queries.ListMembersByPosition(ctx, string(id))
	if err != nil {
		return nil, err
	}

	items := make([]appuser.MembershipView, 0, len(rows))
	for _, row := range rows {
		items = append(items, appuser.MembershipView{
			Membership: domainmembership.Membership{
				UserID:     shared.UserID(row.UserID),
				PositionID: shared.PositionID(row.PositionID),
			},
			Position: domainposition.Position{
				ID:         shared.PositionID(row.PositionID),
				Title:      row.PositionTitle,
				RoleID:     shared.RoleID(row.RoleID),
				DivisionID: shared.DivisionID(row.DivisionID),
				IsArchived: row.PositionArchived,
			},
			Division: domaindivision.Division{
				ID:         shared.DivisionID(row.DivisionID),
				ShortName:  row.DivisionShortName,
				IsArchived: row.DivisionArchived,
			},
			RoleName: row.RoleName,
		})
	}
	return items, nil
}

func (r *PositionRepository) Archive(ctx context.Context, id shared.PositionID) (domainposition.Position, int, error) {
	persisted, removed, err := r.archiveWithQueries(ctx, r.queries, id)
	if err != nil {
		return domainposition.Position{}, 0, err
	}
	return persisted, removed, nil
}

func (r *PositionRepository) ArchiveWithTx(ctx context.Context, tx appeventlog.Transaction, id shared.PositionID) (domainposition.Position, int, error) {
	pgxTx, err := unwrapPGXTx(tx)
	if err != nil {
		return domainposition.Position{}, 0, err
	}
	return r.archiveWithQueries(ctx, r.queries.WithTx(pgxTx), id)
}

func (r *PositionRepository) archiveWithQueries(ctx context.Context, queries *sqlc.Queries, id shared.PositionID) (domainposition.Position, int, error) {
	deletedRows, err := queries.DeleteMembershipsByPosition(ctx, string(id))
	if err != nil {
		return domainposition.Position{}, 0, err
	}

	archived, err := queries.ArchivePosition(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainposition.Position{}, 0, sql.ErrNoRows
		}
		return domainposition.Position{}, 0, err
	}

	return toDomainPosition(archived), int(deletedRows), nil
}

func toDomainPosition(persisted sqlc.Position) domainposition.Position {
	return domainposition.Position{
		ID:         shared.PositionID(persisted.ID),
		Title:      persisted.Title,
		RoleID:     shared.RoleID(persisted.RoleID),
		DivisionID: shared.DivisionID(persisted.DivisionID),
		MaxCount:   fromInt4(persisted.MaxCount),
		IsArchived: persisted.IsArchived,
	}
}

func toInt4(v *int) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*v), Valid: true}
}

func fromInt4(v pgtype.Int4) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int32)
	return &n
}
