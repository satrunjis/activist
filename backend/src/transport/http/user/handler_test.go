package user_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appauth "activist-base/src/application/auth"
	appuser "activist-base/src/application/user"
	"activist-base/src/config"
	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
	transporthttp "activist-base/src/transport/http"
	httpsession "activist-base/src/transport/http/session"
	httpuser "activist-base/src/transport/http/user"
)

func TestUserHandler_GetProfileVisibility(t *testing.T) {
	t.Parallel()

	birthDate := time.Date(2004, 3, 14, 0, 0, 0, 0, time.UTC)
	targetUser := domainuser.User{
		ID:              shared.UserID("target-user"),
		Login:           "target_login",
		PasswordHash:    "secret-hash",
		FirstName:       "Ivan",
		LastName:        "Ivanov",
		MiddleName:      "Ivanovich",
		GradebookNumber: "GB-123",
		GroupNumber:     "A-11",
		Institute:       "Engineering",
		BirthDate:       &birthDate,
		Phone:           "+79990001122",
		SocialLinks: []shared.Link{
			{Platform: "tg", Value: "t.me/target"},
		},
		About: "Active member",
	}

	t.Run("hides restricted fields without can_view_contacts", func(t *testing.T) {
		t.Parallel()

		repo := &fakeUserRepo{users: map[shared.UserID]domainuser.User{
			targetUser.ID: targetUser,
		}}
		server := newUserServer(
			t,
			repo,
			&fakeMembershipReader{},
			withActor(httpsession.Actor{
				UserID: shared.UserID("viewer-1"),
				User: domainuser.User{
					ID:           shared.UserID("viewer-1"),
					Login:        "viewer",
					PasswordHash: "hash",
					FirstName:    "Viewer",
				},
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/target-user", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if _, ok := payload["phone"]; ok {
			t.Fatal("phone must be hidden without can_view_contacts")
		}
		if _, ok := payload["social_links"]; ok {
			t.Fatal("social_links must be hidden without can_view_contacts")
		}
		if _, ok := payload["birth_date"]; ok {
			t.Fatal("birth_date must be hidden for unrelated actor")
		}
		if _, ok := payload["password_hash"]; ok {
			t.Fatal("password_hash must never be serialized")
		}
	})

	t.Run("shows contact fields with can_view_contacts but hides birth_date", func(t *testing.T) {
		t.Parallel()

		repo := &fakeUserRepo{users: map[shared.UserID]domainuser.User{
			targetUser.ID: targetUser,
		}}
		server := newUserServer(
			t,
			repo,
			&fakeMembershipReader{},
			withActor(httpsession.Actor{
				UserID: shared.UserID("viewer-2"),
				User: domainuser.User{
					ID:           shared.UserID("viewer-2"),
					Login:        "viewer2",
					PasswordHash: "hash",
					FirstName:    "Viewer",
				},
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/target-user", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if _, ok := payload["phone"]; !ok {
			t.Fatal("phone must be visible with can_view_contacts")
		}
		if _, ok := payload["social_links"]; !ok {
			t.Fatal("social_links must be visible with can_view_contacts")
		}
		if _, ok := payload["birth_date"]; ok {
			t.Fatal("birth_date must stay hidden without self/system_admin")
		}
	})
}

func TestUserHandler_PatchProfile(t *testing.T) {
	t.Parallel()

	targetUser := domainuser.User{
		ID:           shared.UserID("target-user"),
		Login:        "target_login",
		PasswordHash: "secret-hash",
		FirstName:    "Ivan",
	}

	t.Run("returns 403 for foreign target", func(t *testing.T) {
		t.Parallel()

		repo := &fakeUserRepo{users: map[shared.UserID]domainuser.User{
			targetUser.ID: targetUser,
		}}
		server := newUserServer(
			t,
			repo,
			&fakeMembershipReader{},
			withActor(httpsession.Actor{
				UserID: shared.UserID("viewer-1"),
				User: domainuser.User{
					ID:           shared.UserID("viewer-1"),
					Login:        "viewer",
					PasswordHash: "hash",
					FirstName:    "Viewer",
				},
			}),
		)

		body := []byte(`{"first_name":"Updated"}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/target-user", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for foreign profile patch, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("rejects unknown fields in patch payload", func(t *testing.T) {
		t.Parallel()

		selfUser := domainuser.User{
			ID:           shared.UserID("self-user"),
			Login:        "self_login",
			PasswordHash: "secret-hash",
			FirstName:    "Self",
		}
		repo := &fakeUserRepo{users: map[shared.UserID]domainuser.User{
			selfUser.ID: selfUser,
		}}
		server := newUserServer(
			t,
			repo,
			&fakeMembershipReader{},
			withActor(httpsession.Actor{
				UserID: selfUser.ID,
				User:   selfUser,
			}),
		)

		body := []byte(`{"login":"hijack"}`)
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/self-user", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for unknown patch field, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestUserHandler_Memberships(t *testing.T) {
	t.Parallel()

	targetUser := domainuser.User{
		ID:           shared.UserID("target-user"),
		Login:        "target_login",
		PasswordHash: "hash",
		FirstName:    "Ivan",
	}

	t.Run("requires authentication", func(t *testing.T) {
		t.Parallel()

		repo := &fakeUserRepo{users: map[shared.UserID]domainuser.User{
			targetUser.ID: targetUser,
		}}
		server := newUserServer(t, repo, &fakeMembershipReader{})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/target-user/memberships", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 without session actor, got %d", rec.Code)
		}
	})

	t.Run("returns stable empty contract when no active memberships exist", func(t *testing.T) {
		t.Parallel()

		repo := &fakeUserRepo{users: map[shared.UserID]domainuser.User{
			targetUser.ID: targetUser,
		}}
		server := newUserServer(
			t,
			repo,
			&fakeMembershipReader{},
			withActor(httpsession.Actor{
				UserID: shared.UserID("viewer-1"),
				User: domainuser.User{
					ID:           shared.UserID("viewer-1"),
					Login:        "viewer",
					PasswordHash: "hash",
					FirstName:    "Viewer",
				},
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/target-user/memberships", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload struct {
			Items []any `json:"items"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode memberships response: %v", err)
		}
		if payload.Items == nil {
			t.Fatal("items must be a JSON array")
		}
		if len(payload.Items) != 0 {
			t.Fatalf("expected empty items array, got %d", len(payload.Items))
		}
	})

	t.Run("returns forbidden when resolved access denies memberships read", func(t *testing.T) {
		t.Parallel()

		repo := &fakeUserRepo{users: map[shared.UserID]domainuser.User{
			targetUser.ID: targetUser,
		}}
		server := newUserServer(
			t,
			repo,
			&fakeMembershipReader{},
			withActor(httpsession.Actor{
				UserID: shared.UserID("viewer-3"),
				User: domainuser.User{
					ID:           shared.UserID("viewer-3"),
					Login:        "viewer3",
					PasswordHash: "hash",
					FirstName:    "Viewer",
				},
			}),
		)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/target-user/memberships", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403 when resolver returns forbidden, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestUserRoutes_AreMounted(t *testing.T) {
	t.Parallel()

	repo := &fakeUserRepo{
		users: map[shared.UserID]domainuser.User{
			shared.UserID("target-user"): {
				ID:           shared.UserID("target-user"),
				Login:        "target_login",
				PasswordHash: "hash",
				FirstName:    "Target",
			},
		},
	}
	server := transporthttp.NewRouter(transporthttp.Deps{
		Config: config.Config{
			SessionCookieName:  "__Host-session",
			SessionIdleTTL:     2 * time.Hour,
			SessionAbsoluteTTL: 24 * time.Hour,
		},
		AuthService:       &fakeAuthService{},
		SessionRepository: &fakeSessionRepository{},
		UserProfileStore:  repo,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/target-user", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatalf("expected /api/v1/users/{user_id} route to be mounted, got 404")
	}
}

type fakeUserRepo struct {
	users map[shared.UserID]domainuser.User
}

func TestUserHandler_UsesResolverInConstructor(t *testing.T) {
	t.Parallel()

	repo := &fakeUserRepo{users: map[shared.UserID]domainuser.User{}}
	memberships := &fakeMembershipReader{}
	authz := &fakeUserAuthorizationService{}

	_ = httpuser.NewHandler(repo, memberships, authz)
}

type fakeMembershipReader struct {
	items []appuser.MembershipView
	err   error
}

type fakeUserAuthorizationService struct{}

func (f *fakeUserAuthorizationService) ResolveForDivision(context.Context, shared.UserID, shared.DivisionID) (domainmembership.EffectivePermissions, error) {
	return domainmembership.EffectivePermissions{}, nil
}

func (f *fakeUserAuthorizationService) ResolveForUser(_ context.Context, actorID shared.UserID, targetUserID shared.UserID) (domainmembership.EffectivePermissions, error) {
	if actorID == shared.UserID("viewer-3") && targetUserID == shared.UserID("target-user") {
		return domainmembership.EffectivePermissions{}, shared.ErrForbidden
	}
	if actorID == targetUserID {
		return permissionAccess(role.CanEditSelfProfile, role.ScopeSelf, shared.DivisionID(actorID), shared.DivisionID(targetUserID)), nil
	}
	if actorID == shared.UserID("viewer-2") && targetUserID == shared.UserID("target-user") {
		return permissionAccess(role.CanViewContacts, role.ScopeCurrentDivision, "div-1", "div-1"), nil
	}
	return domainmembership.EffectivePermissions{}, nil
}

func (f *fakeUserAuthorizationService) ResolveGlobal(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
	return domainmembership.EffectivePermissions{}, nil
}

func (f *fakeMembershipReader) ListByUser(_ context.Context, _ shared.UserID) ([]appuser.MembershipView, error) {
	if f.err != nil {
		return nil, f.err
	}
	return append([]appuser.MembershipView(nil), f.items...), nil
}

func (f *fakeUserRepo) GetUserByID(_ context.Context, id shared.UserID) (domainuser.User, error) {
	user, ok := f.users[id]
	if !ok {
		return domainuser.User{}, sql.ErrNoRows
	}
	return user, nil
}

func (f *fakeUserRepo) UpdateProfile(_ context.Context, user domainuser.User) (domainuser.User, error) {
	if _, ok := f.users[user.ID]; !ok {
		return domainuser.User{}, sql.ErrNoRows
	}
	f.users[user.ID] = user
	return user, nil
}

type fakeAuthService struct{}

func (f *fakeAuthService) Register(_ context.Context, _ appauth.RegisterInput) (appauth.AuthResult, error) {
	return appauth.AuthResult{}, nil
}
func (f *fakeAuthService) Login(_ context.Context, _ appauth.LoginInput) (appauth.AuthResult, error) {
	return appauth.AuthResult{}, nil
}
func (f *fakeAuthService) Logout(_ context.Context, _ appauth.LogoutInput) error {
	return nil
}
func (f *fakeAuthService) LogoutAll(_ context.Context, _ appauth.LogoutAllInput) error {
	return nil
}

type fakeSessionRepository struct{}

func (f *fakeSessionRepository) GetSessionByTokenHash(_ context.Context, _ string) (appauth.SessionRecord, error) {
	return appauth.SessionRecord{}, sql.ErrNoRows
}
func (f *fakeSessionRepository) TouchSession(_ context.Context, _ appauth.TouchSessionParams) (appauth.SessionRecord, error) {
	return appauth.SessionRecord{}, sql.ErrNoRows
}
func (f *fakeSessionRepository) RevokeSession(_ context.Context, _ string, _ time.Time, _ string) (bool, error) {
	return false, nil
}
func (f *fakeSessionRepository) GetUserByID(_ context.Context, _ shared.UserID) (domainuser.User, error) {
	return domainuser.User{}, sql.ErrNoRows
}

func newUserServer(t *testing.T, repo httpuser.ProfileRepository, memberships httpuser.MembershipReader, middleware ...func(http.Handler) http.Handler) http.Handler {
	t.Helper()

	handler := httpuser.NewHandler(repo, memberships, &fakeUserAuthorizationService{})
	router := http.NewServeMux()
	chiRouter := httpuser.NewTestRouter(handler)

	var wrapped = chiRouter
	for i := len(middleware) - 1; i >= 0; i-- {
		wrapped = middleware[i](wrapped)
	}
	router.Handle("/", wrapped)
	return router
}

func withActor(actor httpsession.Actor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpsession.WithActor(r.Context(), actor)))
		})
	}
}

func permissionAccess(
	code role.PermissionCode,
	scope role.ScopeMode,
	sourceDivision shared.DivisionID,
	targetDivision shared.DivisionID,
) domainmembership.EffectivePermissions {
	permission, err := role.NewPermission(code, scope)
	if err != nil {
		return domainmembership.EffectivePermissions{}
	}
	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  sourceDivision,
				Permissions: role.NewPermissionSet([]role.Permission{permission}),
			},
		},
		targetDivision,
		nil,
	)
}
