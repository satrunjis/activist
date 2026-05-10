package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
	"activist-base/src/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5"
)

// UserProfileRepository wraps sqlc profile read/update queries for USER-01.
type UserProfileRepository struct {
	queries *sqlc.Queries
}

func NewUserProfileRepository(queries *sqlc.Queries) *UserProfileRepository {
	return &UserProfileRepository{queries: queries}
}

func (r *UserProfileRepository) GetUserByID(ctx context.Context, id shared.UserID) (domainuser.User, error) {
	persisted, err := r.queries.GetUserByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainuser.User{}, sql.ErrNoRows
		}
		return domainuser.User{}, err
	}
	return toDomainUser(persisted)
}

func (r *UserProfileRepository) UpdateProfile(ctx context.Context, user domainuser.User) (domainuser.User, error) {
	socialLinks, err := json.Marshal(user.SocialLinks)
	if err != nil {
		return domainuser.User{}, err
	}

	persisted, err := r.queries.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		ID:              string(user.ID),
		FirstName:       user.FirstName,
		LastName:        textOrNull(user.LastName),
		MiddleName:      textOrNull(user.MiddleName),
		GradebookNumber: textOrNull(user.GradebookNumber),
		GroupNumber:     textOrNull(user.GroupNumber),
		Institute:       textOrNull(user.Institute),
		BirthDate:       dateOrNull(user.BirthDate),
		Phone:           textOrNull(user.Phone),
		SocialLinks:     socialLinks,
		About:           textOrNull(user.About),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainuser.User{}, sql.ErrNoRows
		}
		return domainuser.User{}, err
	}
	return toDomainUser(persisted)
}
