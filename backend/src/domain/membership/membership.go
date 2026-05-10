package membership

import "activist-base/src/domain/shared"

var (
	ErrAlreadyExists = &shared.Error{Code: "membership.already_exists", Message: "membership already exists"}
	ErrNotFound      = &shared.Error{Code: "membership.not_found", Message: "membership does not exist"}
)

type Membership struct {
	UserID     shared.UserID
	PositionID shared.PositionID
}

func (m Membership) Validate() error {
	if err := shared.RequireID("membership.user_id", string(m.UserID)); err != nil {
		return err
	}
	return shared.RequireID("membership.position_id", string(m.PositionID))
}
