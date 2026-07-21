package check

import (
	"context"

	"monorepo/backend-go/internal/core/domain/audit/port"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// Check is the strategy interface for a single AdSense readiness check.
//
// Evaluate MUST NOT return an error. Any failure — unreachable resource,
// timeout, unexpected status, or malformed content — is encoded as a fail
// Audit_Result instead. This keeps each check self-contained so the engine can
// run them concurrently and isolate per-check failures without a partial abort
// (Req 14.2). A check returns one or more results (e.g. the compliance-links
// check returns three).
type Check interface {
	// Title returns the human-readable name of the check (or check group).
	Title() string
	// Evaluate runs the check against target using the provided Fetcher and
	// returns the resulting Audit_Result(s).
	Evaluate(ctx context.Context, target Target, f port.Fetcher) []valueobject.Result
}
