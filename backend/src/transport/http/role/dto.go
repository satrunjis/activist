package role

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"activist-base/src/domain/shared"
)

type createRoleRequest struct {
	Name        string              `json:"name"`
	Permissions []permissionRequest `json:"permissions"`
}

type editRoleRequest struct {
	Name        *string             `json:"name"`
	Permissions []permissionRequest `json:"permissions"`
}

type roleResponse struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Permissions []rolePermissionView `json:"permissions,omitempty"`
}

type rolePermissionView struct {
	Code  string `json:"code"`
	Scope string `json:"scope"`
}

type permissionRequest struct {
	Code  string `json:"code"`
	Scope string `json:"scope,omitempty"`
}

// UnmarshalJSON supports both the legacy string format ("can_add_member") and the
// current object format ({"code": "can_add_member", "scope": "current_division"}).
func (p *permissionRequest) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var code string
		if err := json.Unmarshal(data, &code); err != nil {
			return err
		}
		*p = permissionRequest{Code: code}
		return nil
	}
	type alias permissionRequest
	var payload alias
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	*p = permissionRequest(payload)
	return nil
}

type listRolesResponse struct {
	Items  []roleResponse `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type listRolesRequest struct {
	Limit  int
	Offset int
}

const (
	defaultRolesLimit = 20
	maxRolesLimit     = 100
)

func decodeCreateRoleRequest(r *http.Request) (createRoleRequest, error) {
	var request createRoleRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		return createRoleRequest{}, &shared.Error{
			Code:    "validation.invalid_json",
			Message: "request body must be valid JSON",
		}
	}
	request.Name = strings.TrimSpace(request.Name)
	return request, nil
}

func decodeEditRoleRequest(r *http.Request) (editRoleRequest, error) {
	var request editRoleRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		return editRoleRequest{}, &shared.Error{
			Code:    "validation.invalid_json",
			Message: "request body must be valid JSON",
		}
	}
	if request.Name != nil {
		trimmed := strings.TrimSpace(*request.Name)
		request.Name = &trimmed
	}
	return request, nil
}

func parseListRolesRequest(query url.Values) (listRolesRequest, error) {
	request := listRolesRequest{
		Limit:  defaultRolesLimit,
		Offset: 0,
	}

	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 || value > maxRolesLimit {
			return listRolesRequest{}, &shared.Error{
				Code:    "validation.limit",
				Message: "limit must be between 1 and 100",
			}
		}
		request.Limit = value
	}

	if raw := strings.TrimSpace(query.Get("offset")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			return listRolesRequest{}, &shared.Error{
				Code:    "validation.offset",
				Message: "offset must be zero or positive",
			}
		}
		request.Offset = value
	}

	return request, nil
}
