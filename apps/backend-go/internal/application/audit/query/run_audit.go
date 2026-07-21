// Package query holds the read/evaluate side of the audit CQRS flow. An audit
// inspects an external website and returns a result; it changes no internal
// state, so it belongs on the Query side (there is no Command/repository here).
package query

import (
	"context"
	"errors"

	"monorepo/backend-go/internal/application/audit/dto"
	"monorepo/backend-go/internal/application/audit/validation"
	"monorepo/backend-go/internal/core/domain/audit/check"
	"monorepo/backend-go/internal/core/domain/audit/service"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
	apperrors "monorepo/backend-go/pkg/errors"
)

// auditEngine is the narrow behavior RunAuditHandler needs from the engine.
// The concrete *service.Engine satisfies it, so container wiring is unchanged;
// the interface exists so error-mapping can be exercised with a stub engine.
type auditEngine interface {
	Run(ctx context.Context, target check.Target) ([]valueobject.Result, error)
}

// RunAuditHandler orchestrates a single audit: it validates and SSRF-guards the
// raw URL, runs the engine, and maps the domain results to the API response.
type RunAuditHandler struct {
	validator *validation.URLValidator
	engine    auditEngine
}

// NewRunAuditHandler wires the shared URL validator and the audit engine. It
// accepts the *service.Engine (which satisfies auditEngine).
func NewRunAuditHandler(validator *validation.URLValidator, engine *service.Engine) *RunAuditHandler {
	return &RunAuditHandler{validator: validator, engine: engine}
}

// Handle runs the audit for rawURL and returns the wire response.
//
// The flow, with error mapping, is:
//
//  1. ValidateAndNormalize — a failure is returned directly; it is a
//     *ValidationError → 400 (Req 2.1, 2.2).
//  2. GuardSSRF — a failure is returned directly BEFORE any outbound request;
//     it is a *BadRequestError → 400 (Req 2.3, 2.4, 2.7, 2.8).
//  3. engine.Run — if the site is entirely unreachable, map to a 400
//     BadRequestError with no partial results (Req 14.1); any other error maps
//     to a generic 500 InternalError that leaks no internal detail (Req 14.5).
//  4. NewAuditResponse — build the { auditScore, results } payload (Req 7.1,
//     8.1).
func (h *RunAuditHandler) Handle(ctx context.Context, rawURL string) (dto.AuditResponse, error) {
	// 1. Validate + normalize the user-supplied URL (Req 2.1, 2.2).
	normalized, err := h.validator.ValidateAndNormalize(rawURL)
	if err != nil {
		return dto.AuditResponse{}, err
	}

	// 2. SSRF guard before any outbound request (Req 2.3, 2.4, 2.8).
	if err := h.validator.GuardSSRF(ctx, normalized); err != nil {
		return dto.AuditResponse{}, err
	}

	// 3. Run the audit engine against the validated target.
	target := check.Target{URL: normalized, Host: normalized.Hostname()}
	results, err := h.engine.Run(ctx, target)
	if err != nil {
		if errors.Is(err, service.ErrSiteUnreachable) {
			// Entirely unreachable → 400, no partial results (Req 14.1).
			return dto.AuditResponse{}, apperrors.NewBadRequestError("the target website could not be reached")
		}
		// Any other failure → generic 500, no internal detail leaked (Req 14.5).
		return dto.AuditResponse{}, apperrors.NewInternalError("audit failed", err)
	}

	// 4. Map domain results to the API response (Req 8.1).
	return dto.NewAuditResponse(results), nil
}
