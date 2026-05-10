package position

import (
	"encoding/json"
	"net/http"
	"strings"

	appuser "activist-base/src/application/user"
	"activist-base/src/domain/shared"
)

type createPositionRequest struct {
	Title    string `json:"title"`
	RoleID   string `json:"role_id"`
	MaxCount *int   `json:"max_count,omitempty"`
}

type positionResponse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	RoleID     string `json:"role_id"`
	DivisionID string `json:"division_id"`
	MaxCount   *int   `json:"max_count,omitempty"`
	IsArchived bool   `json:"is_archived"`
}

type listPositionsResponse struct {
	Items []positionResponse `json:"items"`
}

type positionMembersResponse struct {
	Items []PositionMemberItem `json:"items"`
}

type PositionMemberItem struct {
	UserID       string `json:"user_id"`
	PositionID   string `json:"position_id"`
	PositionName string `json:"position_name"`
	RoleID       string `json:"role_id"`
	RoleName     string `json:"role_name"`
	DivisionID   string `json:"division_id"`
	DivisionName string `json:"division_name"`
}

func decodeCreatePositionRequest(r *http.Request) (createPositionRequest, error) {
	var request createPositionRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		return createPositionRequest{}, &shared.Error{
			Code:    "validation.invalid_json",
			Message: "request body must be valid JSON",
		}
	}
	request.Title = strings.TrimSpace(request.Title)
	request.RoleID = strings.TrimSpace(request.RoleID)
	return request, nil
}

func newPositionMembersResponse(items []appuser.MembershipView) positionMembersResponse {
	response := positionMembersResponse{
		Items: make([]PositionMemberItem, 0, len(items)),
	}
	for _, item := range items {
		response.Items = append(response.Items, PositionMemberItem{
			UserID:       string(item.Membership.UserID),
			PositionID:   string(item.Position.ID),
			PositionName: item.Position.Title,
			RoleID:       string(item.Position.RoleID),
			RoleName:     item.RoleName,
			DivisionID:   string(item.Division.ID),
			DivisionName: item.Division.ShortName,
		})
	}
	return response
}
