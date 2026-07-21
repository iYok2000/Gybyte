package check

import (
	"context"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"monorepo/backend-go/internal/core/domain/audit/port"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// Feature: adready-checker, Property 5
//
// Property 5: the ads.txt result is mapped purely from HTTP status + content.
// It is pass iff the status is 200 AND the trimmed body has length >= 1;
// otherwise it is fail with a message naming the cause (status code / empty
// content / unreachable).
//
// Validates: Requirements 3.2, 3.3, 3.4, 3.5
func TestAdsTxtResultMapping_Property5(t *testing.T) {
	// adsTxtScenario is a generated input for the ads.txt check.
	type adsTxtScenario struct {
		unreachable bool
		status      int
		body        string
	}

	property := func(sc adsTxtScenario) bool {
		f := &fakeFetcher{}
		if sc.unreachable {
			f.fetchErr = errUnreachable
		} else {
			f.fetchResult = &port.FetchResult{StatusCode: sc.status, Body: []byte(sc.body)}
		}

		results := NewAdsTxtCheck().Evaluate(context.Background(), testTarget("https://example.com"), f)
		if len(results) != 1 {
			return false
		}
		got := results[0]
		if got.Title != "ads.txt" {
			return false
		}

		wantPass := !sc.unreachable && sc.status == 200 && len(strings.TrimSpace(sc.body)) >= 1
		if wantPass {
			return got.Status == valueobject.StatusPass
		}

		// Every non-pass outcome must be a fail with a non-empty message.
		if got.Status != valueobject.StatusFail || got.Message == "" {
			return false
		}
		return true
	}

	// Custom generator: cover reachable/unreachable, a spread of status codes,
	// and bodies that are empty, whitespace-only, or meaningful.
	gen := func(values []reflect.Value, rnd *rand.Rand) {
		statuses := []int{200, 200, 204, 301, 403, 404, 500}
		bodies := []string{"", "   ", "\n\t ", "google.com, pub-1, DIRECT", "x", " padded "}
		sc := adsTxtScenario{
			unreachable: rnd.Intn(4) == 0,
			status:      statuses[rnd.Intn(len(statuses))],
			body:        bodies[rnd.Intn(len(bodies))],
		}
		values[0] = reflect.ValueOf(sc)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 200, Values: gen}); err != nil {
		t.Fatalf("Property 5 failed: %v", err)
	}
}
