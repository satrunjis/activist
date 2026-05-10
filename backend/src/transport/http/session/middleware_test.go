package session

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	appauth "activist-base/src/application/auth"
	domainauth "activist-base/src/domain/auth"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
)

func TestActorContract_DoesNotExposeActorAccessOrResolverState(t *testing.T) {
	t.Parallel()

	actorType := reflect.TypeOf(Actor{})
	if _, ok := actorType.FieldByName("Access"); ok {
		t.Fatal("actor.Access must not be injected by session middleware")
	}
}

func TestAuthenticate_AddsActorContextWithoutAuthorizationResolver(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 4, 13, 10, 0, 0, 0, time.UTC)
	token := "session-token"
	record := appauth.SessionRecord{
		ID:                "sess-1",
		UserID:            shared.UserID("u-1"),
		TokenHash:         domainauth.HashToken(token),
		CSRFSecret:        "csrf-secret",
		CreatedAt:         now.Add(-5 * time.Minute),
		LastSeenAt:        now.Add(-5 * time.Minute),
		AbsoluteExpiresAt: now.Add(2 * time.Hour),
		IdleExpiresAt:     now.Add(30 * time.Minute),
	}

	repo := &fakeMiddlewareRepository{
		getSessionByTokenHashFunc: func(_ context.Context, tokenHash string) (appauth.SessionRecord, error) {
			if tokenHash != domainauth.HashToken(token) {
				t.Fatalf("unexpected token hash: %q", tokenHash)
			}
			return record, nil
		},
		touchSessionFunc: func(_ context.Context, params appauth.TouchSessionParams) (appauth.SessionRecord, error) {
			record.LastSeenAt = params.LastSeenAt
			record.IdleExpiresAt = params.IdleExpiresAt
			return record, nil
		},
		getUserByIDFunc: func(_ context.Context, id shared.UserID) (domainuser.User, error) {
			return domainuser.User{
				ID:           id,
				Login:        "tester",
				PasswordHash: "hash",
				FirstName:    "Test",
			}, nil
		},
	}

	middleware := NewMiddleware(repo, Options{})
	middleware.nowFn = func() time.Time { return now }

	var captured Actor
	handler := middleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor, ok := ActorFromRequest(r)
		if !ok {
			t.Fatal("expected actor in request context")
		}
		captured = actor
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/divisions/tree", nil)
	req.AddCookie(&http.Cookie{Name: defaultSessionCookieName, Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
	if captured.UserID != shared.UserID("u-1") {
		t.Fatalf("expected actor user id u-1, got %q", captured.UserID)
	}
	if captured.SessionID != "sess-1" {
		t.Fatalf("expected session id sess-1, got %q", captured.SessionID)
	}
	if captured.SessionToken != token {
		t.Fatalf("expected session token %q, got %q", token, captured.SessionToken)
	}
}

type fakeMiddlewareRepository struct {
	getSessionByTokenHashFunc func(ctx context.Context, tokenHash string) (appauth.SessionRecord, error)
	touchSessionFunc          func(ctx context.Context, params appauth.TouchSessionParams) (appauth.SessionRecord, error)
	revokeSessionFunc         func(ctx context.Context, sessionID string, revokedAt time.Time, reason string) (bool, error)
	getUserByIDFunc           func(ctx context.Context, id shared.UserID) (domainuser.User, error)
}

func (f *fakeMiddlewareRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (appauth.SessionRecord, error) {
	if f.getSessionByTokenHashFunc == nil {
		return appauth.SessionRecord{}, nil
	}
	return f.getSessionByTokenHashFunc(ctx, tokenHash)
}

func (f *fakeMiddlewareRepository) TouchSession(ctx context.Context, params appauth.TouchSessionParams) (appauth.SessionRecord, error) {
	if f.touchSessionFunc == nil {
		return appauth.SessionRecord{}, nil
	}
	return f.touchSessionFunc(ctx, params)
}

func (f *fakeMiddlewareRepository) RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time, reason string) (bool, error) {
	if f.revokeSessionFunc == nil {
		return true, nil
	}
	return f.revokeSessionFunc(ctx, sessionID, revokedAt, reason)
}

func (f *fakeMiddlewareRepository) GetUserByID(ctx context.Context, id shared.UserID) (domainuser.User, error) {
	if f.getUserByIDFunc == nil {
		return domainuser.User{}, nil
	}
	return f.getUserByIDFunc(ctx, id)
}
