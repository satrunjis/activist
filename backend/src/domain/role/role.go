package role

import (
	"time"

	"activist-base/src/domain/shared"
)

var (
	ErrInvalidPermCode   = &shared.Error{Code: "role.invalid_permission_code", Message: "unknown permission code"}
	ErrInvalidScope      = &shared.Error{Code: "role.invalid_scope_mode", Message: "unknown scope mode"}
	ErrBadPermSetVersion = &shared.Error{Code: "role.unsupported_permission_set_version", Message: "unsupported permission set format version"}
	ErrBadPermSet        = &shared.Error{Code: "role.malformed_permission_set", Message: "malformed permission set payload"}
	ErrSystemAdminInRole = &shared.Error{Code: "role.system_admin_in_role", Message: "system_admin cannot be assigned through a user-managed role"}
)

type Role struct {
	ID          shared.RoleID
	Name        string
	Permissions PermissionSet
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Normalize trims whitespace from all string fields in place.
// Call before Validate.
func (r *Role) Normalize() {
	r.Name = shared.Trim(r.Name)
}

// Validate checks structural invariants. Assumes fields have been normalized.
func (r Role) Validate() error {
	if err := shared.RequireID("role.id", string(r.ID)); err != nil {
		return err
	}
	if err := shared.RequireText("role.name", r.Name, shared.MaxName); err != nil {
		return err
	}
	return nil
}

// ValidateUserManaged additionally forbids embedding the SystemAdmin permission.
func (r Role) ValidateUserManaged() error {
	if err := r.Validate(); err != nil {
		return err
	}
	if r.Permissions.HasCode(SystemAdmin) {
		return ErrSystemAdminInRole
	}
	return nil
}
