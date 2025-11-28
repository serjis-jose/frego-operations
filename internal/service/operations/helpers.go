package operations

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ============================================================
// HELPER FUNCTIONS - Type Conversions
// ============================================================

// Text type helpers (for priority_level which is now text)
func textFromString(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func textToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

// UUID helpers
func uuidToPgtype(u uuid.UUID) pgtype.UUID {
	var bytes [16]byte
	copy(bytes[:], u[:])
	return pgtype.UUID{Bytes: bytes, Valid: true}
}

func uuidFromPgtype(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id, err := uuid.FromBytes(u.Bytes[:])
	if err != nil {
		return nil
	}
	return &id
}

// Timestamp helpers
func timestampFromTime(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func timestampFromString(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func timestampToStringPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func timeFromPgtype(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// Date helpers
func dateFromTime(t *time.Time) pgtype.Date {
	if t == nil {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: *t, Valid: true}
}

func timeFromDate(d pgtype.Date) *time.Time {
	if !d.Valid {
		return nil
	}
	return &d.Time
}

func timeFromTimestamptz(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// Numeric helpers
func numericFromFloat64(f *float64) pgtype.Numeric {
	if f == nil {
		return pgtype.Numeric{Valid: false}
	}
	var num pgtype.Numeric
	if err := num.Scan(fmt.Sprintf("%f", *f)); err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return num
}

func float64FromNumeric(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, err := n.Float64Value()
	if err != nil {
		return nil
	}
	if f.Valid {
		return &f.Float64
	}
	return nil
}

// Int helpers
func int4FromInt32(i *int32) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *i, Valid: true}
}

func int32FromPgtype(i pgtype.Int4) *int32 {
	if !i.Valid {
		return nil
	}
	return &i.Int32
}

// Boolean helper
func boolFromPgtype(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

// String array helpers (for supporting_doc_url fields)
func stringArrayFromStrings(strs []string) []string {
	if strs == nil {
		return []string{}
	}
	return strs
}

func stringsFromStringArray(arr []string) []string {
	if arr == nil {
		return []string{}
	}
	return arr
}

// numericFromInt32 converts int32 pointer to pgtype.Numeric
func numericFromInt32(i *int32) pgtype.Numeric {
	if i == nil {
		return pgtype.Numeric{Valid: false}
	}
	var num pgtype.Numeric
	if err := num.Scan(fmt.Sprintf("%d", *i)); err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return num
}
