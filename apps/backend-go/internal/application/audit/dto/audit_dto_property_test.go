package dto

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// Feature: adready-checker, Property 10
//
// Property 10: round-trip serialize/deserialize of AuditResponse preserves the
// data and the schema; a status outside {pass, fail} is rejected.
// Validates: Requirements 8.1, 8.2, 8.3

// genResults is a testing/quick Generator that produces a slice of valid
// domain Results. Statuses are constrained to the two valid wire values so the
// generated input space matches what the engine can actually produce (Req 8.3).
type genResults []valueobject.Result

// Generate implements quick.Generator, building 0..N results with random
// titles/messages and a valid pass/fail status.
func (genResults) Generate(rnd *rand.Rand, size int) reflect.Value {
	n := rnd.Intn(size + 1) // include the empty-slice case
	out := make([]valueobject.Result, 0, n)
	for i := 0; i < n; i++ {
		status := valueobject.StatusPass
		if rnd.Intn(2) == 0 {
			status = valueobject.StatusFail
		}
		out = append(out, valueobject.Result{
			Title:   randString(rnd),
			Status:  status,
			Message: randString(rnd),
		})
	}
	return reflect.ValueOf(genResults(out))
}

// randString produces a short random string that may include multi-byte runes,
// exercising JSON encoding across a realistic character space.
func randString(rnd *rand.Rand) string {
	alphabet := []rune("abcXYZ 0123_/.:นโยบายความเป็นส่วนตัว")
	n := rnd.Intn(12)
	rs := make([]rune, n)
	for i := range rs {
		rs[i] = alphabet[rnd.Intn(len(alphabet))]
	}
	return string(rs)
}

// TestProperty10_RoundTrip verifies that marshaling an AuditResponse to JSON
// and unmarshaling it back yields an equivalent value, preserving both the
// data and the schema (Req 8.1, 8.2).
func TestProperty10_RoundTrip(t *testing.T) {
	roundTrip := func(g genResults) bool {
		orig := NewAuditResponse([]valueobject.Result(g))

		data, err := json.Marshal(orig)
		if err != nil {
			return false
		}

		var back AuditResponse
		if err := json.Unmarshal(data, &back); err != nil {
			return false
		}

		// auditScore preserved (including the null/indeterminate case).
		if (orig.AuditScore == nil) != (back.AuditScore == nil) {
			return false
		}
		if orig.AuditScore != nil && *orig.AuditScore != *back.AuditScore {
			return false
		}

		// results preserved, order and every field (title/status/message).
		if len(orig.Results) != len(back.Results) {
			return false
		}
		for i := range orig.Results {
			if orig.Results[i] != back.Results[i] {
				return false
			}
			// status is always one of the valid wire values (Req 8.2, 8.3).
			if s := back.Results[i].Status; s != "pass" && s != "fail" {
				return false
			}
		}
		return true
	}

	if err := quick.Check(roundTrip, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatalf("round-trip property failed: %v", err)
	}
}

// TestProperty10_InvalidStatusRejected verifies that a status outside the set
// {pass, fail} is rejected by the domain constructor, so it can never reach the
// wire (Req 8.3).
func TestProperty10_InvalidStatusRejected(t *testing.T) {
	rejectsInvalid := func(raw string) bool {
		status := valueobject.Status(raw)
		if status == valueobject.StatusPass || status == valueobject.StatusFail {
			// Valid statuses must be accepted.
			_, err := valueobject.NewResult("t", status, "m")
			return err == nil
		}
		// Any other value must be rejected with ErrInvalidStatus.
		_, err := valueobject.NewResult("t", status, "m")
		return err == valueobject.ErrInvalidStatus
	}

	if err := quick.Check(rejectsInvalid, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatalf("invalid-status rejection property failed: %v", err)
	}
}
