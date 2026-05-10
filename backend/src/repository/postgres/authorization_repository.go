package postgres

import (
	"context"

	domainmembership "activist-base/src/domain/membership"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	"activist-base/src/repository/postgres/sqlc"
)

type AuthorizationRepository struct {
	queries sqlc.Querier
}

func NewAuthorizationRepository(queries sqlc.Querier) *AuthorizationRepository {
	return &AuthorizationRepository{queries: queries}
}

func (r *AuthorizationRepository) ListActorMembershipContexts(
	ctx context.Context,
	actorID shared.UserID,
) ([]domainmembership.MembershipContext, error) {
	rows, err := r.queries.ListPermissionContextsByUser(ctx, string(actorID))
	if err != nil {
		return nil, err
	}

	contexts := make([]domainmembership.MembershipContext, 0, len(rows))
	for _, row := range rows {
		permissions, err := domainrole.PermissionSetFromBytes(row.Permissions)
		if err != nil {
			return nil, err
		}
		contexts = append(contexts, domainmembership.MembershipContext{
			DivisionID:  shared.DivisionID(row.DivisionID),
			Permissions: permissions,
		})
	}
	return contexts, nil
}

func (r *AuthorizationRepository) ListTargetDivisionAncestors(
	ctx context.Context,
	targetDivisionID shared.DivisionID,
) ([]shared.DivisionID, error) {
	rows, err := r.queries.GetDivisionAncestors(ctx, string(targetDivisionID))
	if err != nil {
		return nil, err
	}

	ancestors := make([]shared.DivisionID, 0, len(rows))
	for _, row := range rows {
		ancestors = append(ancestors, shared.DivisionID(row.ID))
	}
	return ancestors, nil
}

func (r *AuthorizationRepository) ListTargetUserDivisionIDs(
	ctx context.Context,
	targetUserID shared.UserID,
) ([]shared.DivisionID, error) {
	rows, err := r.queries.ListActiveDivisionIDsByUser(ctx, string(targetUserID))
	if err != nil {
		return nil, err
	}

	divisionIDs := make([]shared.DivisionID, 0, len(rows))
	for _, row := range rows {
		divisionIDs = append(divisionIDs, shared.DivisionID(row))
	}
	return divisionIDs, nil
}
