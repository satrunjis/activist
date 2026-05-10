package session

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	appauth "activist-base/src/application/auth"
	domainauth "activist-base/src/domain/auth"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
)

const (
	CSRFHeader               = "X-CSRF-Token"
	defaultSessionCookieName = "__Host-session"
	defaultSessionIdleTTL    = 8 * time.Hour
)

type Repository interface {
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (appauth.SessionRecord, error)
	TouchSession(ctx context.Context, params appauth.TouchSessionParams) (appauth.SessionRecord, error)
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time, reason string) (bool, error)
	GetUserByID(ctx context.Context, id shared.UserID) (domainuser.User, error)
}

type Options struct {
	CookieName     string
	SessionIdleTTL time.Duration
	DevAssumeAdmin bool
	DevAdminLogin  string
}

type Middleware struct {
	repo           Repository
	cookieName     string
	sessionIdleTTL time.Duration
	devAssumeAdmin bool
	devAdminLogin  string
	nowFn          func() time.Time
}

func NewMiddleware(repo Repository, options Options) *Middleware {
	cookieName := options.CookieName
	if cookieName == "" {
		cookieName = defaultSessionCookieName
	}
	idleTTL := options.SessionIdleTTL
	if idleTTL <= 0 {
		idleTTL = defaultSessionIdleTTL
	}
	return &Middleware{
		repo:           repo,
		cookieName:     cookieName,
		sessionIdleTTL: idleTTL,
		devAssumeAdmin: options.DevAssumeAdmin,
		devAdminLogin:  strings.TrimSpace(options.DevAdminLogin),
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	if m == nil || m.repo == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeError(w, appauth.ErrInvalidSession)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := m.readSessionToken(r)
		if !ok {
			writeError(w, appauth.ErrInvalidSession)
			return
		}

		record, err := m.repo.GetSessionByTokenHash(r.Context(), domainauth.HashToken(token))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, appauth.ErrInvalidSession)
				return
			}
			writeError(w, err)
			return
		}

		now := m.nowFn().UTC()
		if record.RevokedAt != nil || domainauth.SessionExpired(now, record.IdleExpiresAt, record.AbsoluteExpiresAt) {
			_, _ = m.repo.RevokeSession(r.Context(), record.ID, now, "expired")
			writeError(w, appauth.ErrInvalidSession)
			return
		}

		record, err = m.repo.TouchSession(r.Context(), appauth.TouchSessionParams{
			ID:            record.ID,
			LastSeenAt:    now,
			IdleExpiresAt: domainauth.NextIdleExpiry(now, m.sessionIdleTTL, record.AbsoluteExpiresAt),
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, appauth.ErrInvalidSession)
				return
			}
			writeError(w, err)
			return
		}

		persistedUser, err := m.repo.GetUserByID(r.Context(), record.UserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, appauth.ErrInvalidSession)
				return
			}
			writeError(w, err)
			return
		}

		csrfToken := csrfToken(record.CSRFSecret, token)
		if requiresCSRFCheck(r.Method) {
			headerToken := strings.TrimSpace(r.Header.Get(CSRFHeader))
			if headerToken == "" {
				writeError(w, &shared.Error{
					Code:    "auth.csrf_required",
					Message: "csrf token is required",
				})
				return
			}
			if subtle.ConstantTimeCompare([]byte(headerToken), []byte(csrfToken)) != 1 {
				writeError(w, &shared.Error{
					Code:    "access.csrf_mismatch",
					Message: "csrf token mismatch",
				})
				return
			}
		}

		actor := Actor{
			User:                 persistedUser,
			UserID:               persistedUser.ID,
			SessionID:            record.ID,
			SessionToken:         token,
			SessionExpiresAt:     record.AbsoluteExpiresAt,
			SessionIdleExpiresAt: record.IdleExpiresAt,
			CSRFToken:            csrfToken,
		}

		next.ServeHTTP(w, r.WithContext(WithActor(r.Context(), actor)))
	})
}

func (m *Middleware) readSessionToken(r *http.Request) (string, bool) {
	if r == nil {
		return "", false
	}
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return "", false
	}
	token := strings.TrimSpace(cookie.Value)
	return token, token != ""
}

func requiresCSRFCheck(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func csrfToken(secret, token string) string {
	sum := sha256.Sum256([]byte(secret + ":" + token))
	return hex.EncodeToString(sum[:])
}

type errorEnvelope struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	payload := errorPayload{
		Code:    "internal.error",
		Message: "internal server error",
	}

	var domainErr *shared.Error
	if errors.As(err, &domainErr) {
		statusCode = statusFromDomainCode(domainErr.Code)
		payload.Code = string(domainErr.Code)
		payload.Message = domainErr.Message
		if statusCode == http.StatusUnauthorized {
			payload.Message = "unauthorized"
		}
		if statusCode == http.StatusForbidden {
			payload.Message = "operation is forbidden"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: payload})
}

func statusFromDomainCode(code shared.ErrorCode) int {
	switch {
	case strings.HasPrefix(string(code), "access."):
		return http.StatusForbidden
	case strings.HasPrefix(string(code), "auth."):
		return http.StatusUnauthorized
	case strings.HasPrefix(string(code), "not_found."):
		return http.StatusNotFound
	case strings.HasPrefix(string(code), "validation."):
		return http.StatusBadRequest
	case code == "shared.empty_id":
		return http.StatusBadRequest
	default:
		return http.StatusUnprocessableEntity
	}
}
