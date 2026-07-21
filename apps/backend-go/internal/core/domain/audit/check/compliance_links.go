package check

import (
	"bytes"
	"context"
	"strings"
	"time"

	"golang.org/x/net/html"

	"monorepo/backend-go/internal/core/domain/audit/port"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// complianceTimeout bounds the homepage GET (Req 6.1).
const complianceTimeout = 10 * time.Second

// complianceGroupTitle names the check group; it returns three results.
const complianceGroupTitle = "Compliance Links"

// homepageUnreachableMessage is used for all three results when the homepage
// cannot be fetched (Req 6.8).
const homepageUnreachableMessage = "ไม่สามารถเข้าถึงหน้าแรกของเว็บไซต์เพื่อตรวจสอบลิงก์ได้"

// complianceCategory pairs a result title with the synonyms (Thai + English)
// that indicate a link points to that page.
type complianceCategory struct {
	title    string
	synonyms []string
}

// complianceCategories is a SLICE (not a map) so the result order is stable:
// Privacy Policy → Contact → About Us (Req 6.9).
var complianceCategories = []complianceCategory{
	{title: "Privacy Policy", synonyms: []string{"privacy policy", "privacy", "นโยบายความเป็นส่วนตัว", "ความเป็นส่วนตัว"}},
	{title: "Contact", synonyms: []string{"contact", "contact us", "ติดต่อ", "ติดต่อเรา"}},
	{title: "About Us", synonyms: []string{"about", "about us", "เกี่ยวกับ", "เกี่ยวกับเรา"}},
}

// anchor is a normalized <a> element: its visible text and href are both
// lower-cased and trimmed for case-insensitive matching (Req 6.9).
type anchor struct {
	text string
	href string
}

// ComplianceLinksCheck fetches the homepage and reports, per category, whether
// a matching link is present (Req 6.1–6.9).
type ComplianceLinksCheck struct{}

// NewComplianceLinksCheck constructs the compliance-links check group.
func NewComplianceLinksCheck() ComplianceLinksCheck { return ComplianceLinksCheck{} }

// Title returns the group name.
func (ComplianceLinksCheck) Title() string { return complianceGroupTitle }

// Evaluate fetches the homepage, extracts anchors, and returns one result per
// category. It never returns an error (Req 14.2). If the homepage is
// unreachable (transport error, nil response, or status != 200) all three
// results fail with homepageUnreachableMessage and no matching is performed
// (Req 6.8).
func (c ComplianceLinksCheck) Evaluate(ctx context.Context, target Target, f port.Fetcher) []valueobject.Result {
	url := originURL(target, homepagePath)

	res, err := f.Fetch(ctx, "GET", url, complianceTimeout)
	if err != nil || res == nil || res.StatusCode != 200 {
		return allFail(homepageUnreachableMessage)
	}

	anchors := extractAnchors(res.Body)

	results := make([]valueobject.Result, 0, len(complianceCategories))
	for _, cat := range complianceCategories {
		if anchorsMatchCategory(anchors, cat) {
			results = append(results, result(cat.title, valueobject.StatusPass,
				"พบลิงก์ "+cat.title+" บนหน้าแรก"))
		} else {
			results = append(results, result(cat.title, valueobject.StatusFail,
				"ไม่พบลิงก์ "+cat.title+" บนหน้าแรก"))
		}
	}
	return results
}

// allFail returns a failing result for every compliance category with the same
// message, preserving category order (Req 6.8).
func allFail(message string) []valueobject.Result {
	results := make([]valueobject.Result, 0, len(complianceCategories))
	for _, cat := range complianceCategories {
		results = append(results, result(cat.title, valueobject.StatusFail, message))
	}
	return results
}

// extractAnchors parses the HTML body and returns every <a> element as a
// normalized anchor. A parse failure yields an empty slice rather than a panic
// (Req 6.1) — all categories then fail without crashing.
func extractAnchors(body []byte) []anchor {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil
	}

	var anchors []anchor
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			var href string
			for _, attr := range n.Attr {
				if strings.ToLower(attr.Key) == "href" {
					href = attr.Val
					break
				}
			}
			anchors = append(anchors, anchor{
				text: normalize(anchorText(n)),
				href: normalize(href),
			})
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return anchors
}

// anchorText concatenates the text of all descendant text nodes, separated by
// spaces, so nested markup (e.g. <a><span>About</span> Us</a>) still yields the
// full visible text (Req 6.9).
func anchorText(n *html.Node) string {
	var parts []string
	var collect func(*html.Node)
	collect = func(node *html.Node) {
		if node.Type == html.TextNode {
			if t := strings.TrimSpace(node.Data); t != "" {
				parts = append(parts, t)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			collect(child)
		}
	}
	collect(n)
	return strings.Join(parts, " ")
}

// normalize lower-cases and trims a value for case-insensitive matching.
func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// anchorsMatchCategory reports whether any anchor indicates the given category.
// The anchor's visible text is ALWAYS matched; its href is matched only when
// navigable per isNavigableHref (Req 6.9).
func anchorsMatchCategory(anchors []anchor, cat complianceCategory) bool {
	for _, a := range anchors {
		for _, syn := range cat.synonyms {
			if strings.Contains(a.text, syn) {
				return true
			}
			if isNavigableHref(a.href) && strings.Contains(a.href, syn) {
				return true
			}
		}
	}
	return false
}

// isNavigableHref reports whether an href is a real navigation target that may
// participate in href matching (Req 6.9).
//
// Navigable: starts with "/", "http://", "https://", "mailto:", or is a
// scheme-less relative path.
//
// NOT navigable: empty, "#" or starting with "#", "javascript:"/"data:"/
// "tel:", or containing "void" (e.g. "javascript:void(0)"). A value that has a
// "scheme:" prefix before the path (other than the allowed ones) is treated as
// non-navigable.
func isNavigableHref(href string) bool {
	h := strings.TrimSpace(href)
	if h == "" {
		return false
	}
	if strings.HasPrefix(h, "#") {
		return false
	}

	lower := strings.ToLower(h)
	if strings.Contains(lower, "void") {
		return false
	}
	if strings.HasPrefix(lower, "javascript:") ||
		strings.HasPrefix(lower, "data:") ||
		strings.HasPrefix(lower, "tel:") {
		return false
	}

	// Explicitly allowed navigable prefixes.
	if strings.HasPrefix(h, "/") ||
		strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "mailto:") {
		return true
	}

	// A scheme-less relative path is navigable. If the value carries some other
	// "scheme:" (a colon appearing before any '/'), treat it as non-navigable.
	if colon := strings.Index(h, ":"); colon >= 0 {
		slash := strings.Index(h, "/")
		if slash < 0 || colon < slash {
			return false
		}
	}
	return true
}
