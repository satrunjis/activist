package division

import (
	"encoding/json"
	"net/http"
	"strings"

	appdivision "activist-base/src/application/division"
	"activist-base/src/domain/shared"
	postgres "activist-base/src/repository/postgres"
)

type createDivisionRequest struct {
	ParentID      *string       `json:"parent_id,omitempty"`
	ShortName     string        `json:"short_name"`
	FullName      string        `json:"full_name"`
	Description   string        `json:"description"`
	RegulationURL string        `json:"regulation_url"`
	MediaLinks    []shared.Link `json:"media_links"`
}

type patchDivisionRequest struct {
	ParentID      *string        `json:"parent_id"`
	ShortName     *string        `json:"short_name"`
	FullName      *string        `json:"full_name"`
	Description   *string        `json:"description"`
	RegulationURL *string        `json:"regulation_url"`
	MediaLinks    *[]shared.Link `json:"media_links"`
}

type treeResponse struct {
	ID             string         `json:"id"`
	ParentID       *string        `json:"parent_id,omitempty"`
	ShortName      string         `json:"short_name"`
	IsArchived     bool           `json:"is_archived"`
	HasChildren    bool           `json:"has_children"`
	ChildrenCount  int            `json:"children_count"`
	PositionsCount int            `json:"positions_count"`
	MembersCount   int            `json:"members_count"`
	Children       []treeResponse `json:"children,omitempty"`
	FullName       *string        `json:"full_name,omitempty"`
	Description    *string        `json:"description,omitempty"`
	RegulationURL  *string        `json:"regulation_url,omitempty"`
	MediaLinks     []shared.Link  `json:"media_links,omitempty"`
}

func decodeCreateRequest(r *http.Request) (createDivisionRequest, error) {
	var request createDivisionRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		return createDivisionRequest{}, &shared.Error{
			Code:    "validation.invalid_json",
			Message: "request body must be valid JSON",
		}
	}
	return request, nil
}

func decodePatchRequest(r *http.Request) (patchDivisionRequest, error) {
	var request patchDivisionRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		return patchDivisionRequest{}, &shared.Error{
			Code:    "validation.invalid_json",
			Message: "request body must be valid JSON",
		}
	}
	return request, nil
}

func toDivisionPatch(request patchDivisionRequest) appdivision.EditPatch {
	return appdivision.EditPatch{
		ShortName:     trimStringPtr(request.ShortName),
		FullName:      trimStringPtr(request.FullName),
		Description:   trimStringPtr(request.Description),
		RegulationURL: trimStringPtr(request.RegulationURL),
		MediaLinks:    request.MediaLinks,
	}
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	return &trimmed
}

func toDivisionResponse(result appdivision.TreeResult) treeResponse {
	response := treeResponse{
		ID:             string(result.ID),
		ShortName:      result.ShortName,
		IsArchived:     result.IsArchived,
		HasChildren:    result.HasChildren,
		ChildrenCount:  result.ChildrenCount,
		PositionsCount: result.PositionsCount,
		MembersCount:   result.MembersCount,
		FullName:       result.FullName,
		Description:    result.Description,
		RegulationURL:  result.RegulationURL,
		MediaLinks:     result.MediaLinks,
	}
	if result.ParentID != nil {
		parentID := string(*result.ParentID)
		response.ParentID = &parentID
	}
	if len(result.Children) > 0 {
		response.Children = make([]treeResponse, 0, len(result.Children))
		for _, child := range result.Children {
			response.Children = append(response.Children, toDivisionResponse(child))
		}
	}
	return response
}

// childItemDTO is the flat per-node response for GET /api/v1/divisions?parent_id=.
type childItemDTO struct {
	ID            string        `json:"id"`
	ParentID      *string       `json:"parent_id,omitempty"`
	ShortName     string        `json:"short_name"`
	IsArchived    bool          `json:"is_archived"`
	HasChildren   bool          `json:"has_children"`
	ChildrenCount int           `json:"children_count"`
	FullName      *string       `json:"full_name,omitempty"`
	Description   *string       `json:"description,omitempty"`
	RegulationURL *string       `json:"regulation_url,omitempty"`
	MediaLinks    []shared.Link `json:"media_links,omitempty"`
}

func toChildItemDTO(item postgres.DivisionChildItem, canViewCard bool) childItemDTO {
	d := item.Division
	dto := childItemDTO{
		ID:            string(d.ID),
		ShortName:     d.ShortName,
		IsArchived:    d.IsArchived,
		HasChildren:   item.HasChildren,
		ChildrenCount: item.ChildrenCount,
	}
	if d.ParentID != nil {
		s := string(*d.ParentID)
		dto.ParentID = &s
	}
	if canViewCard {
		if d.FullName != "" {
			dto.FullName = &d.FullName
		}
		if d.Description != "" {
			dto.Description = &d.Description
		}
		if d.RegulationURL != "" {
			dto.RegulationURL = &d.RegulationURL
		}
		if len(d.MediaLinks) > 0 {
			dto.MediaLinks = d.MediaLinks
		}
	}
	return dto
}
