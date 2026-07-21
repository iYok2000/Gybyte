package check

import (
	"context"
	"fmt"
	"strings"
	"time"

	"monorepo/backend-go/internal/core/domain/audit/port"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// adsTxtTimeout bounds the GET /ads.txt request (Req 3.1).
const adsTxtTimeout = 10 * time.Second

// adsTxtTitle is the human-readable name of this check.
const adsTxtTitle = "ads.txt"

// AdsTxtCheck verifies that an ads.txt file is present and non-empty at the
// target site root (Req 3.1–3.5).
type AdsTxtCheck struct{}

// NewAdsTxtCheck constructs the ads.txt readiness check.
func NewAdsTxtCheck() AdsTxtCheck { return AdsTxtCheck{} }

// Title returns the check name.
func (AdsTxtCheck) Title() string { return adsTxtTitle }

// Evaluate fetches /ads.txt from the target origin and maps the outcome to a
// single Audit_Result. It never returns an error: fetch failures, non-200
// statuses, and empty content are all encoded as fail results (Req 14.2).
//
// Pass criteria (Req 3.2): HTTP 200 AND the body, after trimming leading/
// trailing whitespace, is at least 1 character long.
func (c AdsTxtCheck) Evaluate(ctx context.Context, target Target, f port.Fetcher) []valueobject.Result {
	url := originURL(target, adsTxtPath)

	res, err := f.Fetch(ctx, "GET", url, adsTxtTimeout)
	if err != nil || res == nil {
		// Connection failed or timed out — resource unreachable (Req 3.5).
		return []valueobject.Result{result(adsTxtTitle, valueobject.StatusFail,
			"ไม่สามารถเข้าถึงไฟล์ ads.txt ได้ (เชื่อมต่อไม่สำเร็จหรือหมดเวลา)")}
	}

	// Non-200 status → fail, naming the received status code (Req 3.3).
	if res.StatusCode != 200 {
		return []valueobject.Result{result(adsTxtTitle, valueobject.StatusFail,
			fmt.Sprintf("ไม่พบไฟล์ ads.txt บนเว็บไซต์ (สถานะ %d)", res.StatusCode))}
	}

	// 200 but empty content after trimming whitespace → fail (Req 3.4).
	if len(strings.TrimSpace(string(res.Body))) < 1 {
		return []valueobject.Result{result(adsTxtTitle, valueobject.StatusFail,
			"พบไฟล์ ads.txt แต่เนื้อหาว่างเปล่า")}
	}

	// 200 and non-empty content → pass (Req 3.2).
	return []valueobject.Result{result(adsTxtTitle, valueobject.StatusPass,
		"พบไฟล์ ads.txt และมีเนื้อหา")}
}

// result builds an Audit_Result, falling back to a fail result if the status
// invariant is somehow violated (defensive — checks only pass pass/fail here).
func result(title string, status valueobject.Status, message string) valueobject.Result {
	r, err := valueobject.NewResult(title, status, message)
	if err != nil {
		return valueobject.Result{Title: title, Status: valueobject.StatusFail, Message: message}
	}
	return r
}
