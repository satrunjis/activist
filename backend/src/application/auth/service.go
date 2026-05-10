package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	domainauth "activist-base/src/domain/auth"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
)

var (
	ErrInvalidCredentials = &shared.Error{Code: "auth.invalid_credentials", Message: "invalid credentials"}
	ErrInvalidSession     = &shared.Error{Code: "auth.invalid_session", Message: "invalid session"}
	ErrRateLimited        = &shared.Error{Code: "auth.rate_limited", Message: "too many login attempts"}
)

type Config struct {
	SessionIdleTTL     time.Duration
	SessionAbsoluteTTL time.Duration
	MaxLoginFailures   int
	LoginBlockDuration time.Duration
}

type Repository interface {
	CreateUser(ctx context.Context, user domainuser.User) (domainuser.User, error)
	GetUserByLogin(ctx context.Context, login string) (domainuser.User, error)
	GetUserByID(ctx context.Context, id shared.UserID) (domainuser.User, error)
	CreateSession(ctx context.Context, params CreateSessionParams) (SessionRecord, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (SessionRecord, error)
	TouchSession(ctx context.Context, params TouchSessionParams) (SessionRecord, error)
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time, reason string) (bool, error)
	RevokeAllUserSessions(ctx context.Context, userID shared.UserID, revokedAt time.Time, reason string) (int64, error)
}

type Service struct {
	repo    Repository
	cfg     Config
	limiter *LoginLimiter
	nowFn   func() time.Time
}

type RegisterInput struct {
	Login           string
	Password        string
	FirstName       string
	GradebookNumber string
	GroupNumber     string
	Institute       string
	BirthDate       time.Time
	LastName        string
	MiddleName      string
	Phone           string
	About           string
	IP              string
	UserAgent       string
}

type LoginInput struct {
	Login     string
	Password  string
	IP        string
	UserAgent string
}

type GetSessionInput struct {
	Token string
}

type LogoutInput struct {
	Token string
}

type LogoutAllInput struct {
	UserID shared.UserID
}

type SessionView struct {
	ID            string
	Token         string
	ExpiresAt     time.Time
	IdleExpiresAt time.Time
}

type AuthResult struct {
	User    domainuser.User
	Session SessionView
}

type GetSessionResult struct {
	User    domainuser.User
	Session SessionView
}

type CreateSessionParams struct {
	ID                string
	UserID            shared.UserID
	TokenHash         string
	CSRFSecret        string
	AbsoluteExpiresAt time.Time
	IdleExpiresAt     time.Time
	CreatedIP         string
	CreatedUserAgent  string
}

type TouchSessionParams struct {
	ID            string
	LastSeenAt    time.Time
	IdleExpiresAt time.Time
}

type SessionRecord struct {
	ID                string
	UserID            shared.UserID
	TokenHash         string
	CSRFSecret        string
	CreatedAt         time.Time
	LastSeenAt        time.Time
	AbsoluteExpiresAt time.Time
	IdleExpiresAt     time.Time
	RevokedAt         *time.Time
	RevokedReason     string
	CreatedIP         string
	CreatedUserAgent  string
}

func NewService(cfg Config, repo Repository) *Service {
	if cfg.SessionIdleTTL <= 0 {
		cfg.SessionIdleTTL = 8 * time.Hour
	}
	if cfg.SessionAbsoluteTTL <= 0 {
		cfg.SessionAbsoluteTTL = 7 * 24 * time.Hour
	}
	svc := &Service{
		repo: repo,
		cfg:  cfg,
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}
	svc.limiter = NewLoginLimiter(cfg.MaxLoginFailures, cfg.LoginBlockDuration)
	return svc
}

func (s *Service) SetNowFunc(nowFn func() time.Time) {
	if nowFn == nil {
		return
	}
	s.nowFn = nowFn
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	if err := validateRegisterInput(input); err != nil {
		return AuthResult{}, err
	}

	passwordHash, err := domainauth.HashPassword(input.Password)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	userID, err := newRandomID()
	if err != nil {
		return AuthResult{}, fmt.Errorf("generate user id: %w", err)
	}

	birthDate := input.BirthDate.UTC()
	user := domainuser.User{
		ID:              shared.UserID(userID),
		Login:           shared.Trim(input.Login),
		PasswordHash:    passwordHash,
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		MiddleName:      input.MiddleName,
		GradebookNumber: input.GradebookNumber,
		GroupNumber:     input.GroupNumber,
		Institute:       input.Institute,
		BirthDate:       &birthDate,
		Phone:           input.Phone,
		About:           input.About,
	}
	user.Normalize()
	if err := user.Validate(); err != nil {
		return AuthResult{}, err
	}

	persistedUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return AuthResult{}, err
	}

	session, err := s.createSession(ctx, persistedUser.ID, input.IP, input.UserAgent)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: persistedUser, Session: session}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	normalizedLogin := shared.Trim(input.Login)
	if normalizedLogin == "" || input.Password == "" {
		return AuthResult{}, ErrInvalidCredentials
	}

	limiterKey := loginLimiterKey(normalizedLogin, input.IP)
	now := s.nowFn().UTC()
	if !s.limiter.Allow(limiterKey, now) {
		return AuthResult{}, ErrRateLimited
	}

	persistedUser, err := s.repo.GetUserByLogin(ctx, normalizedLogin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.limiter.RegisterFailure(limiterKey, now)
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}

	valid, err := domainauth.VerifyPassword(input.Password, persistedUser.PasswordHash)
	if err != nil || !valid {
		s.limiter.RegisterFailure(limiterKey, now)
		return AuthResult{}, ErrInvalidCredentials
	}

	s.limiter.Reset(limiterKey)
	session, err := s.createSession(ctx, persistedUser.ID, input.IP, input.UserAgent)
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: persistedUser, Session: session}, nil
}

func (s *Service) GetSession(ctx context.Context, input GetSessionInput) (GetSessionResult, error) {
	if input.Token == "" {
		return GetSessionResult{}, ErrInvalidSession
	}

	tokenHash := domainauth.HashToken(input.Token)
	record, err := s.repo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GetSessionResult{}, ErrInvalidSession
		}
		return GetSessionResult{}, err
	}

	now := s.nowFn().UTC()
	if record.RevokedAt != nil || domainauth.SessionExpired(now, record.IdleExpiresAt, record.AbsoluteExpiresAt) {
		_, _ = s.repo.RevokeSession(ctx, record.ID, now, "expired")
		return GetSessionResult{}, ErrInvalidSession
	}

	nextIdleExpiry := domainauth.NextIdleExpiry(now, s.cfg.SessionIdleTTL, record.AbsoluteExpiresAt)
	record, err = s.repo.TouchSession(ctx, TouchSessionParams{
		ID:            record.ID,
		LastSeenAt:    now,
		IdleExpiresAt: nextIdleExpiry,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GetSessionResult{}, ErrInvalidSession
		}
		return GetSessionResult{}, err
	}

	persistedUser, err := s.repo.GetUserByID(ctx, record.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GetSessionResult{}, ErrInvalidSession
		}
		return GetSessionResult{}, err
	}

	return GetSessionResult{
		User: persistedUser,
		Session: SessionView{
			ID:            record.ID,
			Token:         input.Token,
			ExpiresAt:     record.AbsoluteExpiresAt,
			IdleExpiresAt: record.IdleExpiresAt,
		},
	}, nil
}

func (s *Service) Logout(ctx context.Context, input LogoutInput) error {
	if input.Token == "" {
		return nil
	}

	tokenHash := domainauth.HashToken(input.Token)
	record, err := s.repo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	_, err = s.repo.RevokeSession(ctx, record.ID, s.nowFn().UTC(), "logout")
	return err
}

func (s *Service) LogoutAll(ctx context.Context, input LogoutAllInput) error {
	if input.UserID == "" {
		return shared.ErrEmptyID
	}
	_, err := s.repo.RevokeAllUserSessions(ctx, input.UserID, s.nowFn().UTC(), "logout_all")
	return err
}

func (s *Service) createSession(ctx context.Context, userID shared.UserID, ip, userAgent string) (SessionView, error) {
	token, err := domainauth.NewOpaqueToken()
	if err != nil {
		return SessionView{}, fmt.Errorf("new opaque token: %w", err)
	}
	csrfSecret, err := domainauth.NewOpaqueToken()
	if err != nil {
		return SessionView{}, fmt.Errorf("new csrf secret: %w", err)
	}
	sessionID, err := newRandomID()
	if err != nil {
		return SessionView{}, fmt.Errorf("generate session id: %w", err)
	}
	now := s.nowFn().UTC()
	absoluteExpiresAt := now.Add(s.cfg.SessionAbsoluteTTL)
	idleExpiresAt := domainauth.NextIdleExpiry(now, s.cfg.SessionIdleTTL, absoluteExpiresAt)

	record, err := s.repo.CreateSession(ctx, CreateSessionParams{
		ID:                sessionID,
		UserID:            userID,
		TokenHash:         domainauth.HashToken(token),
		CSRFSecret:        csrfSecret,
		AbsoluteExpiresAt: absoluteExpiresAt,
		IdleExpiresAt:     idleExpiresAt,
		CreatedIP:         strings.TrimSpace(ip),
		CreatedUserAgent:  strings.TrimSpace(userAgent),
	})
	if err != nil {
		return SessionView{}, err
	}

	return SessionView{
		ID:            record.ID,
		Token:         token,
		ExpiresAt:     record.AbsoluteExpiresAt,
		IdleExpiresAt: record.IdleExpiresAt,
	}, nil
}

func validateRegisterInput(input RegisterInput) error {
	if shared.Trim(input.Login) == "" {
		return shared.BlankField("auth.login")
	}
	if input.Password == "" {
		return shared.BlankField("auth.password")
	}
	if shared.Trim(input.FirstName) == "" {
		return shared.BlankField("auth.first_name")
	}
	if shared.Trim(input.GradebookNumber) == "" {
		return shared.BlankField("auth.gradebook_number")
	}
	if shared.Trim(input.GroupNumber) == "" {
		return shared.BlankField("auth.group_number")
	}
	if shared.Trim(input.Institute) == "" {
		return shared.BlankField("auth.institute")
	}
	if input.BirthDate.IsZero() {
		return shared.BlankField("auth.birth_date")
	}
	return nil
}

func loginLimiterKey(login, ip string) string {
	return login + "|" + strings.TrimSpace(ip)
}

func newRandomID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
