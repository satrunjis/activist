package role

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appeventlog "activist-base/src/application/eventlog"
	domaineventlog "activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	httpsession "activist-base/src/transport/http/session"
)

func TestPermissionsFromRequestApplyCanonicalScopeModesForLegacyPayload(t *testing.T) {
	t.Parallel()

	perms, err := permissionsFromRequest([]permissionRequest{
		{Code: "can_manage_positions"},
		{Code: "can_create_subdivision"},
	})
	if err != nil {
		t.Fatalf("permissionsFromRequest() error = %v", err)
	}

	effective := domainmembership.Calculate(
		[]domainmembership.MembershipContext{{
			DivisionID:  "division-parent",
			Permissions: perms,
		}},
		"division-child",
		[]shared.DivisionID{"division-parent"},
	)

	if effective.Has(domainrole.CanManagePositions, false) {
		t.Fatal("expected CURRENT_DIVISION to deny descendants")
	}
	if !effective.Has(domainrole.CanCreateSubdivision, false) {
		t.Fatal("expected CURRENT_AND_DESCENDANTS to allow descendants")
	}
}

func TestPermissionsFromRequestUsesExplicitScopes(t *testing.T) {
	t.Parallel()

	perms, err := permissionsFromRequest([]permissionRequest{
		{Code: "can_manage_positions", Scope: "current_and_descendants"},
	})
	if err != nil {
		t.Fatalf("permissionsFromRequest() error = %v", err)
	}
	if !perms.Contains(domainrole.CanManagePositions, domainrole.ScopeCurrentAndDescendants) {
		t.Fatal("expected explicit scope to be preserved")
	}
}

func TestCreateReturnsValidationErrorForUnknownPermission(t *testing.T) {
	t.Parallel()

	handler := newRoleHandler(&fakeRoleRepo{}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return systemAdminAccess(t), nil
		},
	})
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "admin-1"}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte(`{
		"name":"Role",
		"permissions":["unknown_permission"]
	}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 validation error, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateAcceptsCanManageRolesPermission(t *testing.T) {
	t.Parallel()

	handler := newRoleHandler(&fakeRoleRepo{}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return systemAdminAccess(t), nil
		},
	})
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "admin-1"}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte(`{
		"name":"Role Manager",
		"permissions":["can_manage_roles"]
	}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 created, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateAcceptsCanViewAuditLogPermission(t *testing.T) {
	t.Parallel()

	handler := newRoleHandler(&fakeRoleRepo{}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return systemAdminAccess(t), nil
		},
	})
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "admin-1"}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte(`{
		"name":"Audit Reader",
		"permissions":["can_view_audit_log"]
	}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 created, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateRole_RollsBackWhenEventInsertFails(t *testing.T) {
	t.Parallel()

	txManager := &roleProbeTxManager{}
	writer := &roleEventWriter{err: errors.New("event insert failed")}
	commands := newRoleCommandServiceWith(txManager, writer)

	persisted := false
	handler := NewHandler(&fakeRoleRepo{
		createFunc: func(_ context.Context, tx appeventlog.Transaction, role domainrole.Role) (domainrole.Role, error) {
			staged, ok := tx.(*roleProbeTx)
			if !ok {
				t.Fatalf("unexpected tx type: %T", tx)
			}
			staged.OnCommit(func() { persisted = true })
			return role, nil
		},
	}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return systemAdminAccess(t), nil
		},
	}, commands)
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "admin-1"}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte(`{
		"name":"Role",
		"permissions":["can_manage_positions"]
	}`)))
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

func TestCreateRole_WritesSingleExpectedEvent(t *testing.T) {
	t.Parallel()

	writer := &roleEventWriter{}
	handler := newRoleHandlerWithCommandService(&fakeRoleRepo{}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return systemAdminAccess(t), nil
		},
	}, newRoleCommandServiceWith(&roleProbeTxManager{}, writer))
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "admin-1"}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader([]byte(`{
		"name":"Role",
		"permissions":["can_manage_positions"]
	}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(writer.entries) != 1 {
		t.Fatalf("event inserts = %d, want 1", len(writer.entries))
	}
	if writer.entries[0].EventType != domaineventlog.EventRoleCreated {
		t.Fatalf("event type = %q, want %q", writer.entries[0].EventType, domaineventlog.EventRoleCreated)
	}
	if writer.entries[0].SubjectType != domaineventlog.SubjectRole {
		t.Fatalf("subject type = %q, want %q", writer.entries[0].SubjectType, domaineventlog.SubjectRole)
	}
}

func TestListResponseIncludesRoles(t *testing.T) {
	t.Parallel()

	handler := newRoleHandler(&fakeRoleRepo{
		listFunc: func(_ context.Context, limit, offset int) ([]domainrole.Role, int64, error) {
			if limit != defaultRolesLimit {
				t.Fatalf("limit = %d, want %d", limit, defaultRolesLimit)
			}
			if offset != 0 {
				t.Fatalf("offset = %d, want 0", offset)
			}
			return []domainrole.Role{
				{ID: "role-leader", Name: "Leader"},
				{ID: "role-deputy", Name: "Deputy"},
			}, 2, nil
		},
	}, &fakeAuthorizationService{})
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "admin-1"}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEditRoleSucceeds(t *testing.T) {
	t.Parallel()

	existing := roleFixture(t, "role-1", "Existing", []domainrole.PermissionCode{domainrole.CanAddMember})
	handler := newRoleHandler(&fakeRoleRepo{
		getByIDFunc: func(context.Context, shared.RoleID) (domainrole.Role, error) { return existing, nil },
		updateFunc: func(_ context.Context, _ appeventlog.Transaction, role domainrole.Role) (domainrole.Role, error) {
			return role, nil
		},
	}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return systemAdminAccess(t), nil
		},
	})
	handler.nowFn = func() time.Time { return time.Date(2026, 4, 14, 12, 0, 0, 0, time.UTC) }
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "admin-1"}))

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/roles/role-1", bytes.NewReader([]byte(`{
		"name":"Renamed",
		"permissions":["can_manage_roles"]
	}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload roleResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload.Name != "Renamed" {
		t.Fatalf("name = %q, want %q", payload.Name, "Renamed")
	}
}

func TestEditRoleForbiddenWithoutCanManageRoles(t *testing.T) {
	t.Parallel()

	existing := roleFixture(t, "role-1", "Existing", nil)
	handler := newRoleHandler(&fakeRoleRepo{
		getByIDFunc: func(context.Context, shared.RoleID) (domainrole.Role, error) { return existing, nil },
	}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return emptyAccess(), nil
		},
	})
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "user-1"}))

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/roles/role-1", bytes.NewReader([]byte(`{
		"name":"Renamed"
	}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden, got %d: %s", rec.Code, rec.Body.String())
	}
	if code := responseErrorCode(t, rec.Body.Bytes()); code != "access.forbidden" {
		t.Fatalf("error code = %q, want %q", code, "access.forbidden")
	}
}

func TestEditRoleForbiddenForNonSystemAdmin(t *testing.T) {
	t.Parallel()

	existing := roleFixture(t, "role-1", "Existing", nil)
	handler := newRoleHandler(&fakeRoleRepo{
		getByIDFunc: func(context.Context, shared.RoleID) (domainrole.Role, error) { return existing, nil },
	}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return canManageRolesAccess(t), nil
		},
	})
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "user-1"}))

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/roles/role-1", bytes.NewReader([]byte(`{
		"permissions":["system_admin"]
	}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden, got %d: %s", rec.Code, rec.Body.String())
	}
	if code := responseErrorCode(t, rec.Body.Bytes()); code != "access.forbidden" {
		t.Fatalf("error code = %q, want %q", code, "access.forbidden")
	}
}

func TestDeleteRoleSucceeds(t *testing.T) {
	t.Parallel()

	called := false
	handler := newRoleHandler(&fakeRoleRepo{
		deleteFunc: func(context.Context, shared.RoleID) error {
			called = true
			return nil
		},
	}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return systemAdminAccess(t), nil
		},
	})
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "admin-1"}))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/roles/role-1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 no content, got %d: %s", rec.Code, rec.Body.String())
	}
	if !called {
		t.Fatal("expected repo.Delete to be called")
	}
}

func TestDeleteRoleForbiddenWithoutCanManageRoles(t *testing.T) {
	t.Parallel()

	handler := newRoleHandler(&fakeRoleRepo{}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return emptyAccess(), nil
		},
	})
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "user-1"}))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/roles/role-1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteRoleInUseReturns409(t *testing.T) {
	t.Parallel()

	handler := newRoleHandler(&fakeRoleRepo{
		deleteFunc: func(context.Context, shared.RoleID) error {
			return &shared.Error{
				Code:    "validation.role_in_use",
				Message: "role is in use",
			}
		},
	}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return systemAdminAccess(t), nil
		},
	})
	server := newRoleServer(handler, withActor(httpsession.Actor{UserID: "admin-1"}))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/roles/role-1", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %d: %s", rec.Code, rec.Body.String())
	}
}

type fakeRoleRepo struct {
	createFunc  func(context.Context, appeventlog.Transaction, domainrole.Role) (domainrole.Role, error)
	listFunc    func(context.Context, int, int) ([]domainrole.Role, int64, error)
	getByIDFunc func(context.Context, shared.RoleID) (domainrole.Role, error)
	updateFunc  func(context.Context, appeventlog.Transaction, domainrole.Role) (domainrole.Role, error)
	deleteFunc  func(context.Context, shared.RoleID) error
}

func (f *fakeRoleRepo) CreateWithTx(ctx context.Context, tx appeventlog.Transaction, role domainrole.Role) (domainrole.Role, error) {
	if f.createFunc != nil {
		return f.createFunc(ctx, tx, role)
	}
	return role, nil
}

func (f *fakeRoleRepo) List(ctx context.Context, limit, offset int) ([]domainrole.Role, int64, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, limit, offset)
	}
	return nil, 0, nil
}

func (f *fakeRoleRepo) GetByID(ctx context.Context, id shared.RoleID) (domainrole.Role, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return roleFixture(nil, id, "Existing", nil), nil
}

func (f *fakeRoleRepo) UpdateWithTx(ctx context.Context, tx appeventlog.Transaction, role domainrole.Role) (domainrole.Role, error) {
	if f.updateFunc != nil {
		return f.updateFunc(ctx, tx, role)
	}
	return role, nil
}

func (f *fakeRoleRepo) Delete(ctx context.Context, id shared.RoleID) error {
	if f.deleteFunc != nil {
		return f.deleteFunc(ctx, id)
	}
	return nil
}

type fakeAuthorizationService struct {
	resolveGlobalFunc func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error)
}

func (f *fakeAuthorizationService) ResolveGlobal(ctx context.Context, actorID shared.UserID) (domainmembership.EffectivePermissions, error) {
	if f.resolveGlobalFunc != nil {
		return f.resolveGlobalFunc(ctx, actorID)
	}
	return domainmembership.EffectivePermissions{}, nil
}

func newRoleHandler(repo Repository, authz AuthorizationService) *Handler {
	return newRoleHandlerWithCommandService(repo, authz, newRoleCommandService())
}

func newRoleHandlerWithCommandService(repo Repository, authz AuthorizationService, commands *appeventlog.CommandService) *Handler {
	return NewHandler(repo, authz, commands)
}

func newRoleServer(handler *Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	chiRouter := NewTestRouter(handler)
	var wrapped = chiRouter
	for i := len(middleware) - 1; i >= 0; i-- {
		wrapped = middleware[i](wrapped)
	}
	return wrapped
}

func withActor(actor httpsession.Actor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpsession.WithActor(r.Context(), actor)))
		})
	}
}

func systemAdminAccess(t *testing.T) domainmembership.EffectivePermissions {
	t.Helper()
	adminPermission, err := domainrole.NewPermission(domainrole.SystemAdmin, domainrole.ScopeCurrentAndDescendants)
	if err != nil {
		t.Fatalf("NewPermission() error = %v", err)
	}
	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{{
			DivisionID:  "division-root",
			Permissions: domainrole.NewPermissionSet([]domainrole.Permission{adminPermission}),
		}},
		"division-root",
		nil,
	)
}

func canManageRolesAccess(t *testing.T, extra ...domainrole.PermissionCode) domainmembership.EffectivePermissions {
	t.Helper()
	perms := make([]domainrole.Permission, 0, 1+len(extra))

	canManageRoles, err := domainrole.NewPermission(domainrole.CanManageRoles, domainrole.ScopeCurrentAndDescendants)
	if err != nil {
		t.Fatalf("NewPermission(CanManageRoles): %v", err)
	}
	perms = append(perms, canManageRoles)

	for _, code := range extra {
		scope, ok := domainrole.CanonicalScopeForPermission(code)
		if !ok {
			t.Fatalf("CanonicalScopeForPermission(%v) not found", code)
		}
		p, err := domainrole.NewPermission(code, scope)
		if err != nil {
			t.Fatalf("NewPermission(%v): %v", code, err)
		}
		perms = append(perms, p)
	}

	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{{
			DivisionID:  "division-root",
			Permissions: domainrole.NewPermissionSet(perms),
		}},
		"division-root",
		nil,
	)
}

func emptyAccess() domainmembership.EffectivePermissions {
	return domainmembership.Calculate(nil, "division-root", nil)
}

func roleFixture(t *testing.T, id shared.RoleID, name string, codes []domainrole.PermissionCode) domainrole.Role {
	var perms []domainrole.Permission
	for _, code := range codes {
		scope, ok := domainrole.CanonicalScopeForPermission(code)
		if !ok {
			if t != nil {
				t.Fatalf("CanonicalScopeForPermission(%v) not found", code)
			}
			continue
		}
		p, err := domainrole.NewPermission(code, scope)
		if err != nil {
			if t != nil {
				t.Fatalf("NewPermission(%v): %v", code, err)
			}
			continue
		}
		perms = append(perms, p)
	}

	return domainrole.Role{
		ID:          id,
		Name:        name,
		Permissions: domainrole.NewPermissionSet(perms),
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
}

func responseErrorCode(t *testing.T, body []byte) string {
	t.Helper()
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("json.Unmarshal(errorEnvelope) error = %v; body=%s", err, string(body))
	}
	return envelope.Error.Code
}

func newRoleCommandService() *appeventlog.CommandService {
	return newRoleCommandServiceWith(&roleTxManager{}, &roleEventWriter{})
}

func newRoleCommandServiceWith(txManager appeventlog.TxManager, writer appeventlog.EventWriter) *appeventlog.CommandService {
	return appeventlog.NewCommandService(txManager, writer)
}

type roleTxManager struct{}

func (roleTxManager) Begin(context.Context) (appeventlog.Transaction, error) {
	return roleTx{}, nil
}

type roleTx struct{}

func (roleTx) Commit(context.Context) error   { return nil }
func (roleTx) Rollback(context.Context) error { return nil }

type roleProbeTx struct {
	afterCommit []func()
}

func (t *roleProbeTx) Commit(context.Context) error {
	for _, fn := range t.afterCommit {
		fn()
	}
	return nil
}

func (t *roleProbeTx) Rollback(context.Context) error { return nil }

func (t *roleProbeTx) OnCommit(fn func()) {
	t.afterCommit = append(t.afterCommit, fn)
}

type roleProbeTxManager struct{}

func (roleProbeTxManager) Begin(context.Context) (appeventlog.Transaction, error) {
	return &roleProbeTx{}, nil
}

type roleEventWriter struct {
	err     error
	entries []domaineventlog.Entry
}

func (w *roleEventWriter) InsertWithTx(_ context.Context, _ appeventlog.Transaction, entry domaineventlog.Entry) (domaineventlog.Entry, error) {
	w.entries = append(w.entries, entry)
	if w.err != nil {
		return domaineventlog.Entry{}, w.err
	}
	return entry, nil
}
