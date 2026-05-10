package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"activist-base/src/domain/shared"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type errorEnvelope struct {
	Error     errorPayload `json:"error"`
	RequestID string       `json:"request_id,omitempty"`
}

type errorPayload struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		WriteInternalError(w, r)
		return
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

func WriteInternalError(w http.ResponseWriter, r *http.Request) {
	writeErrorEnvelope(w, r, http.StatusInternalServerError, errorPayload{
		Code:    "internal.error",
		Message: "internal server error",
	})
}

func writeErrorEnvelope(w http.ResponseWriter, r *http.Request, statusCode int, payload errorPayload) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(errorEnvelope{
		Error:     payload,
		RequestID: chiMiddleware.GetReqID(r.Context()),
	})
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

func payloadFromDomainError(err shared.Error) (int, errorPayload) {
	status := statusFromDomainCode(err.Code)
	payload := errorPayload{
		Code:    string(err.Code),
		Message: "request cannot be processed",
	}

	// Validation errors are intentionally user-facing.
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
