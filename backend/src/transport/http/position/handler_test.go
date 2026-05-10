package position_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	appeventlog "activist-base/src/application/eventlog"
	appuser "activist-base/src/application/user"
	domaindivision "activist-base/src/domain/division"
	domaineventlog "activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	httpposition "activist-base/src/transport/http/position"
	httpsession "activist-base/src/transport/http/session"
)

func TestCreatePosition_Unauthorized(t *testing.T) {
	t.Parallel()

	server := newPositionServer(t, &fakePositionRepo{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions/div-1/positions", bytes.NewReader([]byte(`{"title":"Chair","role_id":"role-1"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreatePosition_OK(t *testing.T) {
	t.Parallel()

	repo := &fakePositionRepo{
		createFunc: func(_ context.Context, _ appeventlog.Transaction, pos domainposition.Position) (domainposition.Position, error) {
			return pos, nil
		},
	}
	server := newPositionServer(t, repo, withActor(httpsession.Actor{
		UserID: shared.UserID("admin-1"),
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions/div-1/positions", bytes.NewReader([]byte(`{"title":"Chair","role_id":"role-1"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListPositionsByDivision_OK(t *testing.T) {
	t.Parallel()

	roleID := shared.RoleID("role-1")
	divisionID := shared.DivisionID("div-1")
	repo := &fakePositionRepo{
		listByDivisionFunc: func(context.Context, shared.DivisionID) ([]domainposition.Position, error) {
			return []domainposition.Position{
				{ID: shared.PositionID("p1"), Title: "Chair", RoleID: roleID, DivisionID: divisionID},
				{ID: shared.PositionID("p2"), Title: "Deputy", RoleID: roleID, DivisionID: divisionID},
			}, nil
		},
	}
	server := newPositionServer(t, repo, withActor(httpsession.Actor{UserID: shared.UserID("viewer-1")}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/divisions/div-1/positions", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	items, ok := payload["items"].([]any)
	if !ok {
		t.Fatalf("expected items array, got %#v", payload["items"])
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestListPositionMembers_Unauthorized(t *testing.T) {
	t.Parallel()

	server := newPositionServer(t, &fakePositionRepo{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/positions/p1/members", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListPositionMembers_Forbidden(t *testing.T) {
	t.Parallel()

	repoCalls := 0
	repo := &fakePositionRepo{
		getByIDFunc: func(_ context.Context, id shared.PositionID) (domainposition.Position, error) {
			return domainposition.Position{
				ID:         id,
				Title:      "Chair",
				RoleID:     shared.RoleID("role-1"),
				DivisionID: shared.DivisionID("child-1"),
			}, nil
		},
		listMembersByPositionFunc: func(context.Context, shared.PositionID) ([]appuser.MembershipView, error) {
			repoCalls++
			return nil, nil
		},
	}
	authz := &fakePositionAuthorizationService{
		ancestorsByTarget: map[shared.DivisionID][]shared.DivisionID{
			shared.DivisionID("child-1"): {shared.DivisionID("parent-1")},
		},
	}
	server := newPositionServerWithAuthz(t, repo, authz, withActor(httpsession.Actor{UserID: shared.UserID("viewer-1")}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/positions/p1/members", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
	if authz.lastResolveTarget != shared.DivisionID("child-1") {
		t.Fatalf("expected resolver target child-1, got %q", authz.lastResolveTarget)
	}
	if repoCalls != 0 {
		t.Fatalf("expected forbidden path to skip member read, got %d calls", repoCalls)
	}
}

func TestListPositionMembers_OK(t *testing.T) {
	t.Parallel()

	repo := &fakePositionRepo{
		getByIDFunc: func(_ context.Context, id shared.PositionID) (domainposition.Position, error) {
			return domainposition.Position{
				ID:         id,
				Title:      "Chair",
				RoleID:     shared.RoleID("role-1"),
				DivisionID: shared.DivisionID("div-1"),
			}, nil
		},
		listMembersByPositionFunc: func(_ context.Context, id shared.PositionID) ([]appuser.MembershipView, error) {
			return []appuser.MembershipView{
				{
					Membership: domainmembership.Membership{
						UserID:     shared.UserID("u1"),
						PositionID: id,
					},
					Position: domainposition.Position{
						ID:         id,
						Title:      "Chair",
						RoleID:     shared.RoleID("role-1"),
						DivisionID: shared.DivisionID("div-1"),
					},
					Division: domaindivision.Division{
						ID:        shared.DivisionID("div-1"),
						ShortName: "Division One",
					},
					RoleName: "Coordinator",
				},
			}, nil
		},
	}
	authz := &fakePositionAuthorizationService{
		resolveForDivisionFunc: func(_ context.Context, actorID shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
			if actorID != shared.UserID("admin-1") {
				t.Fatalf("unexpected actor id: %s", actorID)
			}
			if targetDivisionID != shared.DivisionID("div-1") {
				t.Fatalf("unexpected division id: %s", targetDivisionID)
			}
			return permissionAccess(targetDivisionID, domainrole.CanManagePositions), nil
		},
	}
	server := newPositionServerWithAuthz(t, repo, authz, withActor(httpsession.Actor{UserID: shared.UserID("admin-1")}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/positions/p1/members", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Items []httpposition.PositionMemberItem `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(payload.Items))
	}
	if payload.Items[0].UserID != "u1" {
		t.Fatalf("expected first item user_id=u1, got %q", payload.Items[0].UserID)
	}
}

type fakePositionRepo struct {
	createFunc                func(context.Context, appeventlog.Transaction, domainposition.Position) (domainposition.Position, error)
	getDivisionByID           func(context.Context, shared.DivisionID) (domaindivision.Division, error)
	getRoleByID               func(context.Context, shared.RoleID) (domainrole.Role, error)
	getByIDFunc               func(context.Context, shared.PositionID) (domainposition.Position, error)
	listByDivisionFunc        func(context.Context, shared.DivisionID) ([]domainposition.Position, error)
	listMembersByPositionFunc func(context.Context, shared.PositionID) ([]appuser.MembershipView, error)
	archiveFunc               func(context.Context, appeventlog.Transaction, shared.PositionID) (domainposition.Position, int, error)
}

func (f *fakePositionRepo) CreateWithTx(ctx context.Context, tx appeventlog.Transaction, p domainposition.Position) (domainposition.Position, error) {
	if f.createFunc != nil {
		return f.createFunc(ctx, tx, p)
	}
	return p, nil
}

func (f *fakePositionRepo) GetDivisionByID(ctx context.Context, id shared.DivisionID) (domaindivision.Division, error) {
	if f.getDivisionByID != nil {
		return f.getDivisionByID(ctx, id)
	}
	return domaindivision.Division{ID: id, ShortName: "Division"}, nil
}

func (f *fakePositionRepo) GetRoleByID(ctx context.Context, id shared.RoleID) (domainrole.Role, error) {
	if f.getRoleByID != nil {
		return f.getRoleByID(ctx, id)
	}
	perm, err := domainrole.NewPermission(domainrole.CanManagePositions, domainrole.ScopeCurrentAndDescendants)
	if err != nil {
		return domainrole.Role{}, err
	}
	return domainrole.Role{
		ID:          id,
		Name:        "Role",
		Permissions: domainrole.NewPermissionSet([]domainrole.Permission{perm}),
	}, nil
}

func (f *fakePositionRepo) GetByID(ctx context.Context, id shared.PositionID) (domainposition.Position, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return domainposition.Position{ID: id, Title: "Chair", RoleID: shared.RoleID("role-1"), DivisionID: shared.DivisionID("div-1")}, nil
}

func (f *fakePositionRepo) ListByDivision(ctx context.Context, id shared.DivisionID) ([]domainposition.Position, error) {
	if f.listByDivisionFunc != nil {
		return f.listByDivisionFunc(ctx, id)
	}
	return []domainposition.Position{}, nil
}

func (f *fakePositionRepo) ListMembersByPosition(ctx context.Context, id shared.PositionID) ([]appuser.MembershipView, error) {
	if f.listMembersByPositionFunc != nil {
		return f.listMembersByPositionFunc(ctx, id)
	}
	return []appuser.MembershipView{}, nil
}

func (f *fakePositionRepo) ArchiveWithTx(ctx context.Context, tx appeventlog.Transaction, id shared.PositionID) (domainposition.Position, int, error) {
	if f.archiveFunc != nil {
		return f.archiveFunc(ctx, tx, id)
	}
	return domainposition.Position{ID: id, Title: "Chair", RoleID: shared.RoleID("role-1"), DivisionID: shared.DivisionID("div-1"), IsArchived: true}, 0, nil
}

func newPositionServer(t *testing.T, repo httpposition.Repository, middleware ...func(http.Handler) http.Handler) http.Handler {
	return newPositionServerWithAuthz(t, repo, &fakePositionAuthorizationService{}, middleware...)
}

func newPositionServerWithAuthz(t *testing.T, repo httpposition.Repository, authz *fakePositionAuthorizationService, middleware ...func(http.Handler) http.Handler) http.Handler {
	t.Helper()

	handler := httpposition.NewHandler(repo, authz, newPositionCommandService())
	router := http.NewServeMux()
	chiRouter := httpposition.NewTestRouter(handler)

	var wrapped = chiRouter
	for i := len(middleware) - 1; i >= 0; i-- {
		wrapped = middleware[i](wrapped)
	}
	router.Handle("/", wrapped)
	return router
}

type fakePositionAuthorizationService struct {
	ancestorsByTarget      map[shared.DivisionID][]shared.DivisionID
	lastResolveTarget      shared.DivisionID
	resolveForDivisionFunc func(context.Context, shared.UserID, shared.DivisionID) (domainmembership.EffectivePermissions, error)
}

func (f *fakePositionAuthorizationService) ResolveForDivision(ctx context.Context, actorID shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
	f.lastResolveTarget = targetDivisionID
	if f.resolveForDivisionFunc != nil {
		return f.resolveForDivisionFunc(ctx, actorID, targetDivisionID)
	}
	if strings.HasPrefix(string(actorID), "admin") {
		return permissionAccess(targetDivisionID, domainrole.CanManagePositions), nil
	}
	if strings.HasPrefix(string(actorID), "scope-current") || strings.HasPrefix(string(actorID), "scope-desc") {
		scope := domainrole.ScopeCurrentDivision
		if strings.HasPrefix(string(actorID), "scope-desc") {
			scope = domainrole.ScopeCurrentAndDescendants
		}
		permission, err := domainrole.NewPermission(domainrole.CanManagePositions, scope)
		if err != nil {
			return domainmembership.EffectivePermissions{}, err
		}
		ancestors := f.ancestorsByTarget[targetDivisionID]
		return domainmembership.Calculate(
			[]domainmembership.MembershipContext{
				{
					DivisionID:  shared.DivisionID("parent-1"),
					Permissions: domainrole.NewPermissionSet([]domainrole.Permission{permission}),
				},
			},
			targetDivisionID,
			ancestors,
		), nil
	}
	return domainmembership.EffectivePermissions{}, nil
}

func withActor(actor httpsession.Actor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpsession.WithActor(r.Context(), actor)))
		})
	}
}

func permissionAccess(targetDivisionID shared.DivisionID, code domainrole.PermissionCode) domainmembership.EffectivePermissions {
	permission, err := domainrole.NewPermission(code, domainrole.ScopeCurrentAndDescendants)
	if err != nil {
		return domainmembership.EffectivePermissions{}
	}
	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  targetDivisionID,
				Permissions: domainrole.NewPermissionSet([]domainrole.Permission{permission}),
			},
		},
		targetDivisionID,
		nil,
	)
}

func TestChiImported(t *testing.T) {
	t.Parallel()
	r := chi.NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	r.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "ok") {
		t.Fatalf("expected router response, got %q", rec.Body.String())
	}
}

func TestPositionHandler_CreateScopeBoundaries(t *testing.T) {
	t.Parallel()

	t.Run("CURRENT_DIVISION denies child division for create", func(t *testing.T) {
		t.Parallel()

		createCalls := 0
		authz := &fakePositionAuthorizationService{
			ancestorsByTarget: map[shared.DivisionID][]shared.DivisionID{
				shared.DivisionID("child-1"): {shared.DivisionID("parent-1")},
			},
		}
		server := newPositionServerWithAuthz(t, &fakePositionRepo{
			createFunc: func(_ context.Context, _ appeventlog.Transaction, pos domainposition.Position) (domainposition.Position, error) {
				createCalls++
				return pos, nil
			},
		}, authz, withActor(httpsession.Actor{UserID: shared.UserID("scope-current")}))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions/child-1/positions", bytes.NewReader([]byte(`{"title":"Chair","role_id":"role-1"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected CURRENT_DIVISION deny 403 for child target, got %d: %s", rec.Code, rec.Body.String())
		}
		if authz.lastResolveTarget != shared.DivisionID("child-1") {
			t.Fatalf("expected resolver target child-1, got %q", authz.lastResolveTarget)
		}
		if createCalls != 0 {
			t.Fatalf("expected denied create to skip repository mutation, got %d calls", createCalls)
		}
	})

	t.Run("CURRENT_AND_DESCENDANTS allows child division for create", func(t *testing.T) {
		t.Parallel()

		createCalls := 0
		authz := &fakePositionAuthorizationService{
			ancestorsByTarget: map[shared.DivisionID][]shared.DivisionID{
				shared.DivisionID("child-1"): {shared.DivisionID("parent-1")},
			},
		}
		server := newPositionServerWithAuthz(t, &fakePositionRepo{
			createFunc: func(_ context.Context, _ appeventlog.Transaction, pos domainposition.Position) (domainposition.Position, error) {
				createCalls++
				return pos, nil
			},
		}, authz, withActor(httpsession.Actor{UserID: shared.UserID("scope-desc")}))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions/child-1/positions", bytes.NewReader([]byte(`{"title":"Chair","role_id":"role-1"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected CURRENT_AND_DESCENDANTS allow 201 for child target, got %d: %s", rec.Code, rec.Body.String())
		}
		if authz.lastResolveTarget != shared.DivisionID("child-1") {
			t.Fatalf("expected resolver target child-1, got %q", authz.lastResolveTarget)
		}
		if createCalls != 1 {
			t.Fatalf("expected allowed create to execute one repository mutation, got %d calls", createCalls)
		}
	})
}

func TestPositionHandler_ArchiveScopeBoundaries(t *testing.T) {
	t.Parallel()

	t.Run("CURRENT_DIVISION denies archive in child division", func(t *testing.T) {
		t.Parallel()

		archiveCalls := 0
		repo := &fakePositionRepo{
			getByIDFunc: func(_ context.Context, id shared.PositionID) (domainposition.Position, error) {
				return domainposition.Position{
					ID:         id,
					Title:      "Chair",
					RoleID:     shared.RoleID("role-1"),
					DivisionID: shared.DivisionID("child-1"),
				}, nil
			},
			archiveFunc: func(_ context.Context, _ appeventlog.Transaction, id shared.PositionID) (domainposition.Position, int, error) {
				archiveCalls++
				return domainposition.Position{
					ID:         id,
					Title:      "Chair",
					RoleID:     shared.RoleID("role-1"),
					DivisionID: shared.DivisionID("child-1"),
					IsArchived: true,
				}, 0, nil
			},
		}
		authz := &fakePositionAuthorizationService{
			ancestorsByTarget: map[shared.DivisionID][]shared.DivisionID{
				shared.DivisionID("child-1"): {shared.DivisionID("parent-1")},
			},
		}
		server := newPositionServerWithAuthz(t, repo, authz, withActor(httpsession.Actor{UserID: shared.UserID("scope-current")}))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/positions/pos-1/archive", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected CURRENT_DIVISION deny 403 for child archive, got %d: %s", rec.Code, rec.Body.String())
		}
		if authz.lastResolveTarget != shared.DivisionID("child-1") {
			t.Fatalf("expected resolver target child-1, got %q", authz.lastResolveTarget)
		}
		if archiveCalls != 0 {
			t.Fatalf("expected denied archive to skip repository mutation, got %d archive calls", archiveCalls)
		}
	})

	t.Run("CURRENT_AND_DESCENDANTS allows archive in child division", func(t *testing.T) {
		t.Parallel()

		archiveCalls := 0
		repo := &fakePositionRepo{
			getByIDFunc: func(_ context.Context, id shared.PositionID) (domainposition.Position, error) {
				return domainposition.Position{
					ID:         id,
					Title:      "Chair",
					RoleID:     shared.RoleID("role-1"),
					DivisionID: shared.DivisionID("child-1"),
				}, nil
			},
			archiveFunc: func(_ context.Context, _ appeventlog.Transaction, id shared.PositionID) (domainposition.Position, int, error) {
				archiveCalls++
				return domainposition.Position{
					ID:         id,
					Title:      "Chair",
					RoleID:     shared.RoleID("role-1"),
					DivisionID: shared.DivisionID("child-1"),
					IsArchived: true,
				}, 0, nil
			},
		}
		authz := &fakePositionAuthorizationService{
			ancestorsByTarget: map[shared.DivisionID][]shared.DivisionID{
				shared.DivisionID("child-1"): {shared.DivisionID("parent-1")},
			},
		}
		server := newPositionServerWithAuthz(t, repo, authz, withActor(httpsession.Actor{UserID: shared.UserID("scope-desc")}))
		req := httptest.NewRequest(http.MethodPost, "/api/v1/positions/pos-1/archive", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected CURRENT_AND_DESCENDANTS allow 200 for child archive, got %d: %s", rec.Code, rec.Body.String())
		}
		if authz.lastResolveTarget != shared.DivisionID("child-1") {
			t.Fatalf("expected resolver target child-1, got %q", authz.lastResolveTarget)
		}
		if archiveCalls != 1 {
			t.Fatalf("expected allowed archive to execute one repository mutation, got %d archive calls", archiveCalls)
		}
	})
}

func newPositionCommandService() *appeventlog.CommandService {
	return appeventlog.NewCommandService(&positionTxManager{}, &positionEventWriter{})
}

type positionTxManager struct{}

func (positionTxManager) Begin(context.Context) (appeventlog.Transaction, error) {
	return positionTx{}, nil
}

type positionTx struct{}

func (positionTx) Commit(context.Context) error   { return nil }
func (positionTx) Rollback(context.Context) error { return nil }

type positionEventWriter struct {
	err     error
	entries []domaineventlog.Entry
}

func (w *positionEventWriter) InsertWithTx(_ context.Context, _ appeventlog.Transaction, entry domaineventlog.Entry) (domaineventlog.Entry, error) {
	w.entries = append(w.entries, entry)
	if w.err != nil {
		return domaineventlog.Entry{}, w.err
	}
	return entry, nil
}

type positionProbeTx struct {
	afterCommit []func()
}

func (t *positionProbeTx) Commit(context.Context) error {
	for _, fn := range t.afterCommit {
		fn()
	}
	return nil
}

func (t *positionProbeTx) Rollback(context.Context) error { return nil }

func (t *positionProbeTx) OnCommit(fn func()) {
	t.afterCommit = append(t.afterCommit, fn)
}

type positionProbeTxManager struct {
	last *positionProbeTx
}

func (m *positionProbeTxManager) Begin(context.Context) (appeventlog.Transaction, error) {
	tx := &positionProbeTx{}
	m.last = tx
	return tx, nil
}

func TestCreatePosition_RollsBackWhenEventInsertFails(t *testing.T) {
	t.Parallel()

	txManager := &positionProbeTxManager{}
	writer := &positionEventWriter{err: errors.New("event insert failed")}
	commands := appeventlog.NewCommandService(txManager, writer)

	persisted := false
	repo := &fakePositionRepo{
		createFunc: func(_ context.Context, tx appeventlog.Transaction, pos domainposition.Position) (domainposition.Position, error) {
			staged, ok := tx.(*positionProbeTx)
			if !ok {
				t.Fatalf("unexpected tx type: %T", tx)
			}
			staged.OnCommit(func() { persisted = true })
			return pos, nil
		},
	}

	server := newPositionServerWithCommandService(t, repo, &fakePositionAuthorizationService{}, commands, withActor(httpsession.Actor{
		UserID: shared.UserID("admin-1"),
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions/div-1/positions", bytes.NewReader([]byte(`{"title":"Chair","role_id":"role-1"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
	if persisted {
		t.Fatal("expected mutation rollback when event insert fails")
	}
	if len(writer.entries) != 1 {
		t.Fatalf("event inserts = %d, want 1", len(writer.entries))
	}
}

func TestCreatePosition_WritesSingleExpectedEvent(t *testing.T) {
	t.Parallel()

	txManager := &positionProbeTxManager{}
	writer := &positionEventWriter{}
	commands := appeventlog.NewCommandService(txManager, writer)

	server := newPositionServerWithCommandService(t, &fakePositionRepo{}, &fakePositionAuthorizationService{}, commands, withActor(httpsession.Actor{
		UserID: shared.UserID("admin-1"),
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions/div-1/positions", bytes.NewReader([]byte(`{"title":"Chair","role_id":"role-1"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(writer.entries) != 1 {
		t.Fatalf("event inserts = %d, want 1", len(writer.entries))
	}
	entry := writer.entries[0]
	if entry.EventType != domaineventlog.EventPositionCreated {
		t.Fatalf("event type = %q, want %q", entry.EventType, domaineventlog.EventPositionCreated)
	}
	if entry.SubjectType != domaineventlog.SubjectPosition {
		t.Fatalf("subject type = %q, want %q", entry.SubjectType, domaineventlog.SubjectPosition)
	}
}

func newPositionServerWithCommandService(
	t *testing.T,
	repo httpposition.Repository,
	authz *fakePositionAuthorizationService,
	commands *appeventlog.CommandService,
	middleware ...func(http.Handler) http.Handler,
) http.Handler {
	t.Helper()

	handler := httpposition.NewHandler(repo, authz, commands)
	router := http.NewServeMux()
	chiRouter := httpposition.NewTestRouter(handler)

	var wrapped = chiRouter
	for i := len(middleware) - 1; i >= 0; i-- {
		wrapped = middleware[i](wrapped)
	}
	router.Handle("/", wrapped)
	return router
}
