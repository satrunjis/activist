package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	appauth "activist-base/src/application/auth"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
	"activist-base/src/repository/postgres/sqlc"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

// AuthRepository wraps sqlc queries for auth-specific persistence flows.
type AuthRepository struct {
	queries *sqlc.Queries
}

func NewAuthRepository(queries *sqlc.Queries) *AuthRepository {
	return &AuthRepository{queries: queries}
}

func (r *AuthRepository) CreateUser(ctx context.Context, user domainuser.User) (domainuser.User, error) {
	socialLinks, err := json.Marshal(user.SocialLinks)
	if err != nil {
		return domainuser.User{}, fmt.Errorf("marshal social links: %w", err)
	}

	created, err := r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:              string(user.ID),
		Login:           user.Login,
		PasswordHash:    user.PasswordHash,
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
		if mapped := mapCreateUserError(err); mapped != nil {
			return domainuser.User{}, mapped
		}
		return domainuser.User{}, err
	}

	return toDomainUser(created)
}

func mapCreateUserError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	if pgErr.Code != "23505" {
		return nil
	}

	switch pgErr.ConstraintName {
	case "users_login_uq":
		return &shared.Error{
			Code:    "validation.login_taken",
			Message: "login is already taken",
		}
	case "users_gradebook_number_uq":
		return &shared.Error{
			Code:    "validation.gradebook_number_taken",
			Message: "gradebook_number is already taken",
		}
	default:
		return nil
	}
}

func (r *AuthRepository) GetUserByLogin(ctx context.Context, login string) (domainuser.User, error) {
	persisted, err := r.queries.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainuser.User{}, sql.ErrNoRows
		}
		return domainuser.User{}, err
	}
	return toDomainUser(persisted)
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id shared.UserID) (domainuser.User, error) {
	persisted, err := r.queries.GetUserByID(ctx, string(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainuser.User{}, sql.ErrNoRows
		}
		return domainuser.User{}, err
	}
	return toDomainUser(persisted)
}

func (r *AuthRepository) CreateSession(ctx context.Context, params appauth.CreateSessionParams) (appauth.SessionRecord, error) {
	created, err := r.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:                params.ID,
		UserID:            string(params.UserID),
		TokenHash:         params.TokenHash,
		CsrfSecret:        params.CSRFSecret,
		AbsoluteExpiresAt: timestamptz(params.AbsoluteExpiresAt),
		IdleExpiresAt:     timestamptz(params.IdleExpiresAt),
		CreatedIp:         params.CreatedIP,
		CreatedUserAgent:  params.CreatedUserAgent,
	})
	if err != nil {
		return appauth.SessionRecord{}, err
	}
	return toSessionRecord(created)
}

func (r *AuthRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (appauth.SessionRecord, error) {
	persisted, err := r.queries.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return appauth.SessionRecord{}, sql.ErrNoRows
		}
		return appauth.SessionRecord{}, err
	}
	return toSessionRecord(persisted)
}

func (r *AuthRepository) TouchSession(ctx context.Context, params appauth.TouchSessionParams) (appauth.SessionRecord, error) {
	persisted, err := r.queries.TouchSession(ctx, sqlc.TouchSessionParams{
		ID:            params.ID,
		LastSeenAt:    timestamptz(params.LastSeenAt),
		IdleExpiresAt: timestamptz(params.IdleExpiresAt),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return appauth.SessionRecord{}, sql.ErrNoRows
		}
		return appauth.SessionRecord{}, err
	}
	return toSessionRecord(persisted)
}

func (r *AuthRepository) RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time, reason string) (bool, error) {
	rowsAffected, err := r.queries.RevokeSession(ctx, sqlc.RevokeSessionParams{
		ID:            sessionID,
		RevokedAt:     timestamptz(revokedAt),
		RevokedReason: textOrNull(reason),
	})
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (r *AuthRepository) RevokeAllUserSessions(ctx context.Context, userID shared.UserID, revokedAt time.Time, reason string) (int64, error) {
	rowsAffected, err := r.queries.RevokeAllUserSessions(ctx, sqlc.RevokeAllUserSessionsParams{
		UserID:        string(userID),
		RevokedAt:     timestamptz(revokedAt),
		RevokedReason: textOrNull(reason),
	})
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func toDomainUser(persisted sqlc.User) (domainuser.User, error) {
	var socialLinks []shared.Link
	if len(persisted.SocialLinks) > 0 {
		if err := json.Unmarshal(persisted.SocialLinks, &socialLinks); err != nil {
			return domainuser.User{}, fmt.Errorf("unmarshal social links: %w", err)
		}
	}

	return domainuser.User{
		ID:              shared.UserID(persisted.ID),
		Login:           persisted.Login,
		PasswordHash:    persisted.PasswordHash,
		FirstName:       persisted.FirstName,
		LastName:        textFromPg(persisted.LastName),
		MiddleName:      textFromPg(persisted.MiddleName),
		GradebookNumber: textFromPg(persisted.GradebookNumber),
		GroupNumber:     textFromPg(persisted.GroupNumber),
		Institute:       textFromPg(persisted.Institute),
		BirthDate:       dateFromPg(persisted.BirthDate),
		Phone:           textFromPg(persisted.Phone),
		SocialLinks:     socialLinks,
		About:           textFromPg(persisted.About),
	}, nil
}

func toSessionRecord(persisted sqlc.Session) (appauth.SessionRecord, error) {
	createdAt, err := timeFromTimestamp(persisted.CreatedAt)
	if err != nil {
		return appauth.SessionRecord{}, err
	}
	lastSeenAt, err := timeFromTimestamp(persisted.LastSeenAt)
	if err != nil {
		return appauth.SessionRecord{}, err
	}
	absoluteExpiresAt, err := timeFromTimestamp(persisted.AbsoluteExpiresAt)
	if err != nil {
		return appauth.SessionRecord{}, err
	}
	idleExpiresAt, err := timeFromTimestamp(persisted.IdleExpiresAt)
	if err != nil {
		return appauth.SessionRecord{}, err
	}

	var revokedAt *time.Time
	if persisted.RevokedAt.Valid {
		t, err := timeFromTimestamp(persisted.RevokedAt)
		if err != nil {
			return appauth.SessionRecord{}, err
		}
		revokedAt = &t
	}

	return appauth.SessionRecord{
		ID:                persisted.ID,
		UserID:            shared.UserID(persisted.UserID),
		TokenHash:         persisted.TokenHash,
		CSRFSecret:        persisted.CsrfSecret,
		CreatedAt:         createdAt,
		LastSeenAt:        lastSeenAt,
		AbsoluteExpiresAt: absoluteExpiresAt,
		IdleExpiresAt:     idleExpiresAt,
		RevokedAt:         revokedAt,
		RevokedReason:     textFromPg(persisted.RevokedReason),
		CreatedIP:         persisted.CreatedIp,
		CreatedUserAgent:  persisted.CreatedUserAgent,
	}, nil
}

func textOrNull(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: value, Valid: true}
}

func textFromPg(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func dateOrNull(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: value.UTC(), Valid: true}
}

func dateFromPg(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}
	t := value.Time.UTC()
	return &t
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func timeFromTimestamp(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, fmt.Errorf("timestamp is not valid")
	}
	return value.Time.UTC(), nil
}
