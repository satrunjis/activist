package membership

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	appauth "activist-base/src/application/auth"
	appeventlog "activist-base/src/application/eventlog"
	appmembership "activist-base/src/application/membership"
	domaineventlog "activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	"activist-base/src/domain/shared"
	httpsession "activist-base/src/transport/http/session"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

type Repository interface {
	GetPositionByID(ctx context.Context, id shared.PositionID) (domainposition.Position, error)
	AssignWithCountCheckWithTx(ctx context.Context, tx appeventlog.Transaction, userID shared.UserID, position domainposition.Position, access domainmembership.EffectivePermissions, actorID shared.UserID) (appmembership.AssignResult, error)
	RemoveWithTx(ctx context.Context, tx appeventlog.Transaction, userID shared.UserID, positionID shared.PositionID, access domainmembership.EffectivePermissions, actorID shared.UserID) (appmembership.RemoveResult, error)
}

type AuthorizationService interface {
	ResolveForDivision(ctx context.Context, actorID shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error)
}

type Handler struct {
	repo     Repository
	authz    AuthorizationService
	commands *appeventlog.CommandService
}

func NewHandler(repo Repository, authz AuthorizationService, commands ...*appeventlog.CommandService) *Handler {
	var commandService *appeventlog.CommandService
	if len(commands) > 0 {
		commandService = commands[0]
	}
	return &Handler{repo: repo, authz: authz, commands: commandService}
}

func RegisterRoutes(r chi.Router, handler *Handler, authMiddleware Authenticator) {
	if r == nil || handler == nil {
		return
	}

	if authMiddleware != nil {
		r.With(authMiddleware.Authenticate).Post("/memberships", handler.Assign)
		r.With(authMiddleware.Authenticate).Delete("/positions/{position_id}/members/{user_id}", handler.Remove)
		return
	}

	r.Post("/memberships", handler.Assign)
	r.Delete("/positions/{position_id}/members/{user_id}", handler.Remove)
}

func NewTestRouter(handler *Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Route("/api/v1", func(api chi.Router) {
		RegisterRoutes(api, handler, nil)
	})
	return r
}

func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}

	request, err := decodeAssignMemberRequest(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	userID := shared.UserID(request.UserID)
	positionID := shared.PositionID(request.PositionID)
	if userID == "" || positionID == "" {
		writeError(w, r, shared.ErrEmptyID)
		return
	}

	position, err := h.repo.GetPositionByID(r.Context(), positionID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	access, err := h.resolveAccess(r.Context(), actor.UserID, position.DivisionID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	if h.commands == nil {
		writeError(w, r, errors.New("membership command service is not configured"))
		return
	}

	var assigned domainmembership.Membership
	err = h.commands.ExecuteWithEvent(r.Context(), func(ctx context.Context, tx appeventlog.Transaction) (domaineventlog.Entry, error) {
		assignResult, txErr := h.repo.AssignWithCountCheckWithTx(ctx, tx, userID, position, access, actor.UserID)
		if txErr != nil {
			return domaineventlog.Entry{}, txErr
		}
		assigned = assignResult.Membership
		return assignResult.Event, nil
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, membershipResponse{
		UserID:     string(assigned.UserID),
		PositionID: string(assigned.PositionID),
	})
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}
	positionID := shared.PositionID(strings.TrimSpace(chi.URLParam(r, "position_id")))
	userID := shared.UserID(strings.TrimSpace(chi.URLParam(r, "user_id")))
	if positionID == "" || userID == "" {
		writeError(w, r, shared.ErrEmptyID)
		return
	}
	position, err := h.repo.GetPositionByID(r.Context(), positionID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	access, err := h.resolveAccess(r.Context(), actor.UserID, position.DivisionID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if h.commands == nil {
		writeError(w, r, errors.New("membership command service is not configured"))
		return
	}

	err = h.commands.ExecuteWithEvent(r.Context(), func(ctx context.Context, tx appeventlog.Transaction) (domaineventlog.Entry, error) {
		removeResult, txErr := h.repo.RemoveWithTx(ctx, tx, userID, positionID, access, actor.UserID)
		if txErr != nil {
			return domaineventlog.Entry{}, txErr
		}
		return removeResult.Event, nil
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) resolveAccess(ctx context.Context, actorID shared.UserID, divisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
	if h.authz == nil {
		return domainmembership.EffectivePermissions{}, shared.ErrForbidden
	}
	return h.authz.ResolveForDivision(ctx, actorID, divisionID)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type errorEnvelope struct {
	Error     errorPayload `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
}

type errorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		err = &shared.Error{Code: "not_found.resource", Message: "resource not found"}
	}

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

	writeErrorEnvelope(w, r, statusCode, payload)
}

func writeErrorEnvelope(w http.ResponseWriter, r *http.Request, statusCode int, payload errorPayload) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorEnvelope{
		Error:     payload,
		RequestID: chiMiddleware.GetReqID(r.Context()),
	})
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
	case code == "membership.already_exists":
		return http.StatusConflict
	case code == "membership.not_found":
		return http.StatusNotFound
	default:
		return http.StatusUnprocessableEntity
	}
}
