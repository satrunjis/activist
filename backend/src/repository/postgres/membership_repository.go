package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	appeventlog "activist-base/src/application/eventlog"
	appmembership "activist-base/src/application/membership"
	appuser "activist-base/src/application/user"
	domaindivision "activist-base/src/domain/division"
	domainmembership "activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
	"activist-base/src/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MembershipRepository struct {
	queries *sqlc.Queries
	nowFn   func() time.Time
}

func NewMembershipRepository(queries *sqlc.Queries, pool *pgxpool.Pool) *MembershipRepository {
	_ = pool
	return &MembershipRepository{
		queries: queries,
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (r *MembershipRepository) GetPositionByID(ctx context.Context, id shared.PositionID) (domainposition.Position, error) {
	persisted, err := r.queries.GetPositionByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainposition.Position{}, sql.ErrNoRows
		}
		return domainposition.Position{}, err
	}
	return toDomainPosition(persisted), nil
}

func (r *MembershipRepository) GetRoleByID(ctx context.Context, id shared.RoleID) (domainrole.Role, error) {
	persisted, err := r.queries.GetRoleByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainrole.Role{}, sql.ErrNoRows
		}
		return domainrole.Role{}, err
	}
	return toDomainRole(persisted)
}

func (r *MembershipRepository) GetUserByID(ctx context.Context, id shared.UserID) (domainuser.User, error) {
	persisted, err := r.queries.GetUserByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainuser.User{}, sql.ErrNoRows
		}
		return domainuser.User{}, err
	}
	return toDomainUser(persisted)
}

func (r *MembershipRepository) AssignWithCountCheck(
	ctx context.Context,
	userID shared.UserID,
	position domainposition.Position,
	access domainmembership.EffectivePermissions,
	actorID shared.UserID,
) (appmembership.AssignResult, error) {
	return r.assignWithCountCheckQueries(ctx, r.queries, userID, position, access, actorID)
}

func (r *MembershipRepository) AssignWithCountCheckWithTx(
	ctx context.Context,
	tx appeventlog.Transaction,
	userID shared.UserID,
	position domainposition.Position,
	access domainmembership.EffectivePermissions,
	actorID shared.UserID,
) (appmembership.AssignResult, error) {
	pgxTx, err := unwrapPGXTx(tx)
	if err != nil {
		return appmembership.AssignResult{}, err
	}
	return r.assignWithCountCheckQueries(ctx, r.queries.WithTx(pgxTx), userID, position, access, actorID)
}

func (r *MembershipRepository) assignWithCountCheckQueries(
	ctx context.Context,
	queries *sqlc.Queries,
	userID shared.UserID,
	position domainposition.Position,
	access domainmembership.EffectivePermissions,
	actorID shared.UserID,
) (appmembership.AssignResult, error) {
	lockedPositionRow, err := queries.GetPositionByIDForUpdate(ctx, string(position.ID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return appmembership.AssignResult{}, sql.ErrNoRows
		}
		return appmembership.AssignResult{}, err
	}
	lockedPosition := toDomainPosition(lockedPositionRow)

	currentCount, err := queries.CountMembersByPosition(ctx, string(position.ID))
	if err != nil {
		return appmembership.AssignResult{}, err
	}
	alreadyAssigned, err := queries.MembershipExists(ctx, sqlc.MembershipExistsParams{
		UserID:     string(userID),
		PositionID: string(position.ID),
	})
	if err != nil {
		return appmembership.AssignResult{}, err
	}

	persistedUser, err := queries.GetUserByID(ctx, string(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return appmembership.AssignResult{}, sql.ErrNoRows
		}
		return appmembership.AssignResult{}, err
	}
	user, err := toDomainUser(persistedUser)
	if err != nil {
		return appmembership.AssignResult{}, err
	}

	assignResult, err := appmembership.Assign(appmembership.AssignInput{
		ActorID:              actorID,
		User:                 user,
		Position:             lockedPosition,
		CurrentPositionCount: int(currentCount),
		AlreadyAssigned:      alreadyAssigned,
		Access:               access,
		Now:                  r.nowFn(),
	})
	if err != nil {
		return appmembership.AssignResult{}, err
	}

	if err := queries.CreateMembership(ctx, sqlc.CreateMembershipParams{
		UserID:     string(userID),
		PositionID: string(position.ID),
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return appmembership.AssignResult{}, domainmembership.ErrAlreadyExists
		}
		return appmembership.AssignResult{}, err
	}

	return assignResult, nil
}

func (r *MembershipRepository) Remove(
	ctx context.Context,
	userID shared.UserID,
	positionID shared.PositionID,
	access domainmembership.EffectivePermissions,
	actorID shared.UserID,
) (appmembership.RemoveResult, error) {
	return r.removeWithQueries(ctx, r.queries, userID, positionID, access, actorID)
}

func (r *MembershipRepository) RemoveWithTx(
	ctx context.Context,
	tx appeventlog.Transaction,
	userID shared.UserID,
	positionID shared.PositionID,
	access domainmembership.EffectivePermissions,
	actorID shared.UserID,
) (appmembership.RemoveResult, error) {
	pgxTx, err := unwrapPGXTx(tx)
	if err != nil {
		return appmembership.RemoveResult{}, err
	}
	return r.removeWithQueries(ctx, r.queries.WithTx(pgxTx), userID, positionID, access, actorID)
}

func (r *MembershipRepository) removeWithQueries(
	ctx context.Context,
	queries *sqlc.Queries,
	userID shared.UserID,
	positionID shared.PositionID,
	access domainmembership.EffectivePermissions,
	actorID shared.UserID,
) (appmembership.RemoveResult, error) {
	positionRow, err := queries.GetPositionByID(ctx, string(positionID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return appmembership.RemoveResult{}, sql.ErrNoRows
		}
		return appmembership.RemoveResult{}, err
	}

	exists, err := queries.MembershipExists(ctx, sqlc.MembershipExistsParams{
		UserID:     string(userID),
		PositionID: string(positionID),
	})
	if err != nil {
		return appmembership.RemoveResult{}, err
	}

	removeResult, err := appmembership.Remove(appmembership.RemoveInput{
		ActorID: actorID,
		Membership: domainmembership.Membership{
			UserID:     userID,
			PositionID: positionID,
		},
		Position: toDomainPosition(positionRow),
		Exists:   exists,
		Access:   access,
		Now:      r.nowFn(),
	})
	if err != nil {
		return appmembership.RemoveResult{}, err
	}

	if err := queries.DeleteMembership(ctx, sqlc.DeleteMembershipParams{
		UserID:     string(userID),
		PositionID: string(positionID),
	}); err != nil {
		return appmembership.RemoveResult{}, err
	}
	return removeResult, nil
}

func (r *MembershipRepository) ListByUser(ctx context.Context, userID shared.UserID) ([]appuser.MembershipView, error) {
	rows, err := r.queries.ListMembershipsByUser(ctx, string(userID))
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

func (r *MembershipRepository) ListMembersByPosition(ctx context.Context, positionID shared.PositionID) ([]appuser.MembershipView, error) {
	rows, err := r.queries.ListMembersByPosition(ctx, string(positionID))
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
