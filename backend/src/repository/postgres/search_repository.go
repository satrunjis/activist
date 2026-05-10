package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"activist-base/src/repository/postgres/sqlc"
)

type SearchRepository struct {
	queries *sqlc.Queries
}

func NewSearchRepository(queries *sqlc.Queries) *SearchRepository {
	return &SearchRepository{queries: queries}
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

type SearchUserRecord struct {
	ID              string
	FirstName       string
	LastName        *string
	MiddleName      *string
	Login           string
	GroupNumber     *string
	Institute       *string
	About           *string
	Phone           *string
	SocialLinks     []byte
	BirthDate       *time.Time
	GradebookNumber *string
}

type SearchUsersResult struct {
	Items []SearchUserRecord
	Total int64
}

func (r *SearchRepository) SearchUsers(ctx context.Context, params SearchUsersParams) (SearchUsersResult, error) {
	countParams := sqlc.CountSearchUsersParams{
		FirstName:       textOrNull(params.FirstName),
		LastName:        textOrNull(params.LastName),
		MiddleName:      textOrNull(params.MiddleName),
		Login:           textOrNull(params.Login),
		GroupNumber:     textOrNull(params.GroupNumber),
		Institute:       textOrNull(params.Institute),
		About:           textOrNull(params.About),
		PositionTitle:   textOrNull(params.PositionTitle),
		RoleName:        textOrNull(params.RoleName),
		IncludeArchived: params.IncludeArchived,
	}

	total, err := r.queries.CountSearchUsers(ctx, countParams)
	if err != nil {
		return SearchUsersResult{}, err
	}

	rows, err := r.queries.SearchUsers(ctx, sqlc.SearchUsersParams{
		FirstName:       countParams.FirstName,
		LastName:        countParams.LastName,
		MiddleName:      countParams.MiddleName,
		Login:           countParams.Login,
		GroupNumber:     countParams.GroupNumber,
		Institute:       countParams.Institute,
		About:           countParams.About,
		PositionTitle:   countParams.PositionTitle,
		RoleName:        countParams.RoleName,
		IncludeArchived: countParams.IncludeArchived,
		PageOffset:      params.Offset,
		PageLimit:       params.Limit,
	})
	if err != nil {
		return SearchUsersResult{}, err
	}

	items := make([]SearchUserRecord, 0, len(rows))
	for _, row := range rows {
		items = append(items, SearchUserRecord{
			ID:              row.ID,
			FirstName:       row.FirstName,
			LastName:        textPointer(row.LastName),
			MiddleName:      textPointer(row.MiddleName),
			Login:           row.Login,
			GroupNumber:     textPointer(row.GroupNumber),
			Institute:       textPointer(row.Institute),
			About:           textPointer(row.About),
			Phone:           textPointer(row.Phone),
			SocialLinks:     append([]byte(nil), row.SocialLinks...),
			BirthDate:       dateFromPg(row.BirthDate),
			GradebookNumber: textPointer(row.GradebookNumber),
		})
	}

	return SearchUsersResult{
		Items: items,
		Total: total,
	}, nil
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	s := value.String
	return &s
}
