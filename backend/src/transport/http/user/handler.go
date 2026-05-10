package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	appauth "activist-base/src/application/auth"
	appuser "activist-base/src/application/user"
	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
	httpsession "activist-base/src/transport/http/session"

	"github.com/go-chi/chi/v5"
)

type Authenticator interface {
	Authenticate(next http.Handler) http.Handler
}

type ProfileRepository interface {
	GetUserByID(ctx context.Context, id shared.UserID) (domainuser.User, error)
	UpdateProfile(ctx context.Context, user domainuser.User) (domainuser.User, error)
}

type MembershipReader interface {
	ListByUser(ctx context.Context, userID shared.UserID) ([]appuser.MembershipView, error)
}

type AuthorizationService interface {
	ResolveForUser(ctx context.Context, actorID shared.UserID, targetUserID shared.UserID) (domainmembership.EffectivePermissions, error)
}

type Handler struct {
	repo        ProfileRepository
	memberships MembershipReader
	authz       AuthorizationService
}

func NewHandler(repo ProfileRepository, memberships MembershipReader, authz AuthorizationService) *Handler {
	return &Handler{
		repo:        repo,
		memberships: memberships,
		authz:       authz,
	}
}

func RegisterRoutes(r chi.Router, handler *Handler, authMiddleware Authenticator) {
	if r == nil || handler == nil {
		return
	}

	r.Route("/users", func(users chi.Router) {
		if authMiddleware != nil {
			users.Use(authMiddleware.Authenticate)
		}
		users.Get("/{user_id}", handler.GetProfile)
		users.Patch("/{user_id}", handler.PatchProfile)
		users.Get("/{user_id}/memberships", handler.GetMemberships)
	})
}

// NewTestRouter mounts user endpoints without auth middleware for handler tests.
func NewTestRouter(handler *Handler) http.Handler {
	r := chi.NewRouter()
	r.Route("/api/v1", func(api chi.Router) {
		RegisterRoutes(api, handler, nil)
	})
	return r
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, errEnvelopeFrom(appauth.ErrInvalidSession))
		return
	}

	targetUser, err := h.loadTargetUser(r.Context(), userIDFromRequest(r))
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}
	access, err := h.resolveAccess(r.Context(), actor.UserID, targetUser.ID)
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	readResult, err := appuser.ReadProfile(appuser.ReadProfileInput{
		ActorID:    actor.UserID,
		TargetUser: targetUser,
		Access:     access,
	})
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	writeJSON(w, http.StatusOK, newProfileResponse(readResult.User, readResult.Access))
}

func (h *Handler) PatchProfile(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, errEnvelopeFrom(appauth.ErrInvalidSession))
		return
	}

	targetID := userIDFromRequest(r)
	targetUser, err := h.loadTargetUser(r.Context(), targetID)
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}
	access, err := h.resolveAccess(r.Context(), actor.UserID, targetUser.ID)
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	patch, err := decodePatchProfileRequest(r)
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	updatedResult, err := appuser.EditProfile(appuser.EditProfileInput{
		ActorID:  actor.UserID,
		Existing: targetUser,
		Patch:    patch,
		Access:   access,
	})
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	persistedUser, err := h.repo.UpdateProfile(r.Context(), updatedResult.User)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, errEnvelopeFrom(notFoundUserErr()))
			return
		}
		writeError(w, errEnvelopeFrom(err))
		return
	}

	readResult, err := appuser.ReadProfile(appuser.ReadProfileInput{
		ActorID:    actor.UserID,
		TargetUser: persistedUser,
		Access:     access,
	})
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	writeJSON(w, http.StatusOK, newProfileResponse(readResult.User, readResult.Access))
}

func (h *Handler) GetMemberships(w http.ResponseWriter, r *http.Request) {
	actor, ok := httpsession.ActorFromRequest(r)
	if !ok {
		writeError(w, errEnvelopeFrom(appauth.ErrInvalidSession))
		return
	}

	targetUser, err := h.loadTargetUser(r.Context(), userIDFromRequest(r))
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}
	if _, err := h.resolveAccess(r.Context(), actor.UserID, targetUser.ID); err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	items, err := h.memberships.ListByUser(r.Context(), targetUser.ID)
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	membershipsResult, err := appuser.GetMemberships(appuser.GetMembershipsInput{
		ActorID:    actor.UserID,
		TargetUser: targetUser,
		Items:      items,
	})
	if err != nil {
		writeError(w, errEnvelopeFrom(err))
		return
	}

	writeJSON(w, http.StatusOK, newMembershipsResponse(membershipsResult.Items))
}

func (h *Handler) resolveAccess(ctx context.Context, actorID shared.UserID, targetUserID shared.UserID) (domainmembership.EffectivePermissions, error) {
	if h.authz == nil {
		return domainmembership.EffectivePermissions{}, shared.ErrForbidden
	}
	return h.authz.ResolveForUser(ctx, actorID, targetUserID)
}

func (h *Handler) loadTargetUser(ctx context.Context, targetID shared.UserID) (domainuser.User, error) {
	if targetID == "" {
		return domainuser.User{}, shared.ErrEmptyID
	}

	user, err := h.repo.GetUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domainuser.User{}, notFoundUserErr()
		}
		return domainuser.User{}, err
	}
	return user, nil
}

func userIDFromRequest(r *http.Request) shared.UserID {
	if r == nil {
		return ""
	}
	return shared.UserID(strings.TrimSpace(chi.URLParam(r, "user_id")))
}

func decodePatchProfileRequest(r *http.Request) (appuser.ProfilePatch, error) {
	var request patchProfileRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return appuser.ProfilePatch{}, &shared.Error{
			Code:    "validation.invalid_json",
			Message: "request body must be valid JSON",
		}
	}

	birthDate, err := parseBirthDate(request.BirthDate)
	if err != nil {
		return appuser.ProfilePatch{}, err
	}

	return appuser.ProfilePatch{
		FirstName:       request.FirstName,
		LastName:        request.LastName,
		MiddleName:      request.MiddleName,
		GradebookNumber: request.GradebookNumber,
		GroupNumber:     request.GroupNumber,
		Institute:       request.Institute,
		BirthDate:       birthDate,
		Phone:           request.Phone,
		SocialLinks:     request.SocialLinks,
		About:           request.About,
	}, nil
}

func parseBirthDate(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	parsed, err := time.Parse(dateLayout, strings.TrimSpace(*raw))
	if err != nil {
		return nil, &shared.Error{
			Code:    "validation.birth_date",
			Message: "birth_date must use YYYY-MM-DD",
		}
	}
	parsed = parsed.UTC()
	return &parsed, nil
}

func notFoundUserErr() error {
	return &shared.Error{
		Code:    "not_found.user",
		Message: "user not found",
	}
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

func writeError(w http.ResponseWriter, statusPayload errorEnvelopePayload) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusPayload.statusCode)
	_ = json.NewEncoder(w).Encode(errorEnvelope{
		Error: statusPayload.payload,
	})
}

type errorEnvelopePayload struct {
	statusCode int
	payload    errorPayload
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

	return errorEnvelopePayload{
		statusCode: statusCode,
		payload:    payload,
	}
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
