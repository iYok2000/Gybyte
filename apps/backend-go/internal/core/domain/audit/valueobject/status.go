// Package valueobject defines immutable value objects for the audit domain.
// These types carry no infrastructure concerns (no net/http, no persistence)
// and express the invariants of an audit's outcome.
package valueobject

// Status is the outcome of a single Audit_Check on the wire.
// Only two values are valid: "pass" and "fail" (Req 8.3).
type Status string

const (
	// StatusPass indicates the check succeeded.
	StatusPass Status = "pass"
	// StatusFail indicates the check failed.
	StatusFail Status = "fail"
)

// IsValid reports whether the status is one of the two allowed wire values.
// Any other value (including the empty string) is invalid (Req 8.3).
func (s Status) IsValid() bool {
	return s == StatusPass || s == StatusFail
}

// String returns the wire-format representation ("pass"/"fail") used when
// serializing an Audit_Result to JSON.
func (s Status) String() string {
	return string(s)
}
