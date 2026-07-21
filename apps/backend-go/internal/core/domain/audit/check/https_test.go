package check

import (
	"context"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"monorepo/backend-go/internal/core/domain/audit/port"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// Feature: adready-checker, Property 6
//
// Property 6: a TLS certificate is valid only when ALL conditions hold
// simultaneously (trusted chain, hostname match, within validity window, not
// revoked). The HTTPS result is pass iff TLSInfo.Valid == true; every other
// case (invalid TLS, HTTP-only, connection failure/timeout) is fail with a
// reason.
//
// Validates: Requirements 4.1, 4.2, 4.3, 4.5
func TestHTTPSResultMapping_Property6(t *testing.T) {
	// tlsFacts is a generated set of TLS validity conditions.
	type tlsFacts struct {
		connFail   bool
		chainOK    bool
		hostnameOK bool
		windowOK   bool
		notRevoked bool
	}

	property := func(facts tlsFacts) bool {
		f := &fakeFetcher{}
		// A certificate is valid only when every condition is true (Req 4.1).
		valid := facts.chainOK && facts.hostnameOK && facts.windowOK && facts.notRevoked

		if facts.connFail {
			f.tlsErr = errUnreachable
		} else {
			reason := ""
			if !valid {
				reason = "certificate invalid"
			}
			f.tlsInfo = &port.TLSInfo{Valid: valid, Reason: reason}
		}

		results := NewHTTPSCheck().Evaluate(context.Background(), testTarget("https://example.com"), f)
		if len(results) != 1 {
			return false
		}
		got := results[0]
		if got.Title != "HTTPS" {
			return false
		}

		// Pass iff we connected and the certificate was valid (Req 4.2).
		wantPass := !facts.connFail && valid
		if wantPass {
			return got.Status == valueobject.StatusPass
		}
		return got.Status == valueobject.StatusFail && got.Message != ""
	}

	gen := func(values []reflect.Value, rnd *rand.Rand) {
		facts := tlsFacts{
			connFail:   rnd.Intn(4) == 0,
			chainOK:    rnd.Intn(2) == 0,
			hostnameOK: rnd.Intn(2) == 0,
			windowOK:   rnd.Intn(2) == 0,
			notRevoked: rnd.Intn(2) == 0,
		}
		values[0] = reflect.ValueOf(facts)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 200, Values: gen}); err != nil {
		t.Fatalf("Property 6 failed: %v", err)
	}
}
