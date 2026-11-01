package check

import (
	"context"
	"fmt"
	"strings"
	"time"

	"monorepo/backend-go/internal/core/domain/audit/port"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// robotsTxtTimeout bounds the GET /robots.txt request (Req 5.1).
const robotsTxtTimeout = 10 * time.Second

// robotsTxtTitle is the human-readable name of this check.
const robotsTxtTitle = "robots.txt"

// knownRobotsDirectives is the set of recognized robots.txt directives, keyed
// by their lower-cased field name. A line is well-formed when it is "field:
// value" and field (case-insensitive) is one of these (Req 5.4).
var knownRobotsDirectives = map[string]struct{}{
	"user-agent":  {},
	"disallow":    {},
	"allow":       {},
	"sitemap":     {},
	"crawl-delay": {},
}

// RobotsTxtCheck verifies that robots.txt is present (HTTP 200) and, when it has
// content, that every meaningful line uses a known directive (Req 5.1–5.5).
type RobotsTxtCheck struct{}

// NewRobotsTxtCheck constructs the robots.txt readiness check.
func NewRobotsTxtCheck() RobotsTxtCheck { return RobotsTxtCheck{} }

// Title returns the check name.
func (RobotsTxtCheck) Title() string { return robotsTxtTitle }

// Evaluate fetches /robots.txt from the target origin and maps the outcome to a
// single Audit_Result. It never returns an error (Req 14.2).
//
// Pass criteria (Req 5.2): HTTP 200 (an empty body is acceptable) AND every
// non-empty, non-comment line is a "<known directive>: value" pair.
func (c RobotsTxtCheck) Evaluate(ctx context.Context, target Target, f port.Fetcher) []valueobject.Result {
	url := originURL(target, robotsTxtPath)

	res, err := f.Fetch(ctx, "GET", url, robotsTxtTimeout)
	if err != nil || res == nil {
		// Connection failed or timed out — resource unreachable (Req 5.5).
		return []valueobject.Result{result(robotsTxtTitle, valueobject.StatusFail,
			"ไม่สามารถเข้าถึงไฟล์ robots.txt ได้ (เชื่อมต่อไม่สำเร็จหรือหมดเวลา)")}
	}

	// Presence is determined by HTTP 200 only (Req 5.2, 5.3).
	if res.StatusCode != 200 {
		return []valueobject.Result{result(robotsTxtTitle, valueobject.StatusFail,
			fmt.Sprintf("ไม่พบไฟล์ robots.txt บนเว็บไซต์ (สถานะ %d)", res.StatusCode))}
	}

	// 200: an empty body passes; otherwise every meaningful line must be valid
	// syntax (Req 5.2, 5.4).
	if bad, ok := firstInvalidRobotsLine(string(res.Body)); !ok {
		return []valueobject.Result{result(robotsTxtTitle, valueobject.StatusFail,
			fmt.Sprintf("พบไฟล์ robots.txt แต่ไวยากรณ์ไม่ถูกต้อง: %q", bad))}
	}

	// Present and syntactically valid, but a robots.txt that blocks the whole
	// site for the catch-all crawler prevents Google's bot from reaching any
	// content — that is a functional failure, not a pass.
	if robotsBlocksEntireSite(string(res.Body)) {
		return []valueobject.Result{result(robotsTxtTitle, valueobject.StatusFail,
			"robots.txt บล็อกการเข้าถึงทั้งเว็บไซต์ (Disallow: /) ทำให้บอทของ Google ไม่สามารถเข้าถึงเนื้อหาได้")}
	}

	return []valueobject.Result{result(robotsTxtTitle, valueobject.StatusPass,
		"พบไฟล์ robots.txt และรูปแบบถูกต้อง")}
}

// firstInvalidRobotsLine scans the robots.txt body and returns the first line
// that is neither blank, a comment (starting with '#'), nor a valid
// "<known directive>: value" pair. It normalizes CRLF to LF first. When every
// line is acceptable it returns ("", true); the value after ':' may be empty
// (e.g. "Disallow:") (Req 5.4).
func firstInvalidRobotsLine(body string) (string, bool) {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	for _, line := range strings.Split(normalized, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		idx := strings.Index(trimmed, ":")
		if idx < 0 {
			return line, false
		}

		field := strings.ToLower(strings.TrimSpace(trimmed[:idx]))
		if _, known := knownRobotsDirectives[field]; !known {
			return line, false
		}
		// The value after ':' may be empty — that is valid.
	}
	return "", true
}

// robotsGroup is a single record in robots.txt: one or more consecutive
// User-agent lines followed by the Allow/Disallow rules that apply to them,
// until the next User-agent line begins a new group.
type robotsGroup struct {
	// userAgents holds the trimmed, lower-cased user-agent tokens of the group.
	userAgents []string
	// disallowRoot is true when the group contains a "Disallow: /" rule (path
	// value exactly "/" after trimming).
	disallowRoot bool
	// allowRoot is true when the group contains an "Allow: /" rule (path value
	// exactly "/" after trimming), which overrides a root disallow.
	allowRoot bool
}

// stripInlineComment removes an inline "#..." comment from a directive value
// and trims surrounding whitespace, mirroring how robots.txt parsers treat '#'
// as a comment start anywhere on a line. "/   # block" → "/"; "*  # bots" → "*".
func stripInlineComment(value string) string {
	if i := strings.IndexByte(value, '#'); i >= 0 {
		value = value[:i]
	}
	return strings.TrimSpace(value)
}

// robotsBlocksEntireSite reports whether the catch-all group (User-agent: *)
// effectively blocks the whole site. The site is blocked when that group has a
// "Disallow: /" that is NOT overridden by an "Allow: /" in the same group. When
// there is no catch-all group, the site is not considered blocked here.
func robotsBlocksEntireSite(body string) bool {
	for _, g := range parseRobotsGroups(body) {
		if !g.appliesToAll() {
			continue
		}
		if g.disallowRoot && !g.allowRoot {
			return true
		}
	}
	return false
}

// appliesToAll reports whether the group targets the catch-all user agent "*".
func (g robotsGroup) appliesToAll() bool {
	for _, ua := range g.userAgents {
		if ua == "*" {
			return true
		}
	}
	return false
}

// parseRobotsGroups splits a syntactically valid robots.txt body into
// user-agent groups. A group begins at one or more consecutive User-agent lines
// and absorbs the Allow/Disallow lines that follow until the next User-agent
// line starts a new group. Field names are matched case-insensitively and
// values are compared after TrimSpace, mirroring firstInvalidRobotsLine. Blank
// and comment lines are ignored; non-grouping directives (e.g. Sitemap,
// Crawl-delay) do not affect block detection.
func parseRobotsGroups(body string) []robotsGroup {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	var groups []robotsGroup
	var current *robotsGroup
	// prevWasUserAgent tracks whether the previous meaningful line was a
	// User-agent line, so consecutive User-agent lines stack into one group.
	prevWasUserAgent := false

	for _, line := range strings.Split(normalized, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		idx := strings.Index(trimmed, ":")
		if idx < 0 {
			continue
		}
		field := strings.ToLower(strings.TrimSpace(trimmed[:idx]))
		// An inline "#..." starts a comment (Google's robots.txt parser treats
		// '#' as a comment start anywhere on a line), so strip it before
		// comparing the value. Without this, "Disallow: /   # block" would trim
		// to "/   # block" (≠ "/") and a genuinely site-blocking file with
		// inline comments would slip through as a pass.
		value := stripInlineComment(trimmed[idx+1:])

		switch field {
		case "user-agent":
			if !prevWasUserAgent {
				// Start of a new group.
				groups = append(groups, robotsGroup{})
				current = &groups[len(groups)-1]
			}
			current.userAgents = append(current.userAgents, strings.ToLower(value))
			prevWasUserAgent = true
		case "disallow":
			if current != nil && value == "/" {
				current.disallowRoot = true
			}
			prevWasUserAgent = false
		case "allow":
			if current != nil && value == "/" {
				current.allowRoot = true
			}
			prevWasUserAgent = false
		default:
			prevWasUserAgent = false
		}
	}
	return groups
}
