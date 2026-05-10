package auth_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appauth "activist-base/src/application/auth"
	"activist-base/src/config"
	domainauth "activist-base/src/domain/auth"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
	transporthttp "activist-base/src/transport/http"
)

func TestRegisterHandler(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	svc := &fakeAuthService{
		registerResult: appauth.AuthResult{
			User: domainuser.User{
				ID:              shared.UserID("user-1"),
				Login:           "activist_user",
				PasswordHash:    "hash-should-not-leak",
				FirstName:       "Anna",
				GradebookNumber: "GB-001",
				GroupNumber:     "A-11",
				Institute:       "Engineering",
			},
			Session: appauth.SessionView{
				ID:            "session-1",
				Token:         "opaque-token",
				ExpiresAt:     now.Add(24 * time.Hour),
				IdleExpiresAt: now.Add(2 * time.Hour),
			},
		},
	}

	repo := newFakeSessionRepository(now, 24*time.Hour)
	server := newAuthServer(t, svc, repo)

	body := `{"login":"activist_user","password":"StrongPassw0rd!","first_name":"Anna","gradebook_number":"GB-001","group_number":"A-11","institute":"Engineering","birth_date":"2004-03-14"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body=%s", rec.Code, rec.Body.String())
	}

	cookie := readSetCookie(t, rec.Result(), "__Host-session")
	if !cookie.Secure {
		t.Fatal("expected session cookie to be Secure")
	}
	if !cookie.HttpOnly {
		t.Fatal("expected session cookie to be HttpOnly")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", cookie.SameSite)
	}
	if cookie.Path != "/" {
		t.Fatalf("expected cookie path '/', got %q", cookie.Path)
	}
	if cookie.Domain != "" {
		t.Fatalf("expected empty cookie domain, got %q", cookie.Domain)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := payload["password_hash"]; ok {
		t.Fatal("response must not contain password_hash")
	}

	userObj := payload["user"].(map[string]any)
	if _, ok := userObj["password_hash"]; ok {
		t.Fatal("user payload must not contain password_hash")
	}
}

func TestLoginHandler(t *testing.T) {
	t.Parallel()

	svc := &fakeAuthService{
		loginFn: func(_ context.Context, input appauth.LoginInput) (appauth.AuthResult, error) {
			if input.Login == "missing_user" {
				return appauth.AuthResult{}, appauth.ErrInvalidCredentials
			}
			if input.Login == "known_user" && input.Password == "wrong" {
				return appauth.AuthResult{}, appauth.ErrInvalidCredentials
			}
			return appauth.AuthResult{}, appauth.ErrInvalidCredentials
		},
	}

	now := time.Now().UTC()
	repo := newFakeSessionRepository(now, 24*time.Hour)
	server := newAuthServer(t, svc, repo)

	first := doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"login":    "missing_user",
		"password": "StrongPassw0rd!",
	})
	second := doJSONRequest(t, server, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"login":    "known_user",
		"password": "wrong",
	})

	if first.Code != http.StatusUnauthorized || second.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for both invalid paths, got %d and %d", first.Code, second.Code)
	}

	firstError := readErrorEnvelope(t, first.Body.Bytes())
	secondError := readErrorEnvelope(t, second.Body.Bytes())
	if firstError != secondError {
		t.Fatalf("expected generic invalid credentials envelope, got %#v and %#v", firstError, secondError)
	}
}

func TestSessionHandler(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	svc := &fakeAuthService{}
	repo := newFakeSessionRepository(now, 24*time.Hour)
	server := newAuthServer(t, svc, repo)

	sessionReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	sessionReq.AddCookie(&http.Cookie{Name: "__Host-session", Value: repo.token})
	sessionRec := httptest.NewRecorder()
	server.ServeHTTP(sessionRec, sessionReq)

	if sessionRec.Code != http.StatusOK {
		t.Fatalf("expected 200 for session endpoint, got %d; body=%s", sessionRec.Code, sessionRec.Body.String())
	}

	var sessionBody map[string]any
	if err := json.Unmarshal(sessionRec.Body.Bytes(), &sessionBody); err != nil {
		t.Fatalf("decode session body: %v", err)
	}
	csrfTokenValue, ok := sessionBody["csrf_token"].(string)
	if !ok || csrfTokenValue == "" {
		t.Fatalf("expected non-empty csrf_token, got %#v", sessionBody["csrf_token"])
	}
	if _, ok := sessionBody["user"]; !ok {
		t.Fatal("expected session payload to contain user")
	}
	if _, ok := sessionBody["session"]; !ok {
		t.Fatal("expected session payload to contain session")
	}

	logoutNoCSRF := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutNoCSRF.AddCookie(&http.Cookie{Name: "__Host-session", Value: repo.token})
	logoutNoCSRFRec := httptest.NewRecorder()
	server.ServeHTTP(logoutNoCSRFRec, logoutNoCSRF)
	if logoutNoCSRFRec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without csrf token, got %d; body=%s", logoutNoCSRFRec.Code, logoutNoCSRFRec.Body.String())
	}

	logoutBadCSRF := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutBadCSRF.AddCookie(&http.Cookie{Name: "__Host-session", Value: repo.token})
	logoutBadCSRF.Header.Set("X-CSRF-Token", "mismatch")
	logoutBadCSRFRec := httptest.NewRecorder()
	server.ServeHTTP(logoutBadCSRFRec, logoutBadCSRF)
	if logoutBadCSRFRec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for csrf mismatch, got %d; body=%s", logoutBadCSRFRec.Code, logoutBadCSRFRec.Body.String())
	}

	logoutGoodCSRF := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutGoodCSRF.AddCookie(&http.Cookie{Name: "__Host-session", Value: repo.token})
	logoutGoodCSRF.Header.Set("X-CSRF-Token", csrfTokenValue)
	logoutGoodCSRFRec := httptest.NewRecorder()
	server.ServeHTTP(logoutGoodCSRFRec, logoutGoodCSRF)
	if logoutGoodCSRFRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for csrf match, got %d; body=%s", logoutGoodCSRFRec.Code, logoutGoodCSRFRec.Body.String())
	}
}

func TestLogoutHandler(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	svc := &fakeAuthService{}
	repo := newFakeSessionRepository(now, 24*time.Hour)
	server := newAuthServer(t, svc, repo)

	sessionResp := doSessionRequest(t, server, repo.token)
	csrfTokenValue := sessionResp["csrf_token"].(string)

	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.AddCookie(&http.Cookie{Name: "__Host-session", Value: repo.token})
	logoutReq.Header.Set("X-CSRF-Token", csrfTokenValue)
	logoutRec := httptest.NewRecorder()
	server.ServeHTTP(logoutRec, logoutReq)

	if logoutRec.Code != http.StatusNoContent {
		t.Fatalf("expected logout 204, got %d; body=%s", logoutRec.Code, logoutRec.Body.String())
	}
	if svc.logoutToken != repo.token {
		t.Fatalf("expected logout token %q, got %q", repo.token, svc.logoutToken)
	}

	clearCookie := readSetCookie(t, logoutRec.Result(), "__Host-session")
	if clearCookie.MaxAge != -1 {
		t.Fatalf("expected clearing cookie max age -1, got %d", clearCookie.MaxAge)
	}
	if clearCookie.Path != "/" || clearCookie.Domain != "" || !clearCookie.HttpOnly || !clearCookie.Secure {
		t.Fatalf("clear cookie contract mismatch: %#v", clearCookie)
	}

	repo = newFakeSessionRepository(now, 24*time.Hour)
	server = newAuthServer(t, svc, repo)
	sessionResp = doSessionRequest(t, server, repo.token)
	csrfTokenValue = sessionResp["csrf_token"].(string)

	logoutAllReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout-all", nil)
	logoutAllReq.AddCookie(&http.Cookie{Name: "__Host-session", Value: repo.token})
	logoutAllReq.Header.Set("X-CSRF-Token", csrfTokenValue)
	logoutAllRec := httptest.NewRecorder()
	server.ServeHTTP(logoutAllRec, logoutAllReq)

	if logoutAllRec.Code != http.StatusNoContent {
		t.Fatalf("expected logout-all 204, got %d; body=%s", logoutAllRec.Code, logoutAllRec.Body.String())
	}
	if svc.logoutAllUserID != shared.UserID("user-1") {
		t.Fatalf("expected logout-all for user-1, got %q", svc.logoutAllUserID)
	}
}

type fakeAuthService struct {
	registerResult appauth.AuthResult
	loginResult    appauth.AuthResult

	registerErr error
	loginErr    error
	logoutErr   error

	loginFn func(context.Context, appauth.LoginInput) (appauth.AuthResult, error)

	logoutToken     string
	logoutAllUserID shared.UserID
}

func (f *fakeAuthService) Register(_ context.Context, _ appauth.RegisterInput) (appauth.AuthResult, error) {
	return f.registerResult, f.registerErr
}

func (f *fakeAuthService) Login(ctx context.Context, input appauth.LoginInput) (appauth.AuthResult, error) {
	if f.loginFn != nil {
		return f.loginFn(ctx, input)
	}
	return f.loginResult, f.loginErr
}

func (f *fakeAuthService) Logout(_ context.Context, input appauth.LogoutInput) error {
	f.logoutToken = input.Token
	return f.logoutErr
}

func (f *fakeAuthService) LogoutAll(_ context.Context, input appauth.LogoutAllInput) error {
	f.logoutAllUserID = input.UserID
	return nil
}

type fakeSessionRepository struct {
	token   string
	session appauth.SessionRecord
	user    domainuser.User
}

func newFakeSessionRepository(now time.Time, idleTTL time.Duration) *fakeSessionRepository {
	token := "opaque-cookie-token"
	absolute := now.Add(24 * time.Hour)
	return &fakeSessionRepository{
		token: token,
		session: appauth.SessionRecord{
			ID:                "session-1",
			UserID:            shared.UserID("user-1"),
			TokenHash:         domainauth.HashToken(token),
			CSRFSecret:        "csrf-secret",
			CreatedAt:         now,
			LastSeenAt:        now,
			AbsoluteExpiresAt: absolute,
			IdleExpiresAt:     now.Add(idleTTL),
		},
		user: domainuser.User{
			ID:              shared.UserID("user-1"),
			Login:           "activist_user",
			PasswordHash:    "hash-should-not-leak",
			FirstName:       "Anna",
			GradebookNumber: "GB-001",
			GroupNumber:     "A-11",
			Institute:       "Engineering",
		},
	}
}

func (f *fakeSessionRepository) GetSessionByTokenHash(_ context.Context, tokenHash string) (appauth.SessionRecord, error) {
	if tokenHash != f.session.TokenHash {
		return appauth.SessionRecord{}, sql.ErrNoRows
	}
	return f.session, nil
}

func (f *fakeSessionRepository) TouchSession(_ context.Context, params appauth.TouchSessionParams) (appauth.SessionRecord, error) {
	if params.ID != f.session.ID {
		return appauth.SessionRecord{}, sql.ErrNoRows
	}
	f.session.LastSeenAt = params.LastSeenAt
	f.session.IdleExpiresAt = params.IdleExpiresAt
	return f.session, nil
}

func (f *fakeSessionRepository) RevokeSession(_ context.Context, sessionID string, revokedAt time.Time, reason string) (bool, error) {
	if sessionID != f.session.ID {
		return false, nil
	}
	f.session.RevokedAt = &revokedAt
	f.session.RevokedReason = reason
	return true, nil
}

func (f *fakeSessionRepository) GetUserByID(_ context.Context, id shared.UserID) (domainuser.User, error) {
	if id != f.user.ID {
		return domainuser.User{}, sql.ErrNoRows
	}
	return f.user, nil
}

func newAuthServer(t *testing.T, svc *fakeAuthService, repo *fakeSessionRepository) http.Handler {
	t.Helper()

	return transporthttp.NewRouter(transporthttp.Deps{
		Config: config.Config{
			SessionCookieName:  "__Host-session",
			SessionIdleTTL:     2 * time.Hour,
			SessionAbsoluteTTL: 24 * time.Hour,
		},
		AuthService:       svc,
		SessionRepository: repo,
	})
}

func doJSONRequest(t *testing.T, server http.Handler, method, path string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	return rec
}

func doSessionRequest(t *testing.T, server http.Handler, token string) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/session", nil)
	req.AddCookie(&http.Cookie{Name: "__Host-session", Value: token})
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("session status: %d body=%s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode session response: %v", err)
	}
	return body
}

func readSetCookie(t *testing.T, response *http.Response, name string) *http.Cookie {
	t.Helper()
	for _, cookie := range response.Cookies() {
		if cookie.Name == name {
			return cookie
		}
	}
	t.Fatalf("cookie %q not found", name)
	return nil
}

type errorShape struct {
	Code    string
	Message string
}

func readErrorEnvelope(t *testing.T, body []byte) errorShape {
	t.Helper()

	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode error envelope: %v", err)
	}
	return errorShape{
		Code:    envelope.Error.Code,
		Message: envelope.Error.Message,
	}
}
