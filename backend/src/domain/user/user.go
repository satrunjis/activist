package user

import (
	"time"

	"activist-base/src/domain/shared"
)

type User struct {
	ID              shared.UserID
	Login           string
	PasswordHash    string
	FirstName       string
	LastName        string
	MiddleName      string
	GradebookNumber string
	GroupNumber     string
	Institute       string
	BirthDate       *time.Time
	Phone           string
	SocialLinks     []shared.Link
	About           string
}

// Normalize trims whitespace from all string fields in place.
// Call before Validate.
func (u *User) Normalize() {
	u.Login = shared.Trim(u.Login)
	u.FirstName = shared.Trim(u.FirstName)
	u.LastName = shared.Trim(u.LastName)
	u.MiddleName = shared.Trim(u.MiddleName)
	u.GradebookNumber = shared.Trim(u.GradebookNumber)
	u.GroupNumber = shared.Trim(u.GroupNumber)
	u.Institute = shared.Trim(u.Institute)
	u.Phone = shared.Trim(u.Phone)
	u.About = shared.Trim(u.About)
	shared.NormalizeLinks(u.SocialLinks)
}

// Validate checks structural invariants. Assumes fields have been normalized.
func (u User) Validate() error {
	if err := shared.RequireID("user.id", string(u.ID)); err != nil {
		return err
	}
	if err := validateLogin(u.Login); err != nil {
		return err
	}
	if err := shared.RequireText("user.password_hash", u.PasswordHash, shared.MaxPasswordHash); err != nil {
		return err
	}
	if err := shared.RequireText("user.first_name", u.FirstName, shared.MaxShortText); err != nil {
		return err
	}
	if err := shared.AllowText("user.last_name", u.LastName, shared.MaxShortText); err != nil {
		return err
	}
	if err := shared.AllowText("user.middle_name", u.MiddleName, shared.MaxShortText); err != nil {
		return err
	}
	if err := shared.AllowText("user.gradebook_number", u.GradebookNumber, 32); err != nil {
		return err
	}
	if err := shared.AllowText("user.group_number", u.GroupNumber, shared.MaxShortText); err != nil {
		return err
	}
	if err := shared.AllowText("user.institute", u.Institute, shared.MaxName); err != nil {
		return err
	}
	if err := shared.AllowText("user.phone", u.Phone, 32); err != nil {
		return err
	}
	if err := shared.AllowText("user.about", u.About, shared.MaxDescription); err != nil {
		return err
	}
	return shared.ValidateLinks(u.SocialLinks)
}

func validateLogin(login string) error {
	if len(login) < 3 {
		return shared.TooShort("user.login", 3)
	}
	if len(login) > 32 {
		return shared.TooLong("user.login", 32)
	}
	for _, r := range login {
		if r == '_' || ('0' <= r && r <= '9') || ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') {
			continue
		}
		return shared.BadFormat("user.login")
	}
	return nil
}
