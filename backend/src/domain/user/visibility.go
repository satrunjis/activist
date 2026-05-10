package user

type Field string

const (
	FieldID              Field = "id"
	FieldLogin           Field = "login"
	FieldFirstName       Field = "first_name"
	FieldLastName        Field = "last_name"
	FieldMiddleName      Field = "middle_name"
	FieldGradebookNumber Field = "gradebook_number"
	FieldGroupNumber     Field = "group_number"
	FieldInstitute       Field = "institute"
	FieldBirthDate       Field = "birth_date"
	FieldPhone           Field = "phone"
	FieldSocialLinks     Field = "social_links"
	FieldAbout           Field = "about"
	FieldMemberships     Field = "memberships"
)

type ProfileAccess struct {
	IsSelf          bool
	IsSystemAdmin   bool
	CanViewContacts bool
}

// VisibleFields returns the set of profile fields the actor may read.
func VisibleFields(access ProfileAccess) map[Field]bool {
	visible := map[Field]bool{
		FieldID:          true,
		FieldFirstName:   true,
		FieldLastName:    true,
		FieldMiddleName:  true,
		FieldMemberships: true,
	}
	if access.IsSelf || access.IsSystemAdmin || access.CanViewContacts {
		visible[FieldLogin] = true
		visible[FieldGradebookNumber] = true
		visible[FieldGroupNumber] = true
		visible[FieldInstitute] = true
		visible[FieldPhone] = true
		visible[FieldSocialLinks] = true
		visible[FieldAbout] = true
	}
	if access.IsSelf || access.IsSystemAdmin {
		visible[FieldBirthDate] = true
	}
	return visible
}
