package service

import (
	"context"
	"net/url"
	"testing"
	"testing/quick"
	"time"

	"monorepo/backend-go/internal/core/domain/audit/check"
	"monorepo/backend-go/internal/core/domain/audit/port"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// reachableFetcher is a port.Fetcher whose HEAD/GET always succeed so the
// engine's reachability gate lets the checks run. Its GET body / status and TLS
// validity are configurable for the checks that read them.
type reachableFetcher struct {
	status  int
	body    []byte
	tlsInfo *port.TLSInfo
}

func (f *reachableFetcher) Fetch(_ context.Context, _, _ string, _ time.Duration) (*port.FetchResult, error) {
	return &port.FetchResult{StatusCode: f.status, Body: f.body}, nil
}

func (f *reachableFetcher) FetchTLS(_ context.Context, _ string, _ time.Duration) (*port.TLSInfo, error) {
	if f.tlsInfo == nil {
		return &port.TLSInfo{Valid: false, Reason: "no tls"}, nil
	}
	return f.tlsInfo, nil
}

// okFetcher is a minimal fetcher whose HEAD probe succeeds; used by the stub
// engine test where the checks ignore the fetcher entirely.
type okFetcher struct{}

func (okFetcher) Fetch(_ context.Context, _, _ string, _ time.Duration) (*port.FetchResult, error) {
	return &port.FetchResult{StatusCode: 200}, nil
}

func (okFetcher) FetchTLS(_ context.Context, _ string, _ time.Duration) (*port.TLSInfo, error) {
	return &port.TLSInfo{Valid: true}, nil
}

func testTarget() check.Target {
	u, _ := url.Parse("https://example.com")
	return check.Target{URL: u, Host: u.Hostname()}
}

// stub behavior codes.
const (
	behaviorPass  = 0
	behaviorFail  = 1
	behaviorPanic = 2
	behaviorMulti = 3
)

// stubCheck is a fully controllable Check used to exercise the engine's
// concurrency, ordering, and panic isolation without any real network access.
type stubCheck struct {
	title    string
	behavior int
	// delay staggers completion so the goroutines finish out of declaration
	// order, proving the engine's output order is independent of completion
	// order (Req 7.8).
	delay time.Duration
}

func (s stubCheck) Title() string { return s.title }

func (s stubCheck) Evaluate(_ context.Context, _ check.Target, _ port.Fetcher) []valueobject.Result {
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	switch s.behavior {
	case behaviorPanic:
		panic("stub check panic")
	case behaviorFail:
		return []valueobject.Result{{Title: s.title, Status: valueobject.StatusFail, Message: "stub fail"}}
	case behaviorMulti:
		return []valueobject.Result{
			{Title: s.title + "-a", Status: valueobject.StatusPass, Message: "a"},
			{Title: s.title + "-b", Status: valueobject.StatusFail, Message: "b"},
		}
	default: // behaviorPass
		return []valueobject.Result{{Title: s.title, Status: valueobject.StatusPass, Message: "stub pass"}}
	}
}

// expectedResults mirrors what the engine should return for a stub: a panicking
// check collapses to a single fail result with internalErrorMessage; every
// other check yields its own results verbatim (Req 14.6).
func expectedResults(s stubCheck) []valueobject.Result {
	if s.behavior == behaviorPanic {
		return []valueobject.Result{{Title: s.title, Status: valueobject.StatusFail, Message: internalErrorMessage}}
	}
	return stubCheck{title: s.title, behavior: s.behavior}.Evaluate(context.Background(), check.Target{}, nil)
}

// Feature: adready-checker, Property 12
//
// The engine preserves the deterministic (declaration) order of results and
// isolates failed/timeout/panic checks: a panic in any check degrades to a
// single fail result without crashing the process, and all other checks still
// contribute their results in order (Req 7.8, 14.2, 14.6).
func TestProperty12_EngineOrderAndIsolation(t *testing.T) {
	property := func(behaviors []uint8) bool {
		// Cap the number of checks so staggered delays stay fast.
		if len(behaviors) > 8 {
			behaviors = behaviors[:8]
		}

		checks := make([]check.Check, len(behaviors))
		var want []valueobject.Result
		n := len(behaviors)
		for i, b := range behaviors {
			s := stubCheck{
				title:    "check-" + string(rune('A'+i)),
				behavior: int(b % 4),
				// Later-declared checks finish first, so completion order is the
				// reverse of declaration order.
				delay: time.Duration(n-i) * 200 * time.Microsecond,
			}
			checks[i] = s
			want = append(want, expectedResults(s)...)
		}

		engine := newEngineWithChecks(okFetcher{}, checks)
		got, err := engine.Run(context.Background(), testTarget())
		if err != nil {
			return false
		}

		if len(got) != len(want) {
			return false
		}
		for i := range want {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatalf("Property 12 failed: %v", err)
	}
}

// Feature: adready-checker, Property 11
//
// Every Audit_Result with a fail status carries a non-empty message so the
// dashboard can always tell the user what failed and how to fix it (Req 8.4).
// This drives the real check set via NewEngine with randomized fetcher
// responses.
func TestProperty11_EveryFailHasMessage(t *testing.T) {
	property := func(status int, body string, tlsValid bool) bool {
		fetcher := &reachableFetcher{
			status:  status,
			body:    []byte(body),
			tlsInfo: &port.TLSInfo{Valid: tlsValid, Reason: "invalid cert"},
		}

		engine := NewEngine(fetcher)
		results, err := engine.Run(context.Background(), testTarget())
		if err != nil {
			// A transport-level unreachable is a separate path with no results;
			// the fetcher here always responds, so this should not happen.
			return false
		}

		for _, r := range results {
			if r.Status == valueobject.StatusFail && r.Message == "" {
				return false
			}
		}
		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatalf("Property 11 failed: %v", err)
	}
}
