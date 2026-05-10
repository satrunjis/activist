package membership

import (
	"encoding/json"
	"net/http"
	"strings"

	"activist-base/src/domain/shared"
)

type assignMemberRequest struct {
	UserID     string `json:"user_id"`
	PositionID string `json:"position_id"`
}

type membershipResponse struct {
	UserID     string `json:"user_id"`
	PositionID string `json:"position_id"`
}

func decodeAssignMemberRequest(r *http.Request) (assignMemberRequest, error) {
	var request assignMemberRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		return assignMemberRequest{}, &shared.Error{
			Code:    "validation.invalid_json",
			Message: "request body must be valid JSON",
		}
	}
	request.UserID = strings.TrimSpace(request.UserID)
	request.PositionID = strings.TrimSpace(request.PositionID)
	return request, nil
}
