package search_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appsearch "activist-base/src/application/search"
	"activist-base/src/domain/shared"
	httpsearch "activist-base/src/transport/http/search"
	httpsession "activist-base/src/transport/http/session"
)

func TestGetUsers(t *testing.T) {
	t.Parallel()

	t.Run("authenticated request with filters returns 200 and forwards parsed values", func(t *testing.T) {
		t.Parallel()

		service := &fakeSearchService{
			result: appsearch.SearchUsersResult{
				Items: []appsearch.SearchUser{
					{
						ID:        "target-1",
						FirstName: "Alice",
					},
				},
				Total: 1,
			},
		}
		server := newSearchServer(service, withActor("actor-1"))

		req := httptest.NewRequest(
			http.MethodGet,
			"/api/v1/search/users?first_name=Alice&position_title=Lead&role_name=Coordinator&limit=10&offset=2&include_archived=true",
			nil,
		)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if service.lastActor != "actor-1" {
			t.Fatalf("expected actor forwarded, got %q", service.lastActor)
		}
		if service.lastInput.FirstName != "Alice" {
			t.Fatalf("expected first_name forwarded, got %q", service.lastInput.FirstName)
		}
		if service.lastInput.PositionTitle != "Lead" {
			t.Fatalf("expected position_title forwarded, got %q", service.lastInput.PositionTitle)
		}
		if service.lastInput.RoleName != "Coordinator" {
			t.Fatalf("expected role_name forwarded, got %q", service.lastInput.RoleName)
		}
		if !service.lastInput.IncludeArchived {
			t.Fatal("expected include_archived=true to be forwarded")
		}
		if service.lastInput.Limit != 10 {
			t.Fatalf("expected limit=10, got %d", service.lastInput.Limit)
		}
		if service.lastInput.Offset != 2 {
			t.Fatalf("expected offset=2, got %d", service.lastInput.Offset)
		}
	})

	t.Run("missing session returns 401", func(t *testing.T) {
		t.Parallel()

		server := newSearchServer(&fakeSearchService{})
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", rec.Code)
		}
	})

	t.Run("invalid limit returns 400 with validation.limit", func(t *testing.T) {
		t.Parallel()

		server := newSearchServer(&fakeSearchService{}, withActor("actor-1"))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users?limit=0", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		errorPayload, _ := payload["error"].(map[string]any)
		if code, _ := errorPayload["code"].(string); code != "validation.limit" {
			t.Fatalf("expected validation.limit, got %q", code)
		}
	})

	t.Run("invalid include_archived returns 400 with validation.include_archived", func(t *testing.T) {
		t.Parallel()

		server := newSearchServer(&fakeSearchService{}, withActor("actor-1"))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users?include_archived=not_bool", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		errorPayload, _ := payload["error"].(map[string]any)
		if code, _ := errorPayload["code"].(string); code != "validation.include_archived" {
			t.Fatalf("expected validation.include_archived, got %q", code)
		}
	})

	t.Run("invalid offset returns 400 with validation.offset", func(t *testing.T) {
		t.Parallel()

		server := newSearchServer(&fakeSearchService{}, withActor("actor-1"))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users?offset=abc", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		errorPayload, _ := payload["error"].(map[string]any)
		if code, _ := errorPayload["code"].(string); code != "validation.offset" {
			t.Fatalf("expected validation.offset, got %q", code)
		}
	})

	t.Run("negative offset returns 400 with validation.offset", func(t *testing.T) {
		t.Parallel()

		server := newSearchServer(&fakeSearchService{}, withActor("actor-1"))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users?offset=-1", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		errorPayload, _ := payload["error"].(map[string]any)
		if code, _ := errorPayload["code"].(string); code != "validation.offset" {
			t.Fatalf("expected validation.offset, got %q", code)
		}
	})

	t.Run("response omits phone when service returns masked field", func(t *testing.T) {
		t.Parallel()

		service := &fakeSearchService{
			result: appsearch.SearchUsersResult{
				Items: []appsearch.SearchUser{
					{
						ID:        "target-1",
						FirstName: "Alice",
						Phone:     nil,
					},
				},
				Total: 1,
			},
		}
		server := newSearchServer(service, withActor("actor-1"))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}

		var payload map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		items, _ := payload["items"].([]any)
		if len(items) != 1 {
			t.Fatalf("expected one item, got %d", len(items))
		}
		firstItem, _ := items[0].(map[string]any)
		if _, ok := firstItem["phone"]; ok {
			t.Fatal("phone must be omitted when nil")
		}
	})

	t.Run("default include_archived=false", func(t *testing.T) {
		t.Parallel()

		service := &fakeSearchService{}
		server := newSearchServer(service, withActor("actor-1"))
		req := httptest.NewRequest(http.MethodGet, "/api/v1/search/users", nil)
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if service.lastInput.IncludeArchived {
			t.Fatal("expected include_archived=false by default")
		}
	})
}

type fakeSearchService struct {
	lastActor shared.UserID
	lastInput appsearch.SearchUsersInput
	result    appsearch.SearchUsersResult
	err       error
}

func (f *fakeSearchService) SearchUsers(
	_ context.Context,
	actorID shared.UserID,
	input appsearch.SearchUsersInput,
) (appsearch.SearchUsersResult, error) {
	f.lastActor = actorID
	f.lastInput = input
	if f.err != nil {
		return appsearch.SearchUsersResult{}, f.err
	}
	return f.result, nil
}

func newSearchServer(service httpsearch.Service, middleware ...func(http.Handler) http.Handler) http.Handler {
	handler := httpsearch.NewHandler(service)
	router := httpsearch.NewTestRouter(handler)

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
