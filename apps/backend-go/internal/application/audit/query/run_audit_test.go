package query

import (
	"context"
	"errors"
	"testing"

	"monorepo/backend-go/internal/application/audit/validation"
	"monorepo/backend-go/internal/core/domain/audit/check"
	"monorepo/backend-go/internal/core/domain/audit/service"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
	apperrors "monorepo/backend-go/pkg/errors"
)

// stubEngine is a fake audit engine that returns a preconfigured result/error,
// letting us exercise RunAuditHandler's error mapping without real network I/O.
type stubEngine struct {
	results []valueobject.Result
	err     error
}

func (s *stubEngine) Run(ctx context.Context, target check.Target) ([]valueobject.Result, error) {
	return s.results, s.err
}

// newHandlerWithEngine builds a handler with the real (shared) URL validator
// and a stub engine, so validation/SSRF run for real while engine behavior is
// controlled by the test.
func newHandlerWithEngine(e auditEngine) *RunAuditHandler {
	return &RunAuditHandler{validator: validation.NewURLValidator(), engine: e}
}

// publicURL resolves through validation and the SSRF guard (public IP literal,
// no DNS), so the handler reaches engine.Run.
const publicURL = "http://93.184.216.34"

// TestHandle_UnreachableSiteMapsTo400 verifies that an entirely unreachable
// site (engine returns ErrSiteUnreachable) is mapped to a 400 BadRequestError
// with no partial results (Req 14.1).
func TestHandle_UnreachableSiteMapsTo400(t *testing.T) {
	h := newHandlerWithEngine(&stubEngine{err: service.ErrSiteUnreachable})

	resp, err := h.Handle(context.Background(), publicURL)
	if err == nil {
		t.Fatal("expected an error for an unreachable site")
	}

	var badReq *apperrors.BadRequestError
	if !errors.As(err, &badReq) {
		t.Fatalf("expected *BadRequestError (400), got %T: %v", err, err)
	}

	// No partial results returned (Req 14.1).
	if resp.Results != nil || resp.AuditScore != nil {
		t.Fatalf("expected empty response on unreachable site, got %+v", resp)
	}
}

// TestHandle_InternalErrorMapsTo500 verifies that a non-reachability engine
// error is mapped to a generic 500 InternalError that does not leak the
// underlying cause in its client message (Req 14.5).
func TestHandle_InternalErrorMapsTo500(t *testing.T) {
	cause := errors.New("boom: internal detail that must not leak")
	h := newHandlerWithEngine(&stubEngine{err: cause})

	_, err := h.Handle(context.Background(), publicURL)
	if err == nil {
		t.Fatal("expected an error for an internal engine failure")
	}

	var internal *apperrors.InternalError
	if !errors.As(err, &internal) {
		t.Fatalf("expected *InternalError (500), got %T: %v", err, err)
	}

	// The client-facing message is generic and does not leak the cause (Req 14.5).
	if internal.Message() != "audit failed" {
		t.Fatalf("expected generic message %q, got %q", "audit failed", internal.Message())
	}
	// The real cause is still reachable via Unwrap for internal logging only.
	if !errors.Is(err, cause) {
		t.Fatal("expected the underlying cause to be wrapped for internal logging")
	}
}

// TestHandle_ValidationErrorMapsTo400 verifies an invalid URL is rejected as a
// *ValidationError (400) before any engine call (Req 2.1, 2.2).
func TestHandle_ValidationErrorMapsTo400(t *testing.T) {
	h := newHandlerWithEngine(&stubEngine{err: errors.New("engine should not be called")})

	_, err := h.Handle(context.Background(), "ftp://example.com")
	if err == nil {
		t.Fatal("expected a validation error for a disallowed scheme")
	}

	var valErr *apperrors.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *ValidationError (400), got %T: %v", err, err)
	}
}

// TestHandle_SSRFErrorMapsTo400 verifies a blocked host is rejected as a
// *BadRequestError (400) before any outbound request (Req 2.3, 2.4).
func TestHandle_SSRFErrorMapsTo400(t *testing.T) {
	h := newHandlerWithEngine(&stubEngine{err: errors.New("engine should not be called")})

	_, err := h.Handle(context.Background(), "http://localhost")
	if err == nil {
		t.Fatal("expected an SSRF guard error for localhost")
	}

	var badReq *apperrors.BadRequestError
	if !errors.As(err, &badReq) {
		t.Fatalf("expected *BadRequestError (400), got %T: %v", err, err)
	}
}
