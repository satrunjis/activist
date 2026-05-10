package position

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	appauth "activist-base/src/application/auth"
	appeventlog "activist-base/src/application/eventlog"
	appposition "activist-base/src/application/position"
	appuser "activist-base/src/application/user"
	domaindivision "activist-base/src/domain/division"
	domaineventlog "activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	httpsession "activist-base/src/transport/http/session"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

type Repository interface {
	GetDivisionByID(ctx context.Context, id shared.DivisionID) (domaindivision.Division, error)
	GetRoleByID(ctx context.Context, id shared.RoleID) (domainrole.Role, error)
	CreateWithTx(ctx context.Context, tx appeventlog.Transaction, position domainposition.Position) (domainposition.Position, error)
	ListByDivision(ctx context.Context, divisionID shared.DivisionID) ([]domainposition.Position, error)
	GetByID(ctx context.Context, id shared.PositionID) (domainposition.Position, error)
	ListMembersByPosition(ctx context.Context, id shared.PositionID) ([]appuser.MembershipView, error)
	ArchiveWithTx(ctx context.Context, tx appeventlog.Transaction, id shared.PositionID) (domainposition.Position, int, error)
}

type AuthorizationService interface {
	ResolveForDivision(ctx context.Context, actorID shared.UserID, targetDivisionID shared.DivisionID) (domainmembership.EffectivePermissions, error)
}

type Handler struct {
	repo     Repository
	authz    AuthorizationService
	commands *appeventlog.CommandService
	nowFn    func() time.Time
}

func NewHandler(repo Repository, authz AuthorizationService, commands ...*appeventlog.CommandService) *Handler {
	var commandService *appeventlog.CommandService
	if len(commands) > 0 {
		commandService = commands[0]
	}
	return &Handler{
		repo:     repo,
		authz:    authz,
		commands: commandService,
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func RegisterRoutes(r chi.Router, handler *Handler, authMiddleware Authenticator) {
	if r == nil || handler == nil {
		return
	}

	if authMiddleware != nil {
		r.With(authMiddleware.Authenticate).Route("/divisions/{division_id}/positions", func(positions chi.Router) {
			positions.Post("/", handler.Create)
			positions.Get("/", handler.ListByDivision)
		})
		r.With(authMiddleware.Authenticate).Get("/positions/{position_id}/members", handler.ListMembersByPosition)
		r.With(authMiddleware.Authenticate).Post("/positions/{position_id}/archive", handler.Archive)
		return
	}

	r.Route("/divisions/{division_id}/positions", func(positions chi.Router) {
		positions.Post("/", handler.Create)
		positions.Get("/", handler.ListByDivision)
	})
	r.Get("/positions/{position_id}/members", handler.ListMembersByPosition)
	r.Post("/positions/{position_id}/archive", handler.Archive)
}

func NewTestRouter(handler *Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware.RequestID)
	r.Route("/api/v1", func(api chi.Router) {
		RegisterRoutes(api, handler, nil)
	})
	return r
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}

	divisionID := shared.DivisionID(strings.TrimSpace(chi.URLParam(r, "division_id")))
	if divisionID == "" {
		writeError(w, r, shared.ErrEmptyID)
		return
	}
	request, err := decodeCreatePositionRequest(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	division, err := h.repo.GetDivisionByID(r.Context(), divisionID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	roleEntity, err := h.repo.GetRoleByID(r.Context(), shared.RoleID(request.RoleID))
	if err != nil {
		writeError(w, r, err)
		return
	}

	access, err := h.resolveDivisionAccess(r.Context(), actor.UserID, divisionID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	result, err := appposition.Create(appposition.CreateInput{
		ActorID:  actor.UserID,
		ID:       shared.PositionID(newPositionID()),
		Division: division,
		Role:     roleEntity,
		Title:    request.Title,
		MaxCount: request.MaxCount,
		Access:   access,
		Now:      h.nowFn(),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	if h.commands == nil {
		writeError(w, r, errors.New("position command service is not configured"))
		return
	}

	var persisted domainposition.Position
	err = h.commands.ExecuteWithEvent(r.Context(), func(ctx context.Context, tx appeventlog.Transaction) (domaineventlog.Entry, error) {
		var txErr error
		persisted, txErr = h.repo.CreateWithTx(ctx, tx, result.Position)
		if txErr != nil {
			return domaineventlog.Entry{}, txErr
		}
		return result.Event, nil
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, toPositionResponse(persisted))
}

func (h *Handler) ListByDivision(w http.ResponseWriter, r *http.Request) {
	_, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}
	divisionID := shared.DivisionID(strings.TrimSpace(chi.URLParam(r, "division_id")))
	if divisionID == "" {
		writeError(w, r, shared.ErrEmptyID)
		return
	}

	items, err := h.repo.ListByDivision(r.Context(), divisionID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	response := listPositionsResponse{Items: make([]positionResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, toPositionResponse(item))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) ListMembersByPosition(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}
	positionID := shared.PositionID(strings.TrimSpace(chi.URLParam(r, "position_id")))
	if positionID == "" {
		writeError(w, r, shared.ErrEmptyID)
		return
	}

	position, err := h.repo.GetByID(r.Context(), positionID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	access, err := h.resolveDivisionAccess(r.Context(), actor.UserID, position.DivisionID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if !access.Has(domainrole.CanManagePositions, false) {
		writeError(w, r, shared.ErrForbidden)
		return
	}

	items, err := h.repo.ListMembersByPosition(r.Context(), positionID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newPositionMembersResponse(items))
}

func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}
	positionID := shared.PositionID(strings.TrimSpace(chi.URLParam(r, "position_id")))
	if positionID == "" {
		writeError(w, r, shared.ErrEmptyID)
		return
	}
	existing, err := h.repo.GetByID(r.Context(), positionID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	access, err := h.resolveDivisionAccess(r.Context(), actor.UserID, existing.DivisionID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if _, err := appposition.Archive(appposition.ArchiveInput{
		ActorID:                 actor.UserID,
		Position:                existing,
		RemovedMembershipsCount: 0,
		Access:                  access,
		Now:                     h.nowFn(),
	}); err != nil {
		writeError(w, r, err)
		return
	}

	if h.commands == nil {
		writeError(w, r, errors.New("position command service is not configured"))
		return
	}

	var persisted domainposition.Position
	err = h.commands.ExecuteWithEvent(r.Context(), func(ctx context.Context, tx appeventlog.Transaction) (domaineventlog.Entry, error) {
		archived, removedCount, txErr := h.repo.ArchiveWithTx(ctx, tx, positionID)
		if txErr != nil {
			return domaineventlog.Entry{}, txErr
		}
		result, txErr := appposition.Archive(appposition.ArchiveInput{
			ActorID:                 actor.UserID,
			Position:                existing,
			RemovedMembershipsCount: removedCount,
			Access:                  access,
			Now:                     h.nowFn(),
		})
		if txErr != nil {
			return domaineventlog.Entry{}, txErr
		}
		persisted = archived
		return result.Event, nil
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, toPositionResponse(persisted))
}

func (h *Handler) resolveDivisionAccess(ctx context.Context, actorID shared.UserID, divisionID shared.DivisionID) (domainmembership.EffectivePermissions, error) {
	if h.authz == nil {
		return domainmembership.EffectivePermissions{}, shared.ErrForbidden
	}
	return h.authz.ResolveForDivision(ctx, actorID, divisionID)
}

func toPositionResponse(position domainposition.Position) positionResponse {
	return positionResponse{
		ID:         string(position.ID),
		Title:      position.Title,
		RoleID:     string(position.RoleID),
		DivisionID: string(position.DivisionID),
		MaxCount:   position.MaxCount,
		IsArchived: position.IsArchived,
	}
}

func newPositionID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
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
		err = &shared.Error{Code: "not_found.position", Message: "position not found"}
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
