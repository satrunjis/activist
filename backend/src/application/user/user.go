package user

import (
	"time"

	domaindivision "activist-base/src/domain/division"
	domainmembership "activist-base/src/domain/membership"
	domainposition "activist-base/src/domain/position"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
)

// — EditProfile —————————————————————————————————————————————————————————————

type ProfilePatch struct {
	FirstName       *string
	LastName        *string
	MiddleName      *string
	GradebookNumber *string
	GroupNumber     *string
	Institute       *string
	BirthDate       *time.Time
	Phone           *string
	SocialLinks     *[]shared.Link
	About           *string
}

type EditProfileInput struct {
	ActorID  shared.UserID
	Existing domainuser.User
	Patch    ProfilePatch
	Access   domainmembership.EffectivePermissions
}

type EditProfileResult struct {
	User domainuser.User
}

type ReadProfileInput struct {
	ActorID    shared.UserID
	TargetUser domainuser.User
	Access     domainmembership.EffectivePermissions
}

type ReadProfileResult struct {
	User   domainuser.User
	Access domainuser.ProfileAccess
}

func ReadProfile(input ReadProfileInput) (ReadProfileResult, error) {
	if input.ActorID == "" || input.TargetUser.ID == "" {
		return ReadProfileResult{}, shared.ErrEmptyID
	}

	isSelf := input.ActorID == input.TargetUser.ID
	isSystemAdmin := input.Access.Has(role.SystemAdmin, false)
	canViewContacts := input.Access.Has(role.CanViewContacts, false)

	return ReadProfileResult{
		User: input.TargetUser,
		Access: domainuser.ProfileAccess{
			IsSelf:          isSelf,
			IsSystemAdmin:   isSystemAdmin,
			CanViewContacts: canViewContacts,
		},
	}, nil
}

func EditProfile(input EditProfileInput) (EditProfileResult, error) {
	if input.ActorID != input.Existing.ID {
		return EditProfileResult{}, shared.ErrForbidden
	}
	if !input.Access.Has(role.CanEditSelfProfile, true) {
		return EditProfileResult{}, shared.ErrForbidden
	}

	u := input.Existing
	if input.Patch.FirstName != nil {
		u.FirstName = *input.Patch.FirstName
	}
	if input.Patch.LastName != nil {
		u.LastName = *input.Patch.LastName
	}
	if input.Patch.MiddleName != nil {
		u.MiddleName = *input.Patch.MiddleName
	}
	if input.Patch.GradebookNumber != nil {
		u.GradebookNumber = *input.Patch.GradebookNumber
	}
	if input.Patch.GroupNumber != nil {
		u.GroupNumber = *input.Patch.GroupNumber
	}
	if input.Patch.Institute != nil {
		u.Institute = *input.Patch.Institute
	}
	if input.Patch.BirthDate != nil {
		t := input.Patch.BirthDate.UTC()
		u.BirthDate = &t
	}
	if input.Patch.Phone != nil {
		u.Phone = *input.Patch.Phone
	}
	if input.Patch.SocialLinks != nil {
		u.SocialLinks = append([]shared.Link(nil), *input.Patch.SocialLinks...)
	}
	if input.Patch.About != nil {
		u.About = *input.Patch.About
	}
	u.Normalize()
	if err := u.Validate(); err != nil {
		return EditProfileResult{}, err
	}
	return EditProfileResult{User: u}, nil
}

// — GetMemberships ——————————————————————————————————————————————————————————

type MembershipView struct {
	Membership domainmembership.Membership
	Position   domainposition.Position
	Division   domaindivision.Division
	RoleName   string
}

type GetMembershipsInput struct {
	ActorID    shared.UserID
	TargetUser domainuser.User
	Items      []MembershipView
}

type GetMembershipsResult struct {
	Items []MembershipView
}

func GetMemberships(input GetMembershipsInput) (GetMembershipsResult, error) {
	if input.ActorID == "" || input.TargetUser.ID == "" {
		return GetMembershipsResult{}, shared.ErrEmptyID
	}
	items := make([]MembershipView, 0, len(input.Items))
	for _, item := range input.Items {
		if item.Membership.UserID != input.TargetUser.ID {
			continue
		}
		if item.Position.IsArchived || item.Division.IsArchived {
			continue
		}
		items = append(items, item)
	}
	return GetMembershipsResult{Items: items}, nil
}
