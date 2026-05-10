package user

import (
	"time"

	appuser "activist-base/src/application/user"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
)

const dateLayout = "2006-01-02"

type profileResponse struct {
	ID              string         `json:"id"`
	FirstName       string         `json:"first_name"`
	LastName        string         `json:"last_name"`
	MiddleName      string         `json:"middle_name"`
	Login           *string        `json:"login,omitempty"`
	GradebookNumber *string        `json:"gradebook_number,omitempty"`
	GroupNumber     *string        `json:"group_number,omitempty"`
	Institute       *string        `json:"institute,omitempty"`
	BirthDate       *string        `json:"birth_date,omitempty"`
	Phone           *string        `json:"phone,omitempty"`
	SocialLinks     *[]shared.Link `json:"social_links,omitempty"`
	About           *string        `json:"about,omitempty"`
}

type patchProfileRequest struct {
	FirstName       *string        `json:"first_name"`
	LastName        *string        `json:"last_name"`
	MiddleName      *string        `json:"middle_name"`
	GradebookNumber *string        `json:"gradebook_number"`
	GroupNumber     *string        `json:"group_number"`
	Institute       *string        `json:"institute"`
	BirthDate       *string        `json:"birth_date"`
	Phone           *string        `json:"phone"`
	SocialLinks     *[]shared.Link `json:"social_links"`
	About           *string        `json:"about"`
}

type membershipsResponse struct {
	Items []membershipItem `json:"items"`
}

type membershipItem struct {
	PositionID   string `json:"position_id"`
	PositionName string `json:"position_name"`
	DivisionID   string `json:"division_id"`
	DivisionName string `json:"division_name"`
	RoleName     string `json:"role_name"`
}

func newProfileResponse(user domainuser.User, access domainuser.ProfileAccess) profileResponse {
	visible := domainuser.VisibleFields(access)

	response := profileResponse{
		ID:         string(user.ID),
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		MiddleName: user.MiddleName,
	}
	if visible[domainuser.FieldLogin] {
		response.Login = stringPtr(user.Login)
	}
	if visible[domainuser.FieldGradebookNumber] {
		response.GradebookNumber = stringPtr(user.GradebookNumber)
	}
	if visible[domainuser.FieldGroupNumber] {
		response.GroupNumber = stringPtr(user.GroupNumber)
	}
	if visible[domainuser.FieldInstitute] {
		response.Institute = stringPtr(user.Institute)
	}
	if visible[domainuser.FieldBirthDate] {
		response.BirthDate = formatDate(user.BirthDate)
	}
	if visible[domainuser.FieldPhone] {
		response.Phone = stringPtr(user.Phone)
	}
	if visible[domainuser.FieldSocialLinks] {
		response.SocialLinks = linksPtr(user.SocialLinks)
	}
	if visible[domainuser.FieldAbout] {
		response.About = stringPtr(user.About)
	}
	return response
}

func newMembershipsResponse(items []appuser.MembershipView) membershipsResponse {
	response := membershipsResponse{
		Items: make([]membershipItem, 0, len(items)),
	}
	for _, item := range items {
		response.Items = append(response.Items, membershipItem{
			PositionID:   string(item.Position.ID),
			PositionName: item.Position.Title,
			DivisionID:   string(item.Division.ID),
			DivisionName: item.Division.ShortName,
			RoleName:     item.RoleName,
		})
	}
	return response
}

func stringPtr(value string) *string {
	return &value
}

func linksPtr(links []shared.Link) *[]shared.Link {
	cloned := make([]shared.Link, 0, len(links))
	cloned = append(cloned, links...)
	return &cloned
}

func formatDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(dateLayout)
	return &formatted
}
