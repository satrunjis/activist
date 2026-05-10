package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"activist-base/src/config"
)

func TestNewRouter_AllowsCORSPreflightForAuthRegister(t *testing.T) {
	router := NewRouter(Deps{
		Config: config.Config{
			SessionCookieName: "__Host-session",
		},
	})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/register", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "content-type")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected Access-Control-Allow-Origin to be localhost:5173, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatalf("expected Access-Control-Allow-Methods header to be set")
	}
}
