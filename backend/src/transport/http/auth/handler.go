package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	appauth "activist-base/src/application/auth"
	domainmembership "activist-base/src/domain/membership"
	domainrole "activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	httpsession "activist-base/src/transport/http/session"

	"github.com/go-chi/chi/v5"
)

const defaultSessionCookieName = "__Host-session"

type Service interface {
	Register(ctx context.Context, input appauth.RegisterInput) (appauth.AuthResult, error)
	Login(ctx context.Context, input appauth.LoginInput) (appauth.AuthResult, error)
	Logout(ctx context.Context, input appauth.LogoutInput) error
	LogoutAll(ctx context.Context, input appauth.LogoutAllInput) error
}

type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

type AuthorizationService interface {
	ResolveGlobal(ctx context.Context, actorID shared.UserID) (domainmembership.EffectivePermissions, error)
}

type Handler struct {
	service    Service
	cookieName string
	authz      AuthorizationService
}

func NewHandler(service Service, cookieName string) *Handler {
	if cookieName == "" {
		cookieName = defaultSessionCookieName
	}
	return &Handler{
		service:    service,
		cookieName: cookieName,
	}
}

func (h *Handler) SetAuthorizationService(authz AuthorizationService) {
	if h == nil {
		return
	}
	h.authz = authz
}

func RegisterRoutes(r chi.Router, handler *Handler, authMiddleware Authenticator) {
	if r == nil || handler == nil {
		return
	}

	r.Route("/auth", func(authRouter chi.Router) {
		authRouter.Post("/register", handler.Register)
		authRouter.Post("/login", handler.Login)

		if authMiddleware != nil {
			authRouter.Group(func(protected chi.Router) {
				protected.Use(authMiddleware.Authenticate)
				protected.Get("/session", handler.Session)
				protected.Post("/logout", handler.Logout)
				protected.Post("/logout-all", handler.LogoutAll)
			})
		}
	})
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, &shared.Error{
			Code:    "validation.invalid_json",
			Message: "request body must be valid JSON",
		})
		return
	}

	birthDate, err := time.Parse(dateLayout, request.BirthDate)
	if err != nil {
		writeError(w, &shared.Error{
			Code:    "validation.birth_date",
			Message: "birth_date must use YYYY-MM-DD",
		})
		return
	}

	authResult, err := h.service.Register(r.Context(), appauth.RegisterInput{
		Login:           request.Login,
		Password:        request.Password,
		FirstName:       request.FirstName,
		GradebookNumber: request.GradebookNumber,
		GroupNumber:     request.GroupNumber,
		Institute:       request.Institute,
		BirthDate:       birthDate,
		LastName:        request.LastName,
		MiddleName:      request.MiddleName,
		Phone:           request.Phone,
		About:           request.About,
		IP:              remoteIP(r),
		UserAgent:       userAgent(r),
	})
	if err != nil {
		writeError(w, err)
		return
	}

	h.setSessionCookie(w, authResult.Session.Token, authResult.Session.ExpiresAt)
	writeJSON(w, http.StatusCreated, authResponse{
		User:    newUserResponse(authResult.User),
		Session: newSessionResponse(authResult.Session.ExpiresAt, authResult.Session.IdleExpiresAt),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeError(w, &shared.Error{
			Code:    "validation.invalid_json",
			Message: "request body must be valid JSON",
		})
		return
	}

	authResult, err := h.service.Login(r.Context(), appauth.LoginInput{
		Login:     request.Login,
		Password:  request.Password,
		IP:        remoteIP(r),
		UserAgent: userAgent(r),
	})
	if err != nil {
		writeError(w, err)
		return
	}

	h.setSessionCookie(w, authResult.Session.Token, authResult.Session.ExpiresAt)
	writeJSON(w, http.StatusOK, authResponse{
		User:    newUserResponse(authResult.User),
		Session: newSessionResponse(authResult.Session.ExpiresAt, authResult.Session.IdleExpiresAt),
	})
}

func (h *Handler) Session(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, appauth.ErrInvalidSession)
		return
	}

	permissions := []string{}
	if h.authz != nil {
		access, err := h.authz.ResolveGlobal(r.Context(), actor.UserID)
		if err != nil {
			writeError(w, err)
			return
		}
		permissions = permissionCodesFromAccess(access)
	}

	writeJSON(w, http.StatusOK, sessionInspectResponse{
		User:        newUserResponse(actor.User),
		Session:     newSessionResponse(actor.SessionExpiresAt, actor.SessionIdleExpiresAt),
		CSRFToken:   actor.CSRFToken,
		Permissions: permissions,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, appauth.ErrInvalidSession)
		return
	}

	if err := h.service.Logout(r.Context(), appauth.LogoutInput{
		Token: actor.SessionToken,
	}); err != nil {
		writeError(w, err)
		return
	}

	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, appauth.ErrInvalidSession)
		return
	}

	if err := h.service.LogoutAll(r.Context(), appauth.LogoutAllInput{
		UserID: actor.UserID,
	}); err != nil {
		writeError(w, err)
		return
	}

	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt.UTC(),
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

func userAgent(r *http.Request) string {
	return strings.TrimSpace(r.UserAgent())
}

func permissionCodesFromAccess(access domainmembership.EffectivePermissions) []string {
	candidates := []domainrole.PermissionCode{
		domainrole.CanAddMember,
		domainrole.CanRemoveMember,
		domainrole.CanAssignPosition,
		domainrole.CanEditDivision,
		domainrole.CanManagePositions,
		domainrole.CanCreateSubdivision,
		domainrole.CanArchiveDivision,
		domainrole.CanViewContacts,
		domainrole.CanEditSelfProfile,
		domainrole.CanManageRoles,
		domainrole.CanViewAuditLog,
		domainrole.SystemAdmin,
	}
	permissions := make([]string, 0, len(candidates))
	for _, code := range candidates {
		if access.Has(code, false) {
			permissions = append(permissions, code.String())
		}
	}
	return permissions
}

type errorEnvelope struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func writeError(w http.ResponseWriter, err error) {
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Error: payload})
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
