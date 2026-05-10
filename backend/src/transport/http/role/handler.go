package role

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
	approle "activist-base/src/application/role"
	domaineventlog "activist-base/src/domain/eventlog"
	domainmembership "activist-base/src/domain/membership"
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
	CreateWithTx(ctx context.Context, tx appeventlog.Transaction, role domainrole.Role) (domainrole.Role, error)
	GetByID(ctx context.Context, id shared.RoleID) (domainrole.Role, error)
	UpdateWithTx(ctx context.Context, tx appeventlog.Transaction, role domainrole.Role) (domainrole.Role, error)
	Delete(ctx context.Context, id shared.RoleID) error
	List(ctx context.Context, limit, offset int) ([]domainrole.Role, int64, error)
}

type AuthorizationService interface {
	ResolveGlobal(ctx context.Context, actorID shared.UserID) (domainmembership.EffectivePermissions, error)
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
	r.Route("/roles", func(roles chi.Router) {
		if authMiddleware != nil {
			roles.Use(authMiddleware.Authenticate)
		}
		roles.Post("/", handler.Create)
		roles.Get("/", handler.List)
		roles.Patch("/{roleID}", handler.Edit)
		roles.Delete("/{roleID}", handler.Delete)
	})
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

	request, err := decodeCreateRoleRequest(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	permissions, err := permissionsFromRequest(request.Permissions)
	if err != nil {
		writeError(w, r, err)
		return
	}
	access, err := h.resolveGlobalAccess(r.Context(), actor.UserID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	result, err := approle.Create(approle.CreateInput{
		ActorID:     actor.UserID,
		ID:          shared.RoleID(newRoleID()),
		Name:        request.Name,
		Permissions: permissions,
		Access:      access,
		Now:         h.nowFn(),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	if h.commands == nil {
		writeError(w, r, errors.New("role command service is not configured"))
		return
	}

	var persisted domainrole.Role
	err = h.commands.ExecuteWithEvent(r.Context(), func(ctx context.Context, tx appeventlog.Transaction) (domaineventlog.Entry, error) {
		var txErr error
		persisted, txErr = h.repo.CreateWithTx(ctx, tx, result.Role)
		if txErr != nil {
			return domaineventlog.Entry{}, txErr
		}
		return result.Event, nil
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusCreated, roleResponse{
		ID:          string(persisted.ID),
		Name:        persisted.Name,
		Permissions: permissionViews(persisted.Permissions),
	})
}

func (h *Handler) resolveGlobalAccess(ctx context.Context, actorID shared.UserID) (domainmembership.EffectivePermissions, error) {
	if h.authz == nil {
		return domainmembership.EffectivePermissions{}, shared.ErrForbidden
	}
	return h.authz.ResolveGlobal(ctx, actorID)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	_, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}

	request, err := parseListRolesRequest(r.URL.Query())
	if err != nil {
		writeError(w, r, err)
		return
	}

	items, total, err := h.repo.List(r.Context(), request.Limit, request.Offset)
	if err != nil {
		writeError(w, r, err)
		return
	}

	response := listRolesResponse{
		Items:  make([]roleResponse, 0, len(items)),
		Total:  total,
		Limit:  request.Limit,
		Offset: request.Offset,
	}
	for _, item := range items {
		response.Items = append(response.Items, roleResponse{
			ID:          string(item.ID),
			Name:        item.Name,
			Permissions: permissionViews(item.Permissions),
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Edit(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}

	roleID := shared.RoleID(chi.URLParam(r, "roleID"))

	request, err := decodeEditRoleRequest(r)
	if err != nil {
		writeError(w, r, err)
		return
	}

	existing, err := h.repo.GetByID(r.Context(), roleID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	access, err := h.resolveGlobalAccess(r.Context(), actor.UserID)
	if err != nil {
		writeError(w, r, err)
		return
	}

	patch := approle.EditPatch{}
	if request.Name != nil {
		patch.Name = request.Name
	}
	if request.Permissions != nil {
		perms, err := permissionsFromRequest(request.Permissions)
		if err != nil {
			writeError(w, r, err)
			return
		}
		patch.Permissions = &perms
	}

	result, err := approle.Edit(approle.EditInput{
		ActorID:  actor.UserID,
		Existing: existing,
		Patch:    patch,
		Access:   access,
		Now:      h.nowFn(),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	if h.commands == nil {
		writeError(w, r, errors.New("role command service is not configured"))
		return
	}

	var persisted domainrole.Role
	err = h.commands.ExecuteWithEvent(r.Context(), func(ctx context.Context, tx appeventlog.Transaction) (domaineventlog.Entry, error) {
		var txErr error
		persisted, txErr = h.repo.UpdateWithTx(ctx, tx, result.Role)
		if txErr != nil {
			return domaineventlog.Entry{}, txErr
		}
		return result.Event, nil
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, roleResponse{
		ID:          string(persisted.ID),
		Name:        persisted.Name,
		Permissions: permissionViews(persisted.Permissions),
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, r, appauth.ErrInvalidSession)
		return
	}

	roleID := shared.RoleID(chi.URLParam(r, "roleID"))

	access, err := h.resolveGlobalAccess(r.Context(), actor.UserID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if err := approle.Delete(approle.DeleteInput{
		ActorID: actor.UserID,
		RoleID:  roleID,
		Access:  access,
	}); err != nil {
		writeError(w, r, err)
		return
	}

	if err := h.repo.Delete(r.Context(), roleID); err != nil {
		writeError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func newRoleID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}

func permissionsFromRequest(raw []permissionRequest) (domainrole.PermissionSet, error) {
	if len(raw) == 0 {
		return domainrole.PermissionSet{}, nil
	}

	perms := make([]domainrole.Permission, 0, len(raw))
	for _, entry := range raw {
		code, ok := permissionCodeFromString(entry.Code)
		if !ok {
			return domainrole.PermissionSet{}, &shared.Error{
				Code:    "validation.permissions",
				Message: "unknown permission code",
			}
		}
		scope, err := scopeFromRequest(code, entry.Scope)
		if err != nil {
			return domainrole.PermissionSet{}, err
		}
		perm, err := domainrole.NewPermission(code, scope)
		if err != nil {
			return domainrole.PermissionSet{}, err
		}
		perms = append(perms, perm)
	}
	return domainrole.NewPermissionSet(perms), nil
}

func scopeFromRequest(code domainrole.PermissionCode, raw string) (domainrole.ScopeMode, error) {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		// Backward compatibility for legacy payloads: ["can_add_member", ...]
		scope, ok := domainrole.CanonicalScopeForPermission(code)
		if !ok {
			return 0, &shared.Error{
				Code:    "validation.permissions",
				Message: "unknown permission scope",
			}
		}
		return scope, nil
	}
	switch value {
	case "self":
		return domainrole.ScopeSelf, nil
	case "current_division":
		return domainrole.ScopeCurrentDivision, nil
	case "current_and_descendants":
		return domainrole.ScopeCurrentAndDescendants, nil
	default:
		return 0, &shared.Error{
			Code:    "validation.permissions",
			Message: "unknown permission scope",
		}
	}
}

func permissionViews(set domainrole.PermissionSet) []rolePermissionView {
	perms := set.Permissions()
	if len(perms) == 0 {
		return nil
	}
	out := make([]rolePermissionView, 0, len(perms))
	for _, p := range perms {
		out = append(out, rolePermissionView{
			Code:  p.Code().String(),
			Scope: scopeToString(p.Scope()),
		})
	}
	return out
}

func scopeToString(scope domainrole.ScopeMode) string {
	switch scope {
	case domainrole.ScopeSelf:
		return "self"
	case domainrole.ScopeCurrentDivision:
		return "current_division"
	case domainrole.ScopeCurrentAndDescendants:
		return "current_and_descendants"
	default:
		return "unknown"
	}
}

func permissionCodeFromString(value string) (domainrole.PermissionCode, bool) {
	switch strings.TrimSpace(value) {
	case "can_add_member":
		return domainrole.CanAddMember, true
	case "can_remove_member":
		return domainrole.CanRemoveMember, true
	case "can_assign_position":
		return domainrole.CanAssignPosition, true
	case "can_edit_division":
		return domainrole.CanEditDivision, true
	case "can_manage_positions":
		return domainrole.CanManagePositions, true
	case "can_create_subdivision":
		return domainrole.CanCreateSubdivision, true
	case "can_archive_division":
		return domainrole.CanArchiveDivision, true
	case "can_view_contacts":
		return domainrole.CanViewContacts, true
	case "can_edit_self_profile":
		return domainrole.CanEditSelfProfile, true
	case "can_manage_roles":
		return domainrole.CanManageRoles, true
	case "can_view_audit_log":
		return domainrole.CanViewAuditLog, true
	case "system_admin":
		return domainrole.SystemAdmin, true
	default:
		return 0, false
	}
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
		err = &shared.Error{Code: "not_found.role", Message: "role not found"}
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
	case code == "validation.role_in_use":
		return http.StatusConflict
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
