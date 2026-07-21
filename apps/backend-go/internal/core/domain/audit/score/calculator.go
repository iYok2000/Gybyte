// Package score computes the Audit_Score from a set of Audit_Results.
// It carries no infrastructure concerns and depends only on the audit value
// objects.
package score

import (
	"math"

	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// CalculateScore returns the Audit_Score as an integer in the range [0, 100].
//
// The score is the proportion of passing checks over the total number of
// checks, expressed as a percentage and rounded to the nearest integer with
// ties rounded up (round-half-up). Every check is weighted equally (Req 7.6),
// so all-pass yields 100 (Req 7.2) and no-pass yields 0 (Req 7.3). Only checks
// with StatusPass count toward the numerator; any other status counts as not
// passing (Req 7.4). The result is always within [0, 100] inclusive (Req 7.5).
func CalculateScore(results []valueobject.Result) int {
	total := len(results)
	// Defensive, internal-only guard against division by zero (Req 7.7). In the
	// normal flow the engine always assembles a fixed, non-empty set of checks,
	// so this branch is not reached; the API surface maps an empty result set to
	// a null auditScore separately (Req 8.6).
	if total == 0 {
		return 0
	}

	pass := 0
	for _, r := range results {
		if r.Status == valueobject.StatusPass {
			pass++
		}
	}

	// round-half-up: floor(x + 0.5) rounds to the nearest integer and rounds a
	// fractional part of exactly 0.5 upward (Req 7.1).
	return int(math.Floor(float64(pass)/float64(total)*100 + 0.5))
}
