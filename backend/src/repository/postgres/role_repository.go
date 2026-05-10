package postgres

import (
	"context"
	"database/sql"
	"errors"

	appeventlog "activist-base/src/application/eventlog"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	"activist-base/src/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type RoleRepository struct {
	queries *sqlc.Queries
}

func NewRoleRepository(queries *sqlc.Queries) *RoleRepository {
	return &RoleRepository{queries: queries}
}

func (r *RoleRepository) Create(ctx context.Context, role domainrole.Role) (domainrole.Role, error) {
	return r.createWithQueries(ctx, r.queries, role)
}

func (r *RoleRepository) CreateWithTx(ctx context.Context, tx appeventlog.Transaction, role domainrole.Role) (domainrole.Role, error) {
	pgxTx, err := unwrapPGXTx(tx)
	if err != nil {
		return domainrole.Role{}, err
	}
	return r.createWithQueries(ctx, r.queries.WithTx(pgxTx), role)
}

func (r *RoleRepository) createWithQueries(ctx context.Context, queries *sqlc.Queries, role domainrole.Role) (domainrole.Role, error) {
	persisted, err := queries.CreateRole(ctx, sqlc.CreateRoleParams{
		ID:          string(role.ID),
		Name:        role.Name,
		Permissions: role.Permissions.MarshalBinary(),
		CreatedAt:   timestamptz(role.CreatedAt),
		UpdatedAt:   timestamptz(role.UpdatedAt),
	})
	if err != nil {
		return domainrole.Role{}, err
	}
	return toDomainRole(persisted)
}

func (r *RoleRepository) GetByID(ctx context.Context, id shared.RoleID) (domainrole.Role, error) {
	persisted, err := r.queries.GetRoleByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrole.Role{}, sql.ErrNoRows
		}
		return domainrole.Role{}, err
	}
	return toDomainRole(persisted)
}

func (r *RoleRepository) List(ctx context.Context, limit, offset int) ([]domainrole.Role, int64, error) {
	total, err := r.queries.CountRoles(ctx)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.queries.ListRoles(ctx, sqlc.ListRolesParams{
		PageLimit:  int32(limit),
		PageOffset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	items := make([]domainrole.Role, 0, len(rows))
	for _, row := range rows {
		item, err := toDomainRole(row)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *RoleRepository) Update(ctx context.Context, role domainrole.Role) (domainrole.Role, error) {
	return r.updateWithQueries(ctx, r.queries, role)
}

func (r *RoleRepository) UpdateWithTx(ctx context.Context, tx appeventlog.Transaction, role domainrole.Role) (domainrole.Role, error) {
	pgxTx, err := unwrapPGXTx(tx)
	if err != nil {
		return domainrole.Role{}, err
	}
	return r.updateWithQueries(ctx, r.queries.WithTx(pgxTx), role)
}

func (r *RoleRepository) updateWithQueries(ctx context.Context, queries *sqlc.Queries, role domainrole.Role) (domainrole.Role, error) {
	persisted, err := queries.UpdateRole(ctx, sqlc.UpdateRoleParams{
		ID:          string(role.ID),
		Name:        role.Name,
		Permissions: role.Permissions.MarshalBinary(),
		UpdatedAt:   timestamptz(role.UpdatedAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrole.Role{}, sql.ErrNoRows
		}
		return domainrole.Role{}, err
	}
	return toDomainRole(persisted)
}

func (r *RoleRepository) Delete(ctx context.Context, id shared.RoleID) error {
	err := r.queries.DeleteRole(ctx, string(id))
	if err != nil {
		if isPgFKViolation(err) {
			return &shared.Error{
				Code:    "validation.role_in_use",
				Message: "role is referenced by one or more positions and cannot be deleted",
			}
		}
		return err
	}
	return nil
}

func isPgFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503"
	}
	return false
}

func toDomainRole(persisted sqlc.Role) (domainrole.Role, error) {
	permissions, err := domainrole.PermissionSetFromBytes(persisted.Permissions)
	if err != nil {
		return domainrole.Role{}, err
	}

	createdAt, err := timeFromTimestamp(persisted.CreatedAt)
	if err != nil {
		return domainrole.Role{}, err
	}
	updatedAt, err := timeFromTimestamp(persisted.UpdatedAt)
	if err != nil {
		return domainrole.Role{}, err
	}
	return domainrole.Role{
		ID:          shared.RoleID(persisted.ID),
		Name:        persisted.Name,
		Permissions: permissions,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
}
