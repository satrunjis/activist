package eventlog

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	appauth "activist-base/src/application/auth"
	appeventlog "activist-base/src/application/eventlog"
	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/shared"
	httpsession "activist-base/src/transport/http/session"

	"github.com/go-chi/chi/v5"
)

type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

type Repository interface {
	List(ctx context.Context, input appeventlog.ListInput) (appeventlog.ListResult, error)
}

type AuthorizationService interface {
	ResolveGlobal(ctx context.Context, actorID shared.UserID) (domainmembership.EffectivePermissions, error)
}

type Handler struct {
	repo  Repository
	authz AuthorizationService
}

func NewHandler(repo Repository, authz AuthorizationService) *Handler {
	return &Handler{
		repo:  repo,
		authz: authz,
	}
}

func RegisterRoutes(r chi.Router, handler *Handler, authMiddleware Authenticator) {
	if r == nil || handler == nil {
		return
	}
	r.Route("/eventlog", func(router chi.Router) {
		if authMiddleware != nil {
			router.Use(authMiddleware.Authenticate)
		}
		router.Get("/", handler.List)
	})
}

func NewTestRouter(handler *Handler) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/v1", func(api chi.Router) {
		RegisterRoutes(api, handler, nil)
	})
	return r
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, errEnvelopeFrom(appauth.ErrInvalidSession))
		return
	}

	request, err := parseListRequest(r.URL.Query())
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	access, err := h.resolveGlobalAccess(r.Context(), actor.UserID)
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}
	if err := appeventlog.CheckListAdmissibility(access); err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	result, err := h.repo.List(r.Context(), request.toInput())
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	writeJSON(w, http.StatusOK, newListResponse(result))
}

func (h *Handler) resolveGlobalAccess(ctx context.Context, actorID shared.UserID) (domainmembership.EffectivePermissions, error) {
	if h.authz == nil {
		return domainmembership.EffectivePermissions{}, shared.ErrForbidden
	}
	return h.authz.ResolveGlobal(ctx, actorID)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type errorEnvelope struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type errorEnvelopePayload struct {
	statusCode int
	payload    errorPayload
}

func writeError(w http.ResponseWriter, statusPayload errorEnvelopePayload) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusPayload.statusCode)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: statusPayload.payload})
}

func errEnvelopeFrom(err error) errorEnvelopePayload {
	statusCode := http.StatusInternalServerError
	payload := errorPayload{
		Code:    "internal.error",
		Message: "internal server error",
	}

	var domainErr *shared.Error
	if errors.As(err, &domainErr) {
		statusCode, payload = payloadFromDomainError(*domainErr)
	}

	var fieldErr *shared.FieldError
	if errors.As(err, &fieldErr) {
		statusCode = http.StatusBadRequest
		payload.Code = "validation.field"
		payload.Message = "request validation failed"
		payload.Details = map[string]any{
			"field": fieldErr.Field,
			"kind":  fieldErr.Kind,
		}
		if fieldErr.Limit > 0 {
			payload.Details["limit"] = fieldErr.Limit
		}
	}

	return errorEnvelopePayload{statusCode: statusCode, payload: payload}
}

func payloadFromDomainError(err shared.Error) (int, errorPayload) {
	status := statusFromDomainCode(err.Code)
	payload := errorPayload{
		Code:    string(err.Code),
		Message: "request cannot be processed",
	}

	if status == http.StatusBadRequest {
		payload.Message = err.Message
		return status, payload
	}
	if status == http.StatusForbidden {
		payload.Message = "operation is forbidden"
		return status, payload
	}
	if status == http.StatusNotFound {
		payload.Message = "resource not found"
		return status, payload
	}
	if status == http.StatusUnauthorized {
		payload.Message = "unauthorized"
		return status, payload
	}
	return status, payload
}

func statusFromDomainCode(code shared.ErrorCode) int {
	switch {
	case strings.HasPrefix(string(code), "access."):
		return http.StatusForbidden
	case strings.HasPrefix(string(code), "auth."):
		return http.StatusUnauthorized
	case strings.HasPrefix(string(code), "not_found."):
		return http.StatusNotFound
	case strings.HasPrefix(string(code), "validation."):
		return http.StatusBadRequest
	case code == "shared.empty_id":
		return http.StatusBadRequest
	default:
		return http.StatusUnprocessableEntity
	}
}
