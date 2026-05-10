package role

// PermissionCode identifies an atomic capability.
type PermissionCode uint8

const (
	CanAddMember      PermissionCode = 1
	CanRemoveMember   PermissionCode = 2
	CanAssignPosition PermissionCode = 3

	CanEditDivision      PermissionCode = 4
	CanManagePositions   PermissionCode = 5
	CanCreateSubdivision PermissionCode = 6
	CanArchiveDivision   PermissionCode = 7

	CanViewContacts    PermissionCode = 8
	CanEditSelfProfile PermissionCode = 9
	CanManageRoles     PermissionCode = 10
	CanViewAuditLog    PermissionCode = 11

	SystemAdmin PermissionCode = 255
)

func (c PermissionCode) IsValid() bool {
	switch c {
	case CanAddMember, CanRemoveMember, CanAssignPosition, CanEditDivision,
		CanManagePositions, CanCreateSubdivision, CanArchiveDivision,
		CanViewContacts, CanEditSelfProfile, CanManageRoles, CanViewAuditLog, SystemAdmin:
		return true
	}
	return false
}

func (c PermissionCode) IsSystemAdmin() bool { return c == SystemAdmin }

func (c PermissionCode) String() string {
	switch c {
	case CanAddMember:
		return "can_add_member"
	case CanRemoveMember:
		return "can_remove_member"
	case CanAssignPosition:
		return "can_assign_position"
	case CanEditDivision:
		return "can_edit_division"
	case CanManagePositions:
		return "can_manage_positions"
	case CanCreateSubdivision:
		return "can_create_subdivision"
	case CanArchiveDivision:
		return "can_archive_division"
	case CanViewContacts:
		return "can_view_contacts"
	case CanEditSelfProfile:
		return "can_edit_self_profile"
	case CanManageRoles:
		return "can_manage_roles"
	case CanViewAuditLog:
		return "can_view_audit_log"
	case SystemAdmin:
		return "system_admin"
	default:
		return "unknown"
	}
}

// ScopeMode controls how far a permission reaches in the division tree.
type ScopeMode uint8

const (
	ScopeSelf                  ScopeMode = 1
	ScopeCurrentDivision       ScopeMode = 2
	ScopeCurrentAndDescendants ScopeMode = 3
)

func (s ScopeMode) IsValid() bool {
	switch s {
	case ScopeSelf, ScopeCurrentDivision, ScopeCurrentAndDescendants:
		return true
	}
	return false
}

// Permission pairs a capability code with its scope.
type Permission struct {
	code  PermissionCode
	scope ScopeMode
}

func NewPermission(code PermissionCode, scope ScopeMode) (Permission, error) {
	if !code.IsValid() {
		return Permission{}, ErrInvalidPermCode
	}
	if !scope.IsValid() {
		return Permission{}, ErrInvalidScope
	}
	return Permission{code: code, scope: scope}, nil
}

func (p Permission) Code() PermissionCode { return p.code }
func (p Permission) Scope() ScopeMode     { return p.scope }

// PermissionSet is an unordered deduplicated collection of permissions.
type PermissionSet struct {
	permissions []Permission
}

func NewPermissionSet(perms []Permission) PermissionSet {
	seen := make(map[Permission]struct{}, len(perms))
	unique := make([]Permission, 0, len(perms))
	for _, p := range perms {
		if _, dup := seen[p]; dup {
			continue
		}
		seen[p] = struct{}{}
		unique = append(unique, p)
	}
	return PermissionSet{permissions: unique}
}

func (ps PermissionSet) Permissions() []Permission {
	out := make([]Permission, len(ps.permissions))
	copy(out, ps.permissions)
	return out
}

func (ps PermissionSet) Contains(code PermissionCode, scope ScopeMode) bool {
	for _, p := range ps.permissions {
		if p.code == code && p.scope == scope {
			return true
		}
	}
	return false
}

func (ps PermissionSet) HasCode(code PermissionCode) bool {
	for _, p := range ps.permissions {
		if p.code == code {
			return true
		}
	}
	return false
}

// Equal reports whether two sets contain exactly the same permissions,
// regardless of insertion order.
func (ps PermissionSet) Equal(other PermissionSet) bool {
	if len(ps.permissions) != len(other.permissions) {
		return false
	}
	for _, p := range ps.permissions {
		if !other.Contains(p.code, p.scope) {
			return false
		}
	}
	return true
}

// Binary encoding for database storage.
const (
	permSetVersion byte = 1
	permRecordSize int  = 2
)

func (ps PermissionSet) MarshalBinary() []byte {
	buf := make([]byte, 1+len(ps.permissions)*permRecordSize)
	buf[0] = permSetVersion
	for i, p := range ps.permissions {
		off := 1 + i*permRecordSize
		buf[off] = byte(p.code)
		buf[off+1] = byte(p.scope)
	}
	return buf
}

func PermissionSetFromBytes(data []byte) (PermissionSet, error) {
	if len(data) == 0 {
		return PermissionSet{}, nil
	}
	if data[0] != permSetVersion {
		return PermissionSet{}, ErrBadPermSetVersion
	}
	body := data[1:]
	if len(body)%permRecordSize != 0 {
		return PermissionSet{}, ErrBadPermSet
	}
	perms := make([]Permission, 0, len(body)/permRecordSize)
	for i := 0; i < len(body); i += permRecordSize {
		p, err := NewPermission(PermissionCode(body[i]), ScopeMode(body[i+1]))
		if err != nil {
			return PermissionSet{}, ErrBadPermSet
		}
		perms = append(perms, p)
	}
	return NewPermissionSet(perms), nil
}
