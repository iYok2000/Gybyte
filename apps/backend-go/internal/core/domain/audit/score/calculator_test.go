package score

import (
	"math"
	"testing"
	"testing/quick"

	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// resultsFromFlags builds a slice of Audit_Results from pass flags. A true flag
// yields a StatusPass result; false yields StatusFail. This lets testing/quick
// drive the property with its native []bool generator.
func resultsFromFlags(flags []bool) []valueobject.Result {
	results := make([]valueobject.Result, len(flags))
	for i, pass := range flags {
		status := valueobject.StatusFail
		if pass {
			status = valueobject.StatusPass
		}
		results[i] = valueobject.Result{Title: "check", Status: status, Message: ""}
	}
	return results
}

// Feature: adready-checker, Property 9
// Audit_Score = round-half-up of (pass/total × 100) and is always within
// [0, 100]. All-pass yields 100, no-pass yields 0, only StatusPass counts, and
// every check is weighted equally.
// Validates: Requirements 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 8.5
func TestCalculateScoreProperty(t *testing.T) {
	property := func(flags []bool) bool {
		results := resultsFromFlags(flags)
		got := CalculateScore(results)

		total := len(flags)
		if total == 0 {
			// Empty set is the defensive internal-only branch (Req 7.7).
			return got == 0
		}

		pass := 0
		for _, f := range flags {
			if f {
				pass++
			}
		}

		// round-half-up computed independently (Req 7.1, 7.4, 7.5, 8.5).
		want := int(math.Floor(float64(pass)/float64(total)*100 + 0.5))
		if got != want {
			return false
		}

		// Always within [0, 100] inclusive (Req 7.5).
		if got < 0 || got > 100 {
			return false
		}

		// All-pass → 100 (Req 7.2), no-pass → 0 (Req 7.3).
		switch pass {
		case total:
			return got == 100
		case 0:
			return got == 0
		}
		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 200}); err != nil {
		t.Error(err)
	}
}

// TestCalculateScoreEqualWeighting verifies every check is weighted equally by
// confirming that permuting which checks pass does not change the score, as long
// as the pass count is the same (Req 7.6).
func TestCalculateScoreEqualWeighting(t *testing.T) {
	a := CalculateScore(resultsFromFlags([]bool{true, false, false, true}))
	b := CalculateScore(resultsFromFlags([]bool{false, true, true, false}))
	if a != b {
		t.Errorf("equal pass counts should yield equal scores: got %d and %d", a, b)
	}
}
