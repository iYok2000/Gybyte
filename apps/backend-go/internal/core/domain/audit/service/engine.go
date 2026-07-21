// Package service holds the audit orchestration engine. The Engine runs the
// fixed set of Audit_Checks concurrently against a validated Target and returns
// the collected Audit_Results in a deterministic order.
package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"monorepo/backend-go/internal/core/domain/audit/check"
	"monorepo/backend-go/internal/core/domain/audit/port"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// ErrSiteUnreachable is returned by Run when the target website cannot be
// reached at all. The caller maps this to a 400 error with no partial results
// (Req 14.1).
var ErrSiteUnreachable = errors.New("target website is entirely unreachable")

// reachabilityProbeTimeout bounds the pre-flight HEAD probe used to decide
// whether the site is reachable before any checks are launched (Req 14.1).
const reachabilityProbeTimeout = 10 * time.Second

// perCheckTimeout is the hard ceiling applied to every individual check. It
// sits above the 30s HTTPS/TLS budget so a slow-but-progressing check is not
// cut off, while still guaranteeing the whole run terminates (Req 14.2).
const perCheckTimeout = 35 * time.Second

// internalErrorMessage is the generic, user-facing message recorded when a
// check panics. It reveals no internal detail (Req 14.6).
const internalErrorMessage = "ตรวจสอบรายการนี้ไม่สำเร็จเนื่องจากข้อผิดพลาดภายใน"

// Engine orchestrates the audit: it holds the ordered set of checks and the
// outbound Fetcher they use.
type Engine struct {
	checks  []check.Check
	fetcher port.Fetcher
}

// NewEngine constructs the Engine with the fixed, ordered set of AdSense
// readiness checks: ads.txt → HTTPS → robots.txt → compliance links. The order
// here defines the deterministic order of results returned by Run (Req 7.8).
func NewEngine(f port.Fetcher) *Engine {
	return &Engine{
		fetcher: f,
		checks: []check.Check{
			check.NewAdsTxtCheck(),
			check.NewHTTPSCheck(),
			check.NewRobotsTxtCheck(),
			check.NewComplianceLinksCheck(),
		},
	}
}

// newEngineWithChecks builds an Engine with a caller-supplied set of checks.
// It is unexported and exists only to support tests that inject stub checks;
// production code uses NewEngine.
func newEngineWithChecks(f port.Fetcher, checks []check.Check) *Engine {
	return &Engine{fetcher: f, checks: checks}
}

// Run executes the audit against target and returns the collected results in a
// deterministic order (Req 7.8).
//
// The flow is:
//
//  1. Reachability gate (Req 14.1): a HEAD probe to scheme://host/ within a 10s
//     budget runs BEFORE any goroutine is launched. A transport error (or nil
//     response) means the site is entirely unreachable, so Run returns
//     ErrSiteUnreachable with no partial results. Any HTTP response — including
//     4xx/5xx — counts as reachable. HEAD is used to avoid downloading the body.
//  2. Concurrent checks: each check runs in its own goroutine and writes into
//     its own pre-allocated slot, so no synchronization on shared state is
//     needed beyond the WaitGroup. Results are flattened in check order,
//     independent of completion order (Req 7.8).
//  3. Per-check isolation: runCheck bounds each check with a timeout and
//     recovers from panics, so one misbehaving check degrades to a single fail
//     result rather than crashing the process (Req 14.2, 14.6).
func (e *Engine) Run(ctx context.Context, target check.Target) ([]valueobject.Result, error) {
	// 1. Reachability gate — probe before launching any goroutine (Req 14.1).
	if !e.isSiteReachable(ctx, target) {
		return nil, ErrSiteUnreachable
	}

	// 2. Run every check concurrently, each writing to its own slot so the
	//    final order is deterministic regardless of completion order (Req 7.8).
	slots := make([][]valueobject.Result, len(e.checks))
	var wg sync.WaitGroup
	wg.Add(len(e.checks))
	for idx, c := range e.checks {
		go func(idx int, c check.Check) {
			defer wg.Done()
			slots[idx] = e.runCheck(ctx, c, target)
		}(idx, c)
	}
	wg.Wait()

	// Flatten in check order (Req 7.8).
	var results []valueobject.Result
	for _, slot := range slots {
		results = append(results, slot...)
	}
	return results, nil
}

// isSiteReachable fires a HEAD request at the target origin within the
// reachability budget. It returns true when any HTTP response is received
// (even 4xx/5xx) and false on a transport error, nil response, or timeout
// (Req 14.1).
func (e *Engine) isSiteReachable(ctx context.Context, target check.Target) bool {
	probeCtx, cancel := context.WithTimeout(ctx, reachabilityProbeTimeout)
	defer cancel()

	url := target.URL.Scheme + "://" + target.Host + "/"
	res, err := e.fetcher.Fetch(probeCtx, "HEAD", url, reachabilityProbeTimeout)
	return err == nil && res != nil
}

// runCheck evaluates a single check with a bounded context and panic isolation.
//
// The check is wrapped in context.WithTimeout(perCheckTimeout) so it cannot run
// forever, and in a recover() so a panic in one check does not crash the process
// (Gin's recovery middleware does not cover child goroutines). A recovered panic
// is reduced to a single fail result carrying a generic message that leaks no
// internal detail (Req 14.6).
func (e *Engine) runCheck(ctx context.Context, c check.Check, target check.Target) (results []valueobject.Result) {
	checkCtx, cancel := context.WithTimeout(ctx, perCheckTimeout)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			results = []valueobject.Result{
				newFailResult(c.Title(), internalErrorMessage),
			}
		}
	}()

	return c.Evaluate(checkCtx, target, e.fetcher)
}

// newFailResult builds a fail Audit_Result, falling back to a struct literal if
// the status invariant is somehow rejected (it never is for StatusFail).
func newFailResult(title, message string) valueobject.Result {
	r, err := valueobject.NewResult(title, valueobject.StatusFail, message)
	if err != nil {
		return valueobject.Result{Title: title, Status: valueobject.StatusFail, Message: message}
	}
	return r
}
