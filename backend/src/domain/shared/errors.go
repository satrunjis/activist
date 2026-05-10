package shared

import "fmt"

// Error is a stable, matchable domain error.
type ErrorCode string

type Error struct {
	Code    ErrorCode
	Message string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// FieldError describes a validation problem with a specific input field.
type FieldErrorKind string

const (
	FieldBlank     FieldErrorKind = "blank"
	FieldTooLong   FieldErrorKind = "too_long"
	FieldTooShort  FieldErrorKind = "too_short"
	FieldBadFormat FieldErrorKind = "invalid_format"
)

type FieldError struct {
	Field string
	Kind  FieldErrorKind
	Limit int
}

func (e *FieldError) Error() string {
	if e == nil {
		return ""
	}
	switch e.Kind {
	case FieldBlank:
		return fmt.Sprintf("%q must not be blank", e.Field)
	case FieldTooLong:
		return fmt.Sprintf("%q exceeds max length %d", e.Field, e.Limit)
	case FieldTooShort:
		return fmt.Sprintf("%q must be at least %d characters", e.Field, e.Limit)
	case FieldBadFormat:
		return fmt.Sprintf("%q has invalid format", e.Field)
	default:
		return fmt.Sprintf("%q is invalid", e.Field)
	}
}

func BlankField(field string) error { return &FieldError{Field: field, Kind: FieldBlank} }
func TooLong(field string, max int) error {
	return &FieldError{Field: field, Kind: FieldTooLong, Limit: max}
}
func TooShort(field string, min int) error {
	return &FieldError{Field: field, Kind: FieldTooShort, Limit: min}
}
func BadFormat(field string) error { return &FieldError{Field: field, Kind: FieldBadFormat} }

// Cross-cutting sentinel errors.
var (
	ErrEmptyID   = &Error{Code: "shared.empty_id", Message: "id must not be empty"}
	ErrForbidden = &Error{Code: "access.forbidden", Message: "operation is forbidden"}
)
