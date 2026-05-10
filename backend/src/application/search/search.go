package search

import (
	"context"

	domainmembership "activist-base/src/domain/membership"
	"activist-base/src/domain/role"
	"activist-base/src/domain/shared"
	domainuser "activist-base/src/domain/user"
)

const (
	defaultLimit = 20
	maxLimit     = 50
)

type Repository interface {
	SearchUsers(ctx context.Context, params SearchUsersParams) (SearchUsersResult, error)
}

type AuthorizationService interface {
	ResolveForUser(ctx context.Context, actorID shared.UserID, targetUserID shared.UserID) (domainmembership.EffectivePermissions, error)
}

type Service struct {
	repo  Repository
	authz AuthorizationService
}

func NewService(repo Repository, authz AuthorizationService) *Service {
	return &Service{repo: repo, authz: authz}
}

type SearchUsersInput struct {
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

type SearchUsersParams struct {
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

type SearchUsersResult struct {
	Items []SearchUser
	Total int64
}

type SearchUser struct {
	ID              shared.UserID
	FirstName       string
	LastName        *string
	MiddleName      *string
	Login           *string
	GroupNumber     *string
	Institute       *string
	About           *string
	Phone           *string
	SocialLinks     *[]shared.Link
	BirthDate       *string
	GradebookNumber *string
}

func (s *Service) SearchUsers(ctx context.Context, actorID shared.UserID, input SearchUsersInput) (SearchUsersResult, error) {
	if actorID == "" {
		return SearchUsersResult{}, shared.ErrEmptyID
	}
	if input.Limit < 0 || input.Limit > maxLimit {
		return SearchUsersResult{}, &shared.Error{
			Code:    "validation.limit",
			Message: "limit must be between 1 and 50",
		}
	}
	if input.Offset < 0 {
		return SearchUsersResult{}, &shared.Error{
			Code:    "validation.offset",
			Message: "offset must be zero or positive",
		}
	}

	limit := input.Limit
	if limit == 0 {
		limit = defaultLimit
	}

	params := SearchUsersParams{
		FirstName:       input.FirstName,
		LastName:        input.LastName,
		MiddleName:      input.MiddleName,
		Login:           input.Login,
		GroupNumber:     input.GroupNumber,
		Institute:       input.Institute,
		About:           input.About,
		PositionTitle:   input.PositionTitle,
		RoleName:        input.RoleName,
		IncludeArchived: input.IncludeArchived,
		Limit:           limit,
		Offset:          input.Offset,
	}

	result, err := s.repo.SearchUsers(ctx, params)
	if err != nil {
		return SearchUsersResult{}, err
	}

	items := make([]SearchUser, 0, len(result.Items))
	for _, row := range result.Items {
		access, err := s.authz.ResolveForUser(ctx, actorID, row.ID)
		if err != nil {
			return SearchUsersResult{}, err
		}
		profileAccess := domainuser.ProfileAccess{
			IsSelf:          actorID == row.ID,
			IsSystemAdmin:   access.Has(role.SystemAdmin, false),
			CanViewContacts: access.Has(role.CanViewContacts, false),
		}
		// Security-critical: search uses same visibility policy as profile endpoint.
		visible := domainuser.VisibleFields(profileAccess)
		item := row
		if !visible[domainuser.FieldPhone] {
			item.Phone = nil
		}
		if !visible[domainuser.FieldSocialLinks] {
			item.SocialLinks = nil
		}
		if !visible[domainuser.FieldAbout] {
			item.About = nil
		}
		items = append(items, item)
	}

	return SearchUsersResult{
		Items: items,
		Total: result.Total,
	}, nil
}
