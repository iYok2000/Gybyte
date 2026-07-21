// Package dto defines the JSON response contract for the Audit_API and the
// mapping from domain Audit_Results to that wire shape. It lives in the
// application layer because it is a transport concern, not a domain invariant.
package dto

import (
	"math"

	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// AuditResultDTO is the wire representation of a single Audit_Result. Field tags
// pin the JSON schema (Req 8.2): title, status ("pass"/"fail"), and message.
type AuditResultDTO struct {
	Title   string `json:"title"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// AuditResponse is the top-level Audit_API response (Req 8.1).
//
// AuditScore is a *int so it can serialize to JSON null when the audit could
// not be evaluated (no results). A concrete pointer serializes to the integer
// score; nil serializes to null (indeterminate — Req 8.6).
type AuditResponse struct {
	AuditScore *int             `json:"auditScore"`
	Results    []AuditResultDTO `json:"results"`
}

// NewAuditResponse maps the domain results to the wire response.
//
//   - Every result is mapped to an AuditResultDTO preserving order, using
//     Status.String() for the wire status value (Req 8.1, 8.2).
//   - When there are no results, AuditScore is nil so it serializes to JSON
//     null (indeterminate — Req 8.6). This differs from the engine's internal
//     score-of-0 branch (Req 7.7); at the API surface an empty audit is
//     "unable to evaluate", never a passing/failing score.
//   - Otherwise the score is (pass / total * 100) rounded half-up via
//     floor(x + 0.5), and AuditScore points at that value (Req 8.5).
func NewAuditResponse(results []valueobject.Result) AuditResponse {
	dtos := make([]AuditResultDTO, 0, len(results))
	pass := 0
	for _, r := range results {
		if r.Status == valueobject.StatusPass {
			pass++
		}
		dtos = append(dtos, AuditResultDTO{
			Title:   r.Title,
			Status:  r.Status.String(),
			Message: r.Message,
		})
	}

	total := len(results)
	if total == 0 {
		// Indeterminate: no evaluable checks → auditScore is null (Req 8.6).
		return AuditResponse{AuditScore: nil, Results: dtos}
	}

	// round-half-up = floor(x + 0.5) (Req 8.5).
	score := int(math.Floor(float64(pass)/float64(total)*100 + 0.5))
	return AuditResponse{AuditScore: &score, Results: dtos}
}
