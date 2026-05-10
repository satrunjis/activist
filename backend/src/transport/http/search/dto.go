package search

import (
	"net/url"
	"strconv"
	"strings"

	appsearch "activist-base/src/application/search"
	"activist-base/src/domain/shared"
)

type usersSearchRequest struct {
	FirstName       string
	LastName        string
	MiddleName      string
	Login           string
	GroupNumber     string
	Institute       string
	About           string
	PositionTitle   string
	RoleName        string
	IncludeArchived bool
	Limit           int32
	Offset          int32
}

type usersSearchResponse struct {
	Items []searchUserItem `json:"items"`
	Total int64            `json:"total"`
}

type searchUserItem struct {
	ID              string         `json:"id"`
	FirstName       string         `json:"first_name"`
	LastName        *string        `json:"last_name,omitempty"`
	MiddleName      *string        `json:"middle_name,omitempty"`
	Login           *string        `json:"login,omitempty"`
	GroupNumber     *string        `json:"group_number,omitempty"`
	Institute       *string        `json:"institute,omitempty"`
	Phone           *string        `json:"phone,omitempty"`
	SocialLinks     *[]shared.Link `json:"social_links,omitempty"`
	About           *string        `json:"about,omitempty"`
	BirthDate       *string        `json:"birth_date,omitempty"`
	GradebookNumber *string        `json:"gradebook_number,omitempty"`
}

func parseUsersSearchRequest(query url.Values) (usersSearchRequest, error) {
	request := usersSearchRequest{
		FirstName:     strings.TrimSpace(query.Get("first_name")),
		LastName:      strings.TrimSpace(query.Get("last_name")),
		MiddleName:    strings.TrimSpace(query.Get("middle_name")),
		Login:         strings.TrimSpace(query.Get("login")),
		GroupNumber:   strings.TrimSpace(query.Get("group_number")),
		Institute:     strings.TrimSpace(query.Get("institute")),
		About:         strings.TrimSpace(query.Get("about")),
		PositionTitle: strings.TrimSpace(query.Get("position_title")),
		RoleName:      strings.TrimSpace(query.Get("role_name")),
		Limit:         20,
		Offset:        0,
	}

	rawIncludeArchived := strings.TrimSpace(query.Get("include_archived"))
	if rawIncludeArchived != "" {
		value, err := strconv.ParseBool(rawIncludeArchived)
		if err != nil {
			return usersSearchRequest{}, &shared.Error{
				Code:    "validation.include_archived",
				Message: "include_archived must be boolean",
			}
		}
		request.IncludeArchived = value
	}

	rawLimit := strings.TrimSpace(query.Get("limit"))
	if rawLimit != "" {
		value, err := strconv.ParseInt(rawLimit, 10, 32)
		if err != nil || value < 1 || value > 50 {
			return usersSearchRequest{}, &shared.Error{
				Code:    "validation.limit",
				Message: "limit must be between 1 and 50",
			}
		}
		request.Limit = int32(value)
	}

	rawOffset := strings.TrimSpace(query.Get("offset"))
	if rawOffset != "" {
		value, err := strconv.ParseInt(rawOffset, 10, 32)
		if err != nil || value < 0 {
			return usersSearchRequest{}, &shared.Error{
				Code:    "validation.offset",
				Message: "offset must be zero or positive",
			}
		}
		request.Offset = int32(value)
	}

	return request, nil
}

func (r usersSearchRequest) toInput() appsearch.SearchUsersInput {
	return appsearch.SearchUsersInput{
		FirstName:       r.FirstName,
		LastName:        r.LastName,
		MiddleName:      r.MiddleName,
		Login:           r.Login,
		GroupNumber:     r.GroupNumber,
		Institute:       r.Institute,
		About:           r.About,
		PositionTitle:   r.PositionTitle,
		RoleName:        r.RoleName,
		IncludeArchived: r.IncludeArchived,
		Limit:           r.Limit,
		Offset:          r.Offset,
	}
}

func newUsersSearchResponse(result appsearch.SearchUsersResult) usersSearchResponse {
	response := usersSearchResponse{
		Items: make([]searchUserItem, 0, len(result.Items)),
		Total: result.Total,
	}
	for _, item := range result.Items {
		response.Items = append(response.Items, searchUserItem{
			ID:              string(item.ID),
			FirstName:       item.FirstName,
			LastName:        item.LastName,
			MiddleName:      item.MiddleName,
			Login:           item.Login,
			GroupNumber:     item.GroupNumber,
			Institute:       item.Institute,
			Phone:           item.Phone,
			SocialLinks:     item.SocialLinks,
			About:           item.About,
			BirthDate:       item.BirthDate,
			GradebookNumber: item.GradebookNumber,
		})
	}
	return response
}
