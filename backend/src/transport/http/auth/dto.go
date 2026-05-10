package auth

import (
	"time"

	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
)

const dateLayout = "2006-01-02"

type registerRequest struct {
	Login           string `json:"login"`
	Password        string `json:"password"`
	FirstName       string `json:"first_name"`
	GradebookNumber string `json:"gradebook_number"`
	GroupNumber     string `json:"group_number"`
	Institute       string `json:"institute"`
	BirthDate       string `json:"birth_date"`

	LastName   string        `json:"last_name,omitempty"`
	MiddleName string        `json:"middle_name,omitempty"`
	Phone      string        `json:"phone,omitempty"`
	SocialLink []shared.Link `json:"social_links,omitempty"`
	About      string        `json:"about,omitempty"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type sessionResponse struct {
	ExpiresAt     time.Time `json:"expires_at"`
	IdleExpiresAt time.Time `json:"idle_expires_at"`
}

type authResponse struct {
	User    userResponse    `json:"user"`
	Session sessionResponse `json:"session"`
}

type sessionInspectResponse struct {
	User        userResponse    `json:"user"`
	Session     sessionResponse `json:"session"`
	CSRFToken   string          `json:"csrf_token"`
	Permissions []string        `json:"permissions,omitempty"`
}

type userResponse struct {
	ID              string        `json:"id"`
	Login           string        `json:"login"`
	FirstName       string        `json:"first_name"`
	LastName        string        `json:"last_name,omitempty"`
	MiddleName      string        `json:"middle_name,omitempty"`
	GradebookNumber string        `json:"gradebook_number,omitempty"`
	GroupNumber     string        `json:"group_number,omitempty"`
	Institute       string        `json:"institute,omitempty"`
	BirthDate       string        `json:"birth_date,omitempty"`
	Phone           string        `json:"phone,omitempty"`
	SocialLinks     []shared.Link `json:"social_links,omitempty"`
	About           string        `json:"about,omitempty"`
}

func newUserResponse(user domainuser.User) userResponse {
	response := userResponse{
		ID:              string(user.ID),
		Login:           user.Login,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		MiddleName:      user.MiddleName,
		GradebookNumber: user.GradebookNumber,
		GroupNumber:     user.GroupNumber,
		Institute:       user.Institute,
		Phone:           user.Phone,
		SocialLinks:     user.SocialLinks,
		About:           user.About,
	}
	if user.BirthDate != nil {
		response.BirthDate = user.BirthDate.UTC().Format(dateLayout)
	}
	return response
}

func newSessionResponse(expiresAt, idleExpiresAt time.Time) sessionResponse {
	return sessionResponse{
		ExpiresAt:     expiresAt.UTC(),
		IdleExpiresAt: idleExpiresAt.UTC(),
	}
}
