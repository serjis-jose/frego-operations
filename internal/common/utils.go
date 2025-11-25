package common

import (
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ResolveName determines the best display name from display, legal, and code fields.
func ResolveName(display pgtype.Text, legal pgtype.Text, code pgtype.Text) string {
	if v := textValuePtr(display); v != nil && *v != "" {
		return *v
	}
	if v := textValuePtr(legal); v != nil && *v != "" {
		return *v
	}
	if v := textValuePtr(code); v != nil && *v != "" {
		return *v
	}
	return ""
}

// ResolvePartyName determines the best display name from string pointers.
func ResolvePartyName(display *string, legal *string, code *string) string {
	if display != nil && *display != "" {
		return *display
	}
	if legal != nil && *legal != "" {
		return *legal
	}
	if code != nil && *code != "" {
		return *code
	}
	return ""
}

// TrimString trims whitespace from a string and returns nil if empty.
func TrimString(s string) *string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// Helper functions for pgtype conversions
func TextValue(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func TextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	value := t.String
	return &value
}

func textValuePtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func BoolValue(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

func TimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	value := t.Time
	return &value
}

func StringToPgtypeText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func TrimStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	return TrimString(*s)
}

func UUIDPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func UUIDToPgtype(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	var bytes [16]byte
	copy(bytes[:], u[:])
	return pgtype.UUID{Bytes: bytes, Valid: true}
}

func DatePtr(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	t := d.Time
	return &t
}

func DateToPgtype(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: t.UTC(), Valid: true}
}

func TimeToPgtype(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func BoolToPgtype(b bool) pgtype.Bool {
	return pgtype.Bool{Bool: b, Valid: true}
}

func BoolPtr(b pgtype.Bool) *bool {
	if !b.Valid {
		return nil
	}
	val := b.Bool
	return &val
}

func Int4Ptr(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	value := i.Int32
	return &value
}

func Int4Value(i pgtype.Int4) int32 {
	if !i.Valid {
		return 0
	}
	return i.Int32
}

func Int32ToPgtype(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func NumericPtr(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	val, err := n.Float64Value()
	if err != nil || !val.Valid {
		return nil
	}
	result := val.Float64
	return &result
}

func Float64ToNumeric(val *float64) pgtype.Numeric {
	if val == nil {
		return pgtype.Numeric{Valid: false}
	}
	var num pgtype.Numeric
	decimal := strconv.FormatFloat(*val, 'f', -1, 64)
	if err := num.Scan(decimal); err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return num
}

// GenerateHumanID produces a short unique identifier for human-friendly references.
func GenerateHumanID() string {
	raw := strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))
	if len(raw) > 10 {
		raw = raw[:10]
	}
	return "H-" + raw
}

// PgtypeTextToString converts pgtype.Text to string, returning empty string if invalid
func PgtypeTextToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

// PgtypeTextToStringPtr converts pgtype.Text to *string, returning nil if invalid
func PgtypeTextToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}
