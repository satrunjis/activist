package division_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appeventlog "activist-base/src/application/eventlog"
	domaindivision "activist-base/src/domain/division"
	domaineventlog "activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	postgres "activist-base/src/repository/postgres"
	httpdivision "activist-base/src/transport/http/division"
	httpsession "activist-base/src/transport/http/session"
)

func TestDivisionHandler_Create(t *testing.T) {
	t.Parallel()

	// POST /api/v1/divisions
	t.Run("returns 401 without actor", func(t *testing.T) {
		t.Parallel()

		repo := &fakeDivisionRepo{}
		server := newDivisionServer(t, repo)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions", bytes.NewReader([]byte(`{"id":"d-1","short_name":"HQ"}`)))
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("returns 409 when second root is requested", func(t *testing.T) {
		t.Parallel()

		existingRoot := domaindivision.Division{ID: shared.DivisionID("root-1"), ShortName: "Root"}
		repo := &fakeDivisionRepo{
			getRootFunc: func(context.Context) (domaindivision.Division, error) {
				return existingRoot, nil
			},
		}
		server := newDivisionServer(t, repo, withActor(httpsession.Actor{
			UserID: shared.UserID("admin-1"),
		}))

		req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions", bytes.NewReader([]byte(`{"id":"d-2","short_name":"Another Root"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("ignores client supplied id and generates server id", func(t *testing.T) {
		t.Parallel()

		var createdID shared.DivisionID
		repo := &fakeDivisionRepo{
			createFunc: func(_ context.Context, d domaindivision.Division) (domaindivision.Division, error) {
				createdID = d.ID
				return d, nil
			},
		}
		server := newDivisionServer(t, repo, withActor(httpsession.Actor{
			UserID: shared.UserID("admin-1"),
		}))

		req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions", bytes.NewReader([]byte(`{"id":"client-id","short_name":"HQ"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
		if createdID == "" {
			t.Fatal("expected generated id")
		}
		if createdID == "client-id" {
			t.Fatal("expected server to ignore client-supplied id")
		}
		if strings.Contains(string(createdID), "client") {
			t.Fatalf("unexpected generated id: %q", createdID)
		}
	})
}

func TestDivisionHandler_Patch(t *testing.T) {
	t.Parallel()

	// PATCH /api/v1/divisions/{division_id}
	t.Run("returns 404 for unknown division", func(t *testing.T) {
		t.Parallel()

		repo := &fakeDivisionRepo{
			getByIDFunc: func(context.Context, shared.DivisionID) (domaindivision.Division, error) {
				return domaindivision.Division{}, sql.ErrNoRows
			},
		}
		server := newDivisionServer(t, repo, withActor(httpsession.Actor{
			UserID: shared.UserID("editor-1"),
		}))

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/divisions/missing", bytes.NewReader([]byte(`{"short_name":"Updated"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("returns 422 for invalid body", func(t *testing.T) {
		t.Parallel()

		repo := &fakeDivisionRepo{}
		server := newDivisionServer(t, repo, withActor(httpsession.Actor{
			UserID: shared.UserID("editor-1"),
		}))

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/divisions/div-1", bytes.NewReader([]byte(`{"short_name":`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid JSON, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestDivisionHandler_GetTree(t *testing.T) {
	t.Parallel()

	// GET /api/v1/divisions/tree
	repo := &fakeDivisionRepo{
		listTreeFunc: func(context.Context, bool) ([]domaindivision.Division, error) {
			rootID := shared.DivisionID("root-1")
			return []domaindivision.Division{
				{
					ID:        rootID,
					ShortName: "Root",
					FullName:  "Root Full",
				},
				{
					ID:        shared.DivisionID("child-1"),
					ParentID:  &rootID,
					ShortName: "Child",
					FullName:  "Child Full",
				},
			}, nil
		},
	}
	server := newDivisionServer(t, repo, withActor(httpsession.Actor{
		UserID: shared.UserID("viewer-1"),
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/divisions/tree", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := payload["full_name"]; ok {
		t.Fatal("full_name must be hidden without card permission")
	}
}

type fakeDivisionRepo struct {
	getRootFunc           func(context.Context) (domaindivision.Division, error)
	getByIDFunc           func(context.Context, shared.DivisionID) (domaindivision.Division, error)
	createFunc            func(context.Context, domaindivision.Division) (domaindivision.Division, error)
	updateFunc            func(context.Context, domaindivision.Division) (domaindivision.Division, error)
	archiveCascadeFunc    func(context.Context, appeventlog.Transaction, shared.DivisionID) (domaindivision.Division, int, int, bool, error)
	getAncestorIDsFunc    func(context.Context, shared.DivisionID) ([]shared.DivisionID, error)
	listTreeFunc          func(context.Context, bool) ([]domaindivision.Division, error)
	listChildrenFunc      func(context.Context, shared.DivisionID) ([]postgres.DivisionChildItem, error)
	listRootChildrenFunc  func(context.Context) ([]postgres.DivisionChildItem, error)
}

func (f *fakeDivisionRepo) GetRoot(ctx context.Context) (domaindivision.Division, error) {
	if f.getRootFunc != nil {
		return f.getRootFunc(ctx)
	}
	return domaindivision.Division{}, sql.ErrNoRows
}

func (f *fakeDivisionRepo) GetByID(ctx context.Context, id shared.DivisionID) (domaindivision.Division, error) {
	if f.getByIDFunc != nil {
		return f.getByIDFunc(ctx, id)
	}
	return domaindivision.Division{
		ID:        id,
		ShortName: "Division",
	}, nil
}

func (f *fakeDivisionRepo) Create(ctx context.Context, division domaindivision.Division) (domaindivision.Division, error) {
	if f.createFunc != nil {
		return f.createFunc(ctx, division)
	}
	return division, nil
}

func (f *fakeDivisionRepo) Update(ctx context.Context, division domaindivision.Division) (domaindivision.Division, error) {
	if f.updateFunc != nil {
		return f.updateFunc(ctx, division)
	}
	return division, nil
}

func (f *fakeDivisionRepo) ArchiveCascadeWithTx(ctx context.Context, tx appeventlog.Transaction, id shared.DivisionID) (domaindivision.Division, int, int, bool, error) {
	if f.archiveCascadeFunc != nil {
		return f.archiveCascadeFunc(ctx, tx, id)
	}
	return domaindivision.Division{
		ID:         id,
		ShortName:  "Division",
		IsArchived: true,
	}, 0, 0, true, nil
}

func (f *fakeDivisionRepo) GetAncestorIDs(ctx context.Context, id shared.DivisionID) ([]shared.DivisionID, error) {
	if f.getAncestorIDsFunc != nil {
		return f.getAncestorIDsFunc(ctx, id)
	}
	return nil, nil
}

func (f *fakeDivisionRepo) ListTree(ctx context.Context, includeArchived bool) ([]domaindivision.Division, error) {
	if f.listTreeFunc != nil {
		return f.listTreeFunc(ctx, includeArchived)
	}
	return nil, nil
}

func (f *fakeDivisionRepo) ListChildren(ctx context.Context, parentID shared.DivisionID) ([]postgres.DivisionChildItem, error) {
	if f.listChildrenFunc != nil {
		return f.listChildrenFunc(ctx, parentID)
	}
	return nil, nil
}

func (f *fakeDivisionRepo) ListRootChildren(ctx context.Context) ([]postgres.DivisionChildItem, error) {
	if f.listRootChildrenFunc != nil {
		return f.listRootChildrenFunc(ctx)
	}
	return nil, nil
}

func newDivisionServer(t *testing.T, repo httpdivision.Repository, middleware ...func(http.Handler) http.Handler) http.Handler {
	return newDivisionServerWithAuthz(t, repo, &fakeDivisionAuthorizationService{}, middleware...)
}

func newDivisionServerWithAuthz(t *testing.T, repo httpdivision.Repository, authz *fakeDivisionAuthorizationService, middleware ...func(http.Handler) http.Handler) http.Handler {
	t.Helper()

	handler := httpdivision.NewHandler(repo, authz)
	router := http.NewServeMux()
	chiRouter := httpdivision.NewTestRouter(handler)

	var wrapped = chiRouter
	for i := len(middleware) - 1; i >= 0; i-- {
		wrapped = middleware[i](wrapped)
	}
	router.Handle("/", wrapped)
	return router
}

type fakeDivisionAuthorizationService struct {
	ancestorsByTarget map[shared.DivisionID][]shared.DivisionID
	lastResolveTarget shared.DivisionID
}

func (f *fakeDivisionAuthorizationService) ResolveForDivision(_ context.Context, actorID shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
	f.lastResolveTarget = targetDivisionID
	if strings.HasPrefix(string(actorID), "viewer") {
		return domainmembership.EffectivePermissions{}, nil
	}
	if strings.HasPrefix(string(actorID), "scope-current") || strings.HasPrefix(string(actorID), "scope-desc") {
		scope := role.ScopeCurrentDivision
		if strings.HasPrefix(string(actorID), "scope-desc") {
			scope = role.ScopeCurrentAndDescendants
		}
		permission, err := role.NewPermission(role.CanCreateSubdivision, scope)
		if err != nil {
			return domainmembership.EffectivePermissions{}, err
		}
		ancestors := f.ancestorsByTarget[targetDivisionID]
		return domainmembership.Calculate(
			[]domainmembership.MembershipContext{
				{
					DivisionID:  shared.DivisionID("parent-1"),
					Permissions: role.NewPermissionSet([]role.Permission{permission}),
				},
			},
			targetDivisionID,
			ancestors,
		), nil
	}
	return permissionAccess(targetDivisionID, role.SystemAdmin, role.CanCreateSubdivision, role.CanEditDivision, role.CanArchiveDivision), nil
}

func (f *fakeDivisionAuthorizationService) ResolveGlobal(_ context.Context, actorID shared.UserID) (domainmembership.EffectivePermissions, error) {
	if strings.HasPrefix(string(actorID), "viewer") {
		return domainmembership.EffectivePermissions{}, nil
	}
	return permissionAccess(shared.DivisionID(actorID), role.SystemAdmin, role.CanCreateSubdivision, role.CanEditDivision, role.CanArchiveDivision), nil
}

func withActor(actor httpsession.Actor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpsession.WithActor(r.Context(), actor)))
		})
	}
}

func permissionAccess(targetDivisionID shared.DivisionID, codes ...role.PermissionCode) domainmembership.EffectivePermissions {
	permissions := make([]role.Permission, 0, len(codes))
	for _, code := range codes {
		permission, err := role.NewPermission(code, role.ScopeCurrentAndDescendants)
		if err != nil {
			continue
		}
		permissions = append(permissions, permission)
	}
	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  targetDivisionID,
				Permissions: role.NewPermissionSet(permissions),
			},
		},
		targetDivisionID,
		nil,
	)
}

func TestMapsConflictStatus(t *testing.T) {
	t.Parallel()

	repo := &fakeDivisionRepo{
		createFunc: func(context.Context, domaindivision.Division) (domaindivision.Division, error) {
			return domaindivision.Division{}, domaindivision.ErrArchivedParent
		},
	}
	server := newDivisionServer(t, repo, withActor(httpsession.Actor{
		UserID: shared.UserID("editor-1"),
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions", bytes.NewReader([]byte(`{"id":"d-1","short_name":"Child","parent_id":"p-1"}`)))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}

	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Error.Code == "" {
		t.Fatal("expected error code in response")
	}
}

func TestReturnsUnprocessableEntityForDivisionValidation(t *testing.T) {
	t.Parallel()

	repo := &fakeDivisionRepo{
		createFunc: func(context.Context, domaindivision.Division) (domaindivision.Division, error) {
			return domaindivision.Division{}, &shared.Error{
				Code:    "division.some_rule",
				Message: "rule failed",
			}
		},
	}
	server := newDivisionServer(t, repo, withActor(httpsession.Actor{
		UserID: shared.UserID("editor-1"),
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions", bytes.NewReader([]byte(`{"id":"d-1","short_name":"Child","parent_id":"p-1"}`)))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRepositoryErrorPassesThroughInternal(t *testing.T) {
	t.Parallel()

	repo := &fakeDivisionRepo{
		listTreeFunc: func(context.Context, bool) ([]domaindivision.Division, error) {
			return nil, errors.New("db down")
		},
	}
	server := newDivisionServer(t, repo, withActor(httpsession.Actor{
		UserID: shared.UserID("viewer-1"),
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/divisions/tree", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDivisionHandler_CreateScopeBoundaries(t *testing.T) {
	t.Parallel()

	t.Run("CURRENT_DIVISION allows create in same division and denies child target", func(t *testing.T) {
		t.Parallel()

		repo := &fakeDivisionRepo{
			getByIDFunc: func(_ context.Context, id shared.DivisionID) (domaindivision.Division, error) {
				return domaindivision.Division{ID: id, ShortName: "Child"}, nil
			},
		}
		authz := &fakeDivisionAuthorizationService{
			ancestorsByTarget: map[shared.DivisionID][]shared.DivisionID{
				shared.DivisionID("child-1"): {shared.DivisionID("parent-1")},
			},
		}
		server := newDivisionServerWithAuthz(t, repo, authz, withActor(httpsession.Actor{UserID: shared.UserID("scope-current")}))

		allowReq := httptest.NewRequest(http.MethodPost, "/api/v1/divisions", bytes.NewReader([]byte(`{"short_name":"Child-A","parent_id":"parent-1"}`)))
		allowReq.Header.Set("Content-Type", "application/json")
		allowRec := httptest.NewRecorder()
		server.ServeHTTP(allowRec, allowReq)
		if allowRec.Code != http.StatusCreated {
			t.Fatalf("expected same-division allow 201, got %d: %s", allowRec.Code, allowRec.Body.String())
		}

		denyReq := httptest.NewRequest(http.MethodPost, "/api/v1/divisions", bytes.NewReader([]byte(`{"short_name":"Child-B","parent_id":"child-1"}`)))
		denyReq.Header.Set("Content-Type", "application/json")
		denyRec := httptest.NewRecorder()
		server.ServeHTTP(denyRec, denyReq)
		if denyRec.Code != http.StatusForbidden {
			t.Fatalf("expected CURRENT_DIVISION deny 403 on child target, got %d: %s", denyRec.Code, denyRec.Body.String())
		}
		if authz.lastResolveTarget != shared.DivisionID("child-1") {
			t.Fatalf("expected resolver target child-1, got %q", authz.lastResolveTarget)
		}
	})

	t.Run("CURRENT_AND_DESCENDANTS allows child target", func(t *testing.T) {
		t.Parallel()

		repo := &fakeDivisionRepo{
			getByIDFunc: func(_ context.Context, id shared.DivisionID) (domaindivision.Division, error) {
				return domaindivision.Division{ID: id, ShortName: "Child"}, nil
			},
		}
		authz := &fakeDivisionAuthorizationService{
			ancestorsByTarget: map[shared.DivisionID][]shared.DivisionID{
				shared.DivisionID("child-1"): {shared.DivisionID("parent-1")},
			},
		}
		server := newDivisionServerWithAuthz(t, repo, authz, withActor(httpsession.Actor{UserID: shared.UserID("scope-desc")}))

		req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions", bytes.NewReader([]byte(`{"short_name":"Child-C","parent_id":"child-1"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected CURRENT_AND_DESCENDANTS allow 201 on child target, got %d: %s", rec.Code, rec.Body.String())
		}
		if authz.lastResolveTarget != shared.DivisionID("child-1") {
			t.Fatalf("expected resolver target child-1, got %q", authz.lastResolveTarget)
		}
	})
}

func TestDivisionHandler_ArchiveIsIdempotentAndWritesSingleEvent(t *testing.T) {
	t.Parallel()

	txManager := &divisionProbeTxManager{}
	writer := &divisionEventWriter{}
	commands := appeventlog.NewCommandService(txManager, writer)

	archived := false
	repo := &fakeDivisionRepo{
		getByIDFunc: func(_ context.Context, id shared.DivisionID) (domaindivision.Division, error) {
			return domaindivision.Division{
				ID:         id,
				ShortName:  "Root",
				IsArchived: archived,
			}, nil
		},
		archiveCascadeFunc: func(_ context.Context, tx appeventlog.Transaction, id shared.DivisionID) (domaindivision.Division, int, int, bool, error) {
			staged, ok := tx.(*divisionProbeTx)
			if !ok {
				t.Fatalf("unexpected tx type: %T", tx)
			}
			staged.OnCommit(func() { archived = true })
			return domaindivision.Division{
				ID:         id,
				ShortName:  "Root",
				IsArchived: true,
			}, 2, 2, true, nil
		},
	}

	server := newDivisionServerWithCommandService(t, repo, &fakeDivisionAuthorizationService{}, commands, withActor(httpsession.Actor{
		UserID: shared.UserID("admin-1"),
	}))

	firstReq := httptest.NewRequest(http.MethodPost, "/api/v1/divisions/div-1/archive", nil)
	firstRec := httptest.NewRecorder()
	server.ServeHTTP(firstRec, firstReq)
	if firstRec.Code != http.StatusOK {
		t.Fatalf("expected first archive 200, got %d: %s", firstRec.Code, firstRec.Body.String())
	}

	secondReq := httptest.NewRequest(http.MethodPost, "/api/v1/divisions/div-1/archive", nil)
	secondRec := httptest.NewRecorder()
	server.ServeHTTP(secondRec, secondReq)
	if secondRec.Code != http.StatusOK {
		t.Fatalf("expected second archive 200, got %d: %s", secondRec.Code, secondRec.Body.String())
	}

	if len(writer.entries) != 1 {
		t.Fatalf("event inserts = %d, want 1", len(writer.entries))
	}
	if writer.entries[0].EventType != domaineventlog.EventDivisionArchived {
		t.Fatalf("event type = %q, want %q", writer.entries[0].EventType, domaineventlog.EventDivisionArchived)
	}
}

func TestDivisionHandler_ArchiveRollsBackWhenEventInsertFails(t *testing.T) {
	t.Parallel()

	txManager := &divisionProbeTxManager{}
	writer := &divisionEventWriter{err: errors.New("event insert failed")}
	commands := appeventlog.NewCommandService(txManager, writer)

	persisted := false
	repo := &fakeDivisionRepo{
		getByIDFunc: func(_ context.Context, id shared.DivisionID) (domaindivision.Division, error) {
			return domaindivision.Division{
				ID:         id,
				ShortName:  "Root",
				IsArchived: false,
			}, nil
		},
		archiveCascadeFunc: func(_ context.Context, tx appeventlog.Transaction, id shared.DivisionID) (domaindivision.Division, int, int, bool, error) {
			staged, ok := tx.(*divisionProbeTx)
			if !ok {
				t.Fatalf("unexpected tx type: %T", tx)
			}
			staged.OnCommit(func() { persisted = true })
			return domaindivision.Division{
				ID:         id,
				ShortName:  "Root",
				IsArchived: true,
			}, 1, 1, true, nil
		},
	}

	server := newDivisionServerWithCommandService(t, repo, &fakeDivisionAuthorizationService{}, commands, withActor(httpsession.Actor{
		UserID: shared.UserID("admin-1"),
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions/div-1/archive", nil)
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

func TestDivisionHandler_ArchiveForbiddenDoesNotMutate(t *testing.T) {
	t.Parallel()

	archiveCalls := 0
	repo := &fakeDivisionRepo{
		getByIDFunc: func(_ context.Context, id shared.DivisionID) (domaindivision.Division, error) {
			return domaindivision.Division{
				ID:        id,
				ShortName: "Child",
			}, nil
		},
		archiveCascadeFunc: func(_ context.Context, _ appeventlog.Transaction, id shared.DivisionID) (domaindivision.Division, int, int, bool, error) {
			archiveCalls++
			return domaindivision.Division{
				ID:         id,
				ShortName:  "Child",
				IsArchived: true,
			}, 0, 0, true, nil
		},
	}
	server := newDivisionServer(t, repo, withActor(httpsession.Actor{
		UserID: shared.UserID("viewer-1"),
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/divisions/div-1/archive", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
	if archiveCalls != 0 {
		t.Fatalf("expected no archive mutation calls, got %d", archiveCalls)
	}
}

func newDivisionServerWithCommandService(
	t *testing.T,
	repo httpdivision.Repository,
	authz *fakeDivisionAuthorizationService,
	commands *appeventlog.CommandService,
	middleware ...func(http.Handler) http.Handler,
) http.Handler {
	t.Helper()

	handler := httpdivision.NewHandler(repo, authz, commands)
	router := http.NewServeMux()
	chiRouter := httpdivision.NewTestRouter(handler)

	var wrapped = chiRouter
	for i := len(middleware) - 1; i >= 0; i-- {
		wrapped = middleware[i](wrapped)
	}
	router.Handle("/", wrapped)
	return router
}

type divisionProbeTx struct {
	afterCommit []func()
}

func (t *divisionProbeTx) Commit(context.Context) error {
	for _, fn := range t.afterCommit {
		fn()
	}
	return nil
}

func (t *divisionProbeTx) Rollback(context.Context) error { return nil }

func (t *divisionProbeTx) OnCommit(fn func()) {
	t.afterCommit = append(t.afterCommit, fn)
}

type divisionProbeTxManager struct{}

func (divisionProbeTxManager) Begin(context.Context) (appeventlog.Transaction, error) {
	return &divisionProbeTx{}, nil
}

type divisionEventWriter struct {
	err     error
	entries []domaineventlog.Entry
}

func (w *divisionEventWriter) InsertWithTx(_ context.Context, _ appeventlog.Transaction, entry domaineventlog.Entry) (domaineventlog.Entry, error) {
	w.entries = append(w.entries, entry)
	if w.err != nil {
		return domaineventlog.Entry{}, w.err
	}
	return entry, nil
}

func TestDivisionHandler_GetChildren(t *testing.T) {
	t.Parallel()

	child1 := postgres.DivisionChildItem{
		Division:      domaindivision.Division{ID: shared.DivisionID("child-1"), ShortName: "Child One"},
		HasChildren:   true,
		ChildrenCount: 2,
	}
	child2 := postgres.DivisionChildItem{
		Division:      domaindivision.Division{ID: shared.DivisionID("child-2"), ShortName: "Child Two"},
		HasChildren:   false,
		ChildrenCount: 0,
	}

	t.Run("returns 401 without actor", func(t *testing.T) {
		t.Parallel()

		repo := &fakeDivisionRepo{}
		server := newDivisionServer(t, repo)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/divisions", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("omitted parent_id returns root children", func(t *testing.T) {
		t.Parallel()

		repo := &fakeDivisionRepo{
			listRootChildrenFunc: func(context.Context) ([]postgres.DivisionChildItem, error) {
				return []postgres.DivisionChildItem{child1, child2}, nil
			},
		}
		server := newDivisionServer(t, repo, withActor(httpsession.Actor{UserID: shared.UserID("viewer-1")}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/divisions", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var payload map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		items, ok := payload["items"].([]any)
		if !ok || len(items) != 2 {
			t.Fatalf("expected 2 items, got %v", payload["items"])
		}
	})

	t.Run("parent_id=root returns root children", func(t *testing.T) {
		t.Parallel()

		called := false
		repo := &fakeDivisionRepo{
			listRootChildrenFunc: func(context.Context) ([]postgres.DivisionChildItem, error) {
				called = true
				return []postgres.DivisionChildItem{child1}, nil
			},
		}
		server := newDivisionServer(t, repo, withActor(httpsession.Actor{UserID: shared.UserID("viewer-1")}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/divisions?parent_id=root", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if !called {
			t.Fatal("expected ListRootChildren to be called")
		}
	})

	t.Run("parent_id=<id> returns children of that division", func(t *testing.T) {
		t.Parallel()

		var capturedParent shared.DivisionID
		repo := &fakeDivisionRepo{
			listChildrenFunc: func(_ context.Context, id shared.DivisionID) ([]postgres.DivisionChildItem, error) {
				capturedParent = id
				return []postgres.DivisionChildItem{child2}, nil
			},
		}
		server := newDivisionServer(t, repo, withActor(httpsession.Actor{UserID: shared.UserID("viewer-1")}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/divisions?parent_id=parent-1", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if capturedParent != shared.DivisionID("parent-1") {
			t.Fatalf("expected parent-1, got %q", capturedParent)
		}
	})

	t.Run("repository error returns 500", func(t *testing.T) {
		t.Parallel()

		repo := &fakeDivisionRepo{
			listRootChildrenFunc: func(context.Context) ([]postgres.DivisionChildItem, error) {
				return nil, errors.New("db error")
			},
		}
		server := newDivisionServer(t, repo, withActor(httpsession.Actor{UserID: shared.UserID("viewer-1")}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/divisions", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("has_children metadata propagated in response", func(t *testing.T) {
		t.Parallel()

		repo := &fakeDivisionRepo{
			listRootChildrenFunc: func(context.Context) ([]postgres.DivisionChildItem, error) {
				return []postgres.DivisionChildItem{child1}, nil
			},
		}
		server := newDivisionServer(t, repo, withActor(httpsession.Actor{UserID: shared.UserID("viewer-1")}))

		req := httptest.NewRequest(http.MethodGet, "/api/v1/divisions", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var payload map[string]any
		if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		items := payload["items"].([]any)
		item := items[0].(map[string]any)
		if item["has_children"] != true {
			t.Fatalf("expected has_children=true, got %v", item["has_children"])
		}
		if item["children_count"].(float64) != 2 {
			t.Fatalf("expected children_count=2, got %v", item["children_count"])
		}
	})
}
