package shared

import "strings"

const (
	MaxIDLength       = 64
	MaxShortText      = 128
	MaxName           = 255
	MaxURL            = 2048
	MaxDescription    = 4096
	MaxPayloadKey     = 128
	MaxPayloadValue   = 4096
	MaxPasswordHash   = 255
	MaxEventSubjectID = 255
)

// Trim removes leading and trailing whitespace.
func Trim(s string) string {
	return strings.TrimSpace(s)
}

// RequireText returns an error if value is blank or exceeds max bytes.
// Assumes value has already been trimmed.
func RequireText(field, value string, max int) error {
	if value == "" {
		return BlankField(field)
	}
	if len(value) > max {
		return TooLong(field, max)
	}
	return nil
}

// AllowText returns an error only if value exceeds max bytes (blank is OK).
// Assumes value has already been trimmed.
func AllowText(field, value string, max int) error {
	if len(value) > max {
		return TooLong(field, max)
	}
	return nil
}

// RequireID is RequireText with the shared ID length limit.
func RequireID(field, value string) error {
	return RequireText(field, value, MaxIDLength)
}
