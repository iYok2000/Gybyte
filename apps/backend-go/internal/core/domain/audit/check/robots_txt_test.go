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
// Property 7: the robots.txt result is mapped from HTTP status 200, syntax
// validity, and whether the catch-all group blocks the whole site. Pass iff
// status is 200 AND every non-empty, non-comment line is a
// "<known directive>: value" pair AND the "User-agent: *" group does NOT block
// the entire site (a "Disallow: /" not overridden by an "Allow: /"). An empty
// body with 200 passes. Non-200, a malformed line, or a site-wide block yields
// fail. The generator covers valid/invalid syntax, non-ASCII values, odd
// whitespace, CRLF line endings, and blocking / non-blocking catch-all groups.
//
// Validates: Requirements 5.2, 5.3, 5.4, 5.5
func TestRobotsTxtResultMapping_Property7(t *testing.T) {
	// robotsScenario carries a generated body plus the independently-computed
	// expectations of whether its syntax is valid and whether it blocks the
	// whole site.
	type robotsScenario struct {
		unreachable       bool
		status            int
		body              string
		expectValidSyntax bool
		expectSiteBlocked bool
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

		wantPass := !sc.unreachable && sc.status == 200 && sc.expectValidSyntax && !sc.expectSiteBlocked
		if wantPass {
			return got.Status == valueobject.StatusPass
		}
		return got.Status == valueobject.StatusFail && got.Message != ""
	}

	// blockGroups are appended (as valid syntax) to exercise site-wide block
	// detection. Each entry pairs a set of lines for a catch-all group with the
	// independently-known expectation of whether it blocks the whole site.
	type blockGroup struct {
		lines   []string
		blocked bool
	}
	blockGroups := []blockGroup{
		// No explicit group at all.
		{lines: nil, blocked: false},
		// Catch-all fully blocked.
		{lines: []string{"User-agent: *", "Disallow: /"}, blocked: true},
		// Catch-all blocked then re-allowed in the same group — not blocked.
		{lines: []string{"User-agent: *", "Disallow: /", "Allow: /"}, blocked: false},
		// Catch-all disallowing only a sub-path — not blocked.
		{lines: []string{"User-agent: *", "Disallow: /private"}, blocked: false},
		// Catch-all with an empty disallow (allow everything) — not blocked.
		{lines: []string{"User-agent: *", "Disallow:"}, blocked: false},
		// A specific bot is fully blocked but the catch-all is not affected.
		{lines: []string{"User-agent: Googlebot", "Disallow: /"}, blocked: false},
		// Odd whitespace around the blocking rule is still a block.
		{lines: []string{"User-agent:   *", "  Disallow:   /  "}, blocked: true},
		// Inline comments after the UA token and the blocking rule are stripped,
		// so this still counts as a full-site block.
		{lines: []string{"User-agent: *  # all bots", "Disallow: /   # block everything"}, blocked: true},
		// An inline comment must not turn a sub-path disallow into a root block.
		{lines: []string{"User-agent: *", "Disallow: /private   # secret area"}, blocked: false},
	}

	gen := func(values []reflect.Value, rnd *rand.Rand) {
		statuses := []int{200, 200, 200, 404, 500, 301}
		sep := "\n"
		if rnd.Intn(2) == 0 {
			sep = "\r\n" // exercise CRLF normalization
		}

		// The block group is emitted FIRST. Because every block group ends with
		// a rule line (not a User-agent line), the random lines that follow
		// start fresh groups and cannot merge into the catch-all group, so
		// bg.blocked remains an accurate, independent expectation. None of the
		// random line pool entries introduce a "Disallow: /" or "Allow: /", so
		// they never alter site-wide block detection.
		bg := blockGroups[rnd.Intn(len(blockGroups))]

		numLines := rnd.Intn(6)
		lines := make([]string, 0, numLines+len(bg.lines))
		lines = append(lines, bg.lines...)
		valid := true
		for range numLines {
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
			expectSiteBlocked: bg.blocked,
		}
		values[0] = reflect.ValueOf(sc)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 300, Values: gen}); err != nil {
		t.Fatalf("Property 7 failed: %v", err)
	}
}
