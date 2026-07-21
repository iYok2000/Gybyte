package check

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"monorepo/backend-go/internal/core/domain/audit/port"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// Feature: adready-checker, Property 8
//
// Property 8: compliance-link matching is case-insensitive, supports Thai and
// English synonyms, and matches href only for navigable hrefs (text always
// matches). Swapping the case of the whole page must not change any category's
// result. The generator produces random <a> sets (text + href) including
// non-ASCII text and non-navigable hrefs (#, javascript:void(0), data:, tel:).
//
// Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.9
func TestComplianceLinkMatching_Property8(t *testing.T) {
	// linkSpec is one generated <a> element.
	type linkSpec struct {
		text string
		href string
	}

	texts := []string{
		"Privacy Policy", "privacy", "นโยบายความเป็นส่วนตัว", "ความเป็นส่วนตัว",
		"Contact", "Contact Us", "ติดต่อ", "ติดต่อเรา",
		"About", "About Us", "เกี่ยวกับ", "เกี่ยวกับเรา",
		"Home", "Products", "ข่าวสาร", "",
	}
	hrefs := []string{
		"/privacy", "https://example.com/contact", "mailto:hi@example.com", "/about-us",
		"#", "#privacy", "javascript:void(0)", "data:text/html,privacy",
		"tel:123456", "", "privacy-policy.html", "javascript:contact()",
	}

	buildHTML := func(links []linkSpec) string {
		var b strings.Builder
		b.WriteString("<html><body>")
		for _, l := range links {
			b.WriteString(fmt.Sprintf(`<a href=%q>%s</a>`, l.href, l.text))
		}
		b.WriteString("</body></html>")
		return b.String()
	}

	statusesFor := func(body string) map[string]valueobject.Status {
		f := &fakeFetcher{fetchResult: &port.FetchResult{StatusCode: 200, Body: []byte(body)}}
		results := NewComplianceLinksCheck().Evaluate(context.Background(), testTarget("https://example.com"), f)
		m := make(map[string]valueobject.Status, len(results))
		for _, r := range results {
			m[r.Title] = r.Status
		}
		return m
	}

	property := func(links []linkSpec) bool {
		body := buildHTML(links)
		original := statusesFor(body)

		// Exactly three category results, in the fixed order.
		if len(original) != 3 {
			return false
		}
		if _, ok := original["Privacy Policy"]; !ok {
			return false
		}
		if _, ok := original["Contact"]; !ok {
			return false
		}
		if _, ok := original["About Us"]; !ok {
			return false
		}

		// Case-swapping the whole page must not change any result.
		swapped := statusesFor(swapCase(body))
		for title, st := range original {
			if swapped[title] != st {
				return false
			}
		}
		return true
	}

	gen := func(values []reflect.Value, rnd *rand.Rand) {
		n := rnd.Intn(6)
		links := make([]linkSpec, 0, n)
		for i := 0; i < n; i++ {
			links = append(links, linkSpec{
				text: texts[rnd.Intn(len(texts))],
				href: hrefs[rnd.Intn(len(hrefs))],
			})
		}
		values[0] = reflect.ValueOf(links)
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 300, Values: gen}); err != nil {
		t.Fatalf("Property 8 failed: %v", err)
	}
}

// swapCase inverts the case of ASCII letters, leaving non-ASCII (e.g. Thai)
// unchanged, so the case-insensitivity of matching can be probed.
func swapCase(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r - 32
		case r >= 'A' && r <= 'Z':
			return r + 32
		default:
			return r
		}
	}, s)
}
