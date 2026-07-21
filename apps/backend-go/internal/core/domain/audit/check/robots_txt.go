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
