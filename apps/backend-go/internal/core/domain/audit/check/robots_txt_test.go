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

// Feature: adready-checker, Property 7
//
// Property 7: the robots.txt result is mapped from HTTP status 200 plus syntax
// validity. Pass iff status is 200 AND every non-empty, non-comment line is a
// "<known directive>: value" pair; an empty body with 200 passes. Non-200 or a
// malformed line yields fail. The generator covers valid/invalid syntax,
// non-ASCII values, odd whitespace, and CRLF line endings.
//
// Validates: Requirements 5.2, 5.3, 5.4, 5.5
func TestRobotsTxtResultMapping_Property7(t *testing.T) {
	// robotsScenario carries a generated body plus the independently-computed
	// expectation of whether its syntax is valid.
	type robotsScenario struct {
		unreachable       bool
		status            int
		body              string
		expectValidSyntax bool
	}

	// validLines are always acceptable (blank, comment, or known directive with
	// odd whitespace, empty value, or non-ASCII value).
	validLines := []string{
		"",
		"   ",
		"\t",
		"# a comment",
		"   # indented comment",
		"User-agent: *",
		"user-agent:*",
		"Disallow:",
		"  Disallow:   /private  ",
		"ALLOW: /public",
		"Sitemap: https://example.com/sitemap.xml",
		"Crawl-delay: 10",
		"Disallow: /ผู้ใช้/ข้อมูล",
		"User-agent: บอทภาษาไทย",
	}
	// invalidLines are meaningful lines that violate the syntax rule.
	invalidLines := []string{
		"this line has no colon",
		"unknown-directive: value",
		"เมนูภาษาไทยไม่มีโคลอน",
		"randomfield: something",
		"12345",
	}

	property := func(sc robotsScenario) bool {
		f := &fakeFetcher{}
		if sc.unreachable {
			f.fetchErr = errUnreachable
		} else {
			f.fetchResult = &port.FetchResult{StatusCode: sc.status, Body: []byte(sc.body)}
		}

		results := NewRobotsTxtCheck().Evaluate(context.Background(), testTarget("https://example.com"), f)
		if len(results) != 1 {
			return false
		}
		got := results[0]
		if got.Title != "robots.txt" {
			return false
		}

		wantPass := !sc.unreachable && sc.status == 200 && sc.expectValidSyntax
		if wantPass {
			return got.Status == valueobject.StatusPass
		}
		return got.Status == valueobject.StatusFail && got.Message != ""
	}

	gen := func(values []reflect.Value, rnd *rand.Rand) {
		statuses := []int{200, 200, 200, 404, 500, 301}
		sep := "\n"
		if rnd.Intn(2) == 0 {
			sep = "\r\n" // exercise CRLF normalization
		}

		numLines := rnd.Intn(6)
		lines := make([]string, 0, numLines)
		valid := true
		for i := 0; i < numLines; i++ {
			if rnd.Intn(3) == 0 {
				lines = append(lines, invalidLines[rnd.Intn(len(invalidLines))])
				valid = false
			} else {
				lines = append(lines, validLines[rnd.Intn(len(validLines))])
			}
		}

		sc := robotsScenario{
			unreachable:       rnd.Intn(5) == 0,
			status:            statuses[rnd.Intn(len(statuses))],
			body:              strings.Join(lines, sep),
			expectValidSyntax: valid,
		}
		values[0] = reflect.ValueOf(sc)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 300, Values: gen}); err != nil {
		t.Fatalf("Property 7 failed: %v", err)
	}
}
