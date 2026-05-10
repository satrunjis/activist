package membership_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appeventlog "activist-base/src/application/eventlog"
	appmembership "activist-base/src/application/membership"
	domaineventlog "activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
	httpmembership "activist-base/src/transport/http/membership"
	httpsession "activist-base/src/transport/http/session"
)

func TestAssignMember_OK(t *testing.T) {
	t.Parallel()

	repo := &fakeMembershipRepo{
		assignFunc: func(_ context.Context, _ appeventlog.Transaction, userID shared.UserID, position domainposition.Position, access domainmembership.EffectivePermissions, actorID shared.UserID) (appmembership.AssignResult, error) {
			if !access.Has(role.CanAssignPosition, false) {
				t.Fatal("expected resolver-derived can_assign_position access")
			}
			return appmembership.AssignResult{Membership: domainmembership.Membership{UserID: userID, PositionID: position.ID}}, nil
		},
	}
	authz := &fakeAuthorizationService{
		resolveForDivisionFunc: func(_ context.Context, actorID shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
			if actorID != shared.UserID("admin-1") {
				t.Fatalf("unexpected actor id: %s", actorID)
			}
			if targetDivisionID != shared.DivisionID("div-1") {
				t.Fatalf("unexpected division id: %s", targetDivisionID)
			}
			return permissionAccess(t, role.CanAssignPosition, targetDivisionID), nil
		},
	}
	server := newMembershipServer(t, repo, authz, withActor(httpsession.Actor{UserID: shared.UserID("admin-1")}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/memberships", bytes.NewReader([]byte(`{"user_id":"u1","position_id":"p1"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAssignMember_AlreadyExists(t *testing.T) {
	t.Parallel()

	repo := &fakeMembershipRepo{
		assignFunc: func(_ context.Context, _ appeventlog.Transaction, _ shared.UserID, _ domainposition.Position, _ domainmembership.EffectivePermissions, _ shared.UserID) (appmembership.AssignResult, error) {
			return appmembership.AssignResult{}, domainmembership.ErrAlreadyExists
		},
	}
	server := newMembershipServer(t, repo, &fakeAuthorizationService{
		resolveForDivisionFunc: func(_ context.Context, _ shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
			return permissionAccess(t, role.CanAssignPosition, targetDivisionID), nil
		},
	}, withActor(httpsession.Actor{UserID: shared.UserID("admin-1")}))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/memberships", bytes.NewReader([]byte(`{"user_id":"u1","position_id":"p1"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAssignStandardRoleAllowedWithCanAssignPosition(t *testing.T) {
	t.Parallel()

	repo := &fakeMembershipRepo{
		getPositionFunc: func(context.Context, shared.PositionID) (domainposition.Position, error) {
			return domainposition.Position{
				ID:         shared.PositionID("p-standard"),
				Title:      "Standard",
				RoleID:     shared.RoleID("role-standard"),
				DivisionID: shared.DivisionID("div-1"),
			}, nil
		},
		assignFunc: func(_ context.Context, _ appeventlog.Transaction, userID shared.UserID, position domainposition.Position, access domainmembership.EffectivePermissions, _ shared.UserID) (appmembership.AssignResult, error) {
			if !access.Has(role.CanAssignPosition, false) {
				t.Fatal("expected can_assign_position permission for standard role assignment")
			}
			return appmembership.AssignResult{Membership: domainmembership.Membership{UserID: userID, PositionID: position.ID}}, nil
		},
	}

	server := newMembershipServer(t, repo, &fakeAuthorizationService{
		resolveForDivisionFunc: func(_ context.Context, _ shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
			return permissionAccess(t, role.CanAssignPosition, targetDivisionID), nil
		},
	}, withActor(httpsession.Actor{UserID: shared.UserID("operator-3")}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memberships", bytes.NewReader([]byte(`{"user_id":"u3","position_id":"p-standard"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for standard role assignment, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRemoveMember_OK(t *testing.T) {
	t.Parallel()

	repo := &fakeMembershipRepo{}
	server := newMembershipServer(t, repo, &fakeAuthorizationService{
		resolveForDivisionFunc: func(_ context.Context, _ shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
			return permissionAccess(t, role.CanRemoveMember, targetDivisionID), nil
		},
	}, withActor(httpsession.Actor{UserID: shared.UserID("admin-1")}))
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/positions/p1/members/u1", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRemoveMember_NotFound(t *testing.T) {
	t.Parallel()

	repo := &fakeMembershipRepo{
		removeFunc: func(context.Context, appeventlog.Transaction, shared.UserID, shared.PositionID, domainmembership.EffectivePermissions, shared.UserID) (appmembership.RemoveResult, error) {
			return appmembership.RemoveResult{}, domainmembership.ErrNotFound
		},
	}
	server := newMembershipServer(t, repo, &fakeAuthorizationService{
		resolveForDivisionFunc: func(_ context.Context, _ shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
			return permissionAccess(t, role.CanRemoveMember, targetDivisionID), nil
		},
	}, withActor(httpsession.Actor{UserID: shared.UserID("admin-1")}))
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/positions/p1/members/u1", nil)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAssignMember_RollsBackWhenEventInsertFails(t *testing.T) {
	t.Parallel()

	txManager := &membershipProbeTxManager{}
	writer := &membershipEventWriter{err: errors.New("event insert failed")}
	commands := newMembershipCommandServiceWith(txManager, writer)

	persisted := false
	repo := &fakeMembershipRepo{
		assignFunc: func(_ context.Context, tx appeventlog.Transaction, userID shared.UserID, position domainposition.Position, _ domainmembership.EffectivePermissions, _ shared.UserID) (appmembership.AssignResult, error) {
			staged, ok := tx.(*membershipProbeTx)
			if !ok {
				t.Fatalf("unexpected tx type: %T", tx)
			}
			staged.OnCommit(func() { persisted = true })

			event := membershipEvent(t, domainuser.User{ID: userID}, position)
			return appmembership.AssignResult{
				Membership: domainmembership.Membership{UserID: userID, PositionID: position.ID},
				Event:      event,
			}, nil
		},
	}
	server := newMembershipServerWithCommandService(t, repo, &fakeAuthorizationService{
		resolveForDivisionFunc: func(_ context.Context, _ shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
			return permissionAccess(t, role.CanAssignPosition, targetDivisionID), nil
		},
	}, commands, withActor(httpsession.Actor{UserID: shared.UserID("admin-1")}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memberships", bytes.NewReader([]byte(`{"user_id":"u1","position_id":"p1"}`)))
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

func TestAssignMember_WritesSingleExpectedEvent(t *testing.T) {
	t.Parallel()

	writer := &membershipEventWriter{}
	commands := newMembershipCommandServiceWith(&membershipProbeTxManager{}, writer)

	repo := &fakeMembershipRepo{
		assignFunc: func(_ context.Context, _ appeventlog.Transaction, userID shared.UserID, position domainposition.Position, _ domainmembership.EffectivePermissions, _ shared.UserID) (appmembership.AssignResult, error) {
			event := membershipEvent(t, domainuser.User{ID: userID}, position)
			return appmembership.AssignResult{
				Membership: domainmembership.Membership{UserID: userID, PositionID: position.ID},
				Event:      event,
			}, nil
		},
	}
	server := newMembershipServerWithCommandService(t, repo, &fakeAuthorizationService{
		resolveForDivisionFunc: func(_ context.Context, _ shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
			return permissionAccess(t, role.CanAssignPosition, targetDivisionID), nil
		},
	}, commands, withActor(httpsession.Actor{UserID: shared.UserID("admin-1")}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/memberships", bytes.NewReader([]byte(`{"user_id":"u1","position_id":"p1"}`)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if len(writer.entries) != 1 {
		t.Fatalf("event inserts = %d, want 1", len(writer.entries))
	}
	if writer.entries[0].EventType != domaineventlog.EventPositionAssigned {
		t.Fatalf("event type = %q, want %q", writer.entries[0].EventType, domaineventlog.EventPositionAssigned)
	}
	if writer.entries[0].SubjectType != domaineventlog.SubjectMembership {
		t.Fatalf("subject type = %q, want %q", writer.entries[0].SubjectType, domaineventlog.SubjectMembership)
	}
}

type fakeMembershipRepo struct {
	getPositionFunc func(context.Context, shared.PositionID) (domainposition.Position, error)
	getUserFunc     func(context.Context, shared.UserID) (domainuser.User, error)
	assignFunc      func(context.Context, appeventlog.Transaction, shared.UserID, domainposition.Position, domainmembership.EffectivePermissions, shared.UserID) (appmembership.AssignResult, error)
	removeFunc      func(context.Context, appeventlog.Transaction, shared.UserID, shared.PositionID, domainmembership.EffectivePermissions, shared.UserID) (appmembership.RemoveResult, error)
}

func (f *fakeMembershipRepo) GetPositionByID(ctx context.Context, id shared.PositionID) (domainposition.Position, error) {
	if f.getPositionFunc != nil {
		return f.getPositionFunc(ctx, id)
	}
	return domainposition.Position{
		ID:         id,
		Title:      "Chair",
		RoleID:     shared.RoleID("role-1"),
		DivisionID: shared.DivisionID("div-1"),
	}, nil
}

func (f *fakeMembershipRepo) GetUserByID(ctx context.Context, id shared.UserID) (domainuser.User, error) {
	if f.getUserFunc != nil {
		return f.getUserFunc(ctx, id)
	}
	return domainuser.User{ID: id, Login: "user", PasswordHash: "hash", FirstName: "User"}, nil
}

func (f *fakeMembershipRepo) AssignWithCountCheckWithTx(ctx context.Context, tx appeventlog.Transaction, userID shared.UserID, position domainposition.Position, access domainmembership.EffectivePermissions, actorID shared.UserID) (appmembership.AssignResult, error) {
	if f.assignFunc != nil {
		return f.assignFunc(ctx, tx, userID, position, access, actorID)
	}
	return appmembership.AssignResult{
		Membership: domainmembership.Membership{UserID: userID, PositionID: position.ID},
	}, nil
}

func (f *fakeMembershipRepo) RemoveWithTx(ctx context.Context, tx appeventlog.Transaction, userID shared.UserID, positionID shared.PositionID, access domainmembership.EffectivePermissions, actorID shared.UserID) (appmembership.RemoveResult, error) {
	if f.removeFunc != nil {
		return f.removeFunc(ctx, tx, userID, positionID, access, actorID)
	}
	return appmembership.RemoveResult{}, nil
}

func newMembershipServer(
	t *testing.T,
	repo httpmembership.Repository,
	authz httpmembership.AuthorizationService,
	middleware ...func(http.Handler) http.Handler,
) http.Handler {
	return newMembershipServerWithCommandService(t, repo, authz, newMembershipCommandService(), middleware...)
}

func newMembershipServerWithCommandService(
	t *testing.T,
	repo httpmembership.Repository,
	authz httpmembership.AuthorizationService,
	commands *appeventlog.CommandService,
	middleware ...func(http.Handler) http.Handler,
) http.Handler {
	t.Helper()

	handler := httpmembership.NewHandler(repo, authz, commands)
	router := http.NewServeMux()
	chiRouter := httpmembership.NewTestRouter(handler)

	var wrapped = chiRouter
	for i := len(middleware) - 1; i >= 0; i-- {
		wrapped = middleware[i](wrapped)
	}
	router.Handle("/", wrapped)
	return router
}

type fakeAuthorizationService struct {
	resolveForDivisionFunc func(context.Context, shared.UserID, shared.DivisionID) (domainmembership.EffectivePermissions, error)
}

func (f *fakeAuthorizationService) ResolveForDivision(ctx context.Context, actorID shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
	if f.resolveForDivisionFunc != nil {
		return f.resolveForDivisionFunc(ctx, actorID, targetDivisionID)
	}
	return domainmembership.EffectivePermissions{}, nil
}

func (f *fakeAuthorizationService) ResolveForUser(context.Context, shared.UserID, shared.UserID) (domainmembership.EffectivePermissions, error) {
	return domainmembership.EffectivePermissions{}, nil
}

func (f *fakeAuthorizationService) ResolveGlobal(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
	return domainmembership.EffectivePermissions{}, nil
}

func permissionAccess(t *testing.T, code role.PermissionCode, targetDivisionID shared.DivisionID) domainmembership.EffectivePermissions {
	t.Helper()

	permission, err := role.NewPermission(code, role.ScopeCurrentAndDescendants)
	if err != nil {
		t.Fatalf("new permission: %v", err)
	}

	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{
			{
				DivisionID:  targetDivisionID,
				Permissions: role.NewPermissionSet([]role.Permission{permission}),
			},
		},
		targetDivisionID,
		nil,
	)
}

func withActor(actor httpsession.Actor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpsession.WithActor(r.Context(), actor)))
		})
	}
}

func newMembershipCommandService() *appeventlog.CommandService {
	return newMembershipCommandServiceWith(&membershipTxManager{}, &membershipEventWriter{})
}

func newMembershipCommandServiceWith(txManager appeventlog.TxManager, writer appeventlog.EventWriter) *appeventlog.CommandService {
	return appeventlog.NewCommandService(txManager, writer)
}

type membershipTxManager struct{}

func (membershipTxManager) Begin(context.Context) (appeventlog.Transaction, error) {
	return membershipTx{}, nil
}

type membershipTx struct{}

func (membershipTx) Commit(context.Context) error   { return nil }
func (membershipTx) Rollback(context.Context) error { return nil }

type membershipProbeTx struct {
	afterCommit []func()
}

func (t *membershipProbeTx) Commit(context.Context) error {
	for _, fn := range t.afterCommit {
		fn()
	}
	return nil
}

func (t *membershipProbeTx) Rollback(context.Context) error { return nil }

func (t *membershipProbeTx) OnCommit(fn func()) {
	t.afterCommit = append(t.afterCommit, fn)
}

type membershipProbeTxManager struct{}

func (membershipProbeTxManager) Begin(context.Context) (appeventlog.Transaction, error) {
	return &membershipProbeTx{}, nil
}

type membershipEventWriter struct {
	err     error
	entries []domaineventlog.Entry
}

func (w *membershipEventWriter) InsertWithTx(_ context.Context, _ appeventlog.Transaction, entry domaineventlog.Entry) (domaineventlog.Entry, error) {
	w.entries = append(w.entries, entry)
	if w.err != nil {
		return domaineventlog.Entry{}, w.err
	}
	return entry, nil
}

func membershipEvent(t *testing.T, user domainuser.User, position domainposition.Position) domaineventlog.Entry {
	t.Helper()
	payload, err := domaineventlog.NewPayload(map[string]string{
		"user_id":        string(user.ID),
		"position_id":    string(position.ID),
		"position_title": position.Title,
		"division_id":    string(position.DivisionID),
	})
	if err != nil {
		t.Fatalf("NewPayload() error = %v", err)
	}
	entry, err := domaineventlog.NewEntry(
		shared.UserID("admin-1"),
		domaineventlog.EventPositionAssigned,
		domaineventlog.SubjectMembership,
		string(user.ID)+":"+string(position.ID),
		payload,
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("NewEntry() error = %v", err)
	}
	return entry
}
