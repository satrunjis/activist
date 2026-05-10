package eventlog

import (
	"context"
	"encoding/json"
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

func TestListReturns403WithoutViewAuditLogPermission(t *testing.T) {
	t.Parallel()

	handler := NewHandler(&fakeEventLogRepo{}, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return domainmembership.Calculate(nil, "division-root", nil), nil
		},
	})
	server := newEventLogServer(handler, withActor("actor-1"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/eventlog", nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 forbidden, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListForwardsAllowedFiltersAndPagination(t *testing.T) {
	t.Parallel()

	repo := &fakeEventLogRepo{
		listFunc: func(_ context.Context, input appeventlog.ListInput) (appeventlog.ListResult, error) {
			if input.EventType == nil || *input.EventType != domaineventlog.EventDivisionArchived {
				t.Fatalf("event_type not forwarded")
			}
			if input.SubjectType == nil || *input.SubjectType != domaineventlog.SubjectDivision {
				t.Fatalf("subject_type not forwarded")
			}
			if input.SubjectID == nil || *input.SubjectID != "division-1" {
				t.Fatalf("subject_id not forwarded")
			}
			if input.Limit != 200 {
				t.Fatalf("limit = %d, want 200", input.Limit)
			}
			if input.Offset != 5 {
				t.Fatalf("offset = %d, want 5", input.Offset)
			}
			return appeventlog.ListResult{
				Items: []domaineventlog.Entry{eventFixture(t)},
				Total: 1,
			}, nil
		},
	}
	handler := NewHandler(repo, &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return viewAuditLogAccess(t), nil
		},
	})
	server := newEventLogServer(handler, withActor("actor-1"))

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/eventlog?event_type=division_archived&subject_type=division&subject_id=division-1&limit=500&offset=5",
		nil,
	)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListSupportsEachAllowedFilterIndividually(t *testing.T) {
	t.Parallel()

	t.Run("event_type only", func(t *testing.T) {
		t.Parallel()

		repo := &fakeEventLogRepo{
			listFunc: func(_ context.Context, input appeventlog.ListInput) (appeventlog.ListResult, error) {
				if input.EventType == nil || *input.EventType != domaineventlog.EventRoleCreated {
					t.Fatalf("event_type not forwarded")
				}
				if input.SubjectType != nil {
					t.Fatalf("unexpected subject_type: %v", *input.SubjectType)
				}
				if input.SubjectID != nil {
					t.Fatalf("unexpected subject_id: %s", *input.SubjectID)
				}
				return appeventlog.ListResult{Items: []domaineventlog.Entry{}, Total: 0, Limit: input.Limit, Offset: input.Offset}, nil
			},
		}
		server := newEventLogServer(NewHandler(repo, authWithViewAudit(t)), withActor("actor-1"))
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/eventlog?event_type=role_created", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("subject_type only", func(t *testing.T) {
		t.Parallel()

		repo := &fakeEventLogRepo{
			listFunc: func(_ context.Context, input appeventlog.ListInput) (appeventlog.ListResult, error) {
				if input.SubjectType == nil || *input.SubjectType != domaineventlog.SubjectMembership {
					t.Fatalf("subject_type not forwarded")
				}
				return appeventlog.ListResult{Items: []domaineventlog.Entry{}, Total: 0, Limit: input.Limit, Offset: input.Offset}, nil
			},
		}
		server := newEventLogServer(NewHandler(repo, authWithViewAudit(t)), withActor("actor-1"))
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/eventlog?subject_type=membership", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("subject_id only", func(t *testing.T) {
		t.Parallel()

		repo := &fakeEventLogRepo{
			listFunc: func(_ context.Context, input appeventlog.ListInput) (appeventlog.ListResult, error) {
				if input.SubjectID == nil || *input.SubjectID != "subject-42" {
					t.Fatalf("subject_id not forwarded")
				}
				return appeventlog.ListResult{Items: []domaineventlog.Entry{}, Total: 0, Limit: input.Limit, Offset: input.Offset}, nil
			},
		}
		server := newEventLogServer(NewHandler(repo, authWithViewAudit(t)), withActor("actor-1"))
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/eventlog?subject_id=subject-42", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestListUsesDefaultLimit50(t *testing.T) {
	t.Parallel()

	server := newEventLogServer(NewHandler(&fakeEventLogRepo{
		listFunc: func(_ context.Context, input appeventlog.ListInput) (appeventlog.ListResult, error) {
			if input.Limit != 50 {
				t.Fatalf("limit = %d, want 50", input.Limit)
			}
			return appeventlog.ListResult{
				Items:  []domaineventlog.Entry{},
				Total:  0,
				Limit:  input.Limit,
				Offset: input.Offset,
			}, nil
		},
	}, authWithViewAudit(t)), withActor("actor-1"))

	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/eventlog", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Limit int `json:"limit"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if payload.Limit != 50 {
		t.Fatalf("response limit = %d, want 50", payload.Limit)
	}
}

func TestListRejectsUnexpectedQueryParam(t *testing.T) {
	t.Parallel()

	server := newEventLogServer(NewHandler(&fakeEventLogRepo{}, authWithViewAudit(t)), withActor("actor-1"))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/eventlog?division_id=div-1", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if errorCode(t, rec.Body.Bytes()) != "validation.query" {
		t.Fatalf("error code = %q, want %q", errorCode(t, rec.Body.Bytes()), "validation.query")
	}
}

func TestListRejectsNegativeOffset(t *testing.T) {
	t.Parallel()

	server := newEventLogServer(NewHandler(&fakeEventLogRepo{}, authWithViewAudit(t)), withActor("actor-1"))
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/eventlog?offset=-1", nil))

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if errorCode(t, rec.Body.Bytes()) != "validation.offset" {
		t.Fatalf("error code = %q, want %q", errorCode(t, rec.Body.Bytes()), "validation.offset")
	}
}

type fakeEventLogRepo struct {
	listFunc func(context.Context, appeventlog.ListInput) (appeventlog.ListResult, error)
}

func (f *fakeEventLogRepo) List(ctx context.Context, input appeventlog.ListInput) (appeventlog.ListResult, error) {
	if f.listFunc != nil {
		return f.listFunc(ctx, input)
	}
	return appeventlog.ListResult{}, nil
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

func newEventLogServer(handler *Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	router := NewTestRouter(handler)
	var wrapped = router
	for i := len(middleware) - 1; i >= 0; i-- {
		wrapped = middleware[i](wrapped)
	}
	return wrapped
}

func withActor(userID shared.UserID) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpsession.WithActor(r.Context(), httpsession.Actor{UserID: userID})))
		})
	}
}

func viewAuditLogAccess(t *testing.T) domainmembership.EffectivePermissions {
	t.Helper()

	perm, err := domainrole.NewPermission(domainrole.CanViewAuditLog, domainrole.ScopeCurrentAndDescendants)
	if err != nil {
		t.Fatalf("NewPermission(CanViewAuditLog): %v", err)
	}
	return domainmembership.Calculate(
		[]domainmembership.MembershipContext{{
			DivisionID:  "division-root",
			Permissions: domainrole.NewPermissionSet([]domainrole.Permission{perm}),
		}},
		"division-root",
		nil,
	)
}

func authWithViewAudit(t *testing.T) *fakeAuthorizationService {
	t.Helper()
	return &fakeAuthorizationService{
		resolveGlobalFunc: func(context.Context, shared.UserID) (domainmembership.EffectivePermissions, error) {
			return viewAuditLogAccess(t), nil
		},
	}
}

func eventFixture(t *testing.T) domaineventlog.Entry {
	t.Helper()

	payload, err := domaineventlog.NewPayload(map[string]string{
		"division_id": "division-1",
	})
	if err != nil {
		t.Fatalf("NewPayload() error = %v", err)
	}
	entry, err := domaineventlog.NewEntry(
		"actor-1",
		domaineventlog.EventDivisionArchived,
		domaineventlog.SubjectDivision,
		"division-1",
		payload,
		time.Date(2026, 4, 18, 8, 12, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewEntry() error = %v", err)
	}
	return entry
}

func errorCode(t *testing.T, body []byte) string {
	t.Helper()
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("json.Unmarshal(errorEnvelope) error = %v", err)
	}
	return envelope.Error.Code
}
