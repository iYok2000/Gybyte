package check

import (
	"context"
	"time"

	"monorepo/backend-go/internal/core/domain/audit/port"
	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// httpsTimeout bounds the TLS handshake / HTTPS probe (Req 4.1).
const httpsTimeout = 30 * time.Second

// httpsTitle is the human-readable name of this check.
const httpsTitle = "HTTPS"

// HTTPSCheck verifies that the target is served over HTTPS with a valid TLS
// certificate (Req 4.1–4.5).
type HTTPSCheck struct{}

// NewHTTPSCheck constructs the HTTPS readiness check.
func NewHTTPSCheck() HTTPSCheck { return HTTPSCheck{} }

// Title returns the check name.
func (HTTPSCheck) Title() string { return httpsTitle }

// Evaluate performs a TLS handshake against the target host and maps the
// outcome to a single Audit_Result. It never returns an error (Req 14.2).
//
// A certificate is considered valid only when the trust chain verifies, the
// hostname matches, the certificate is within its validity window, and it is
// not revoked (revocation is best-effort — Req 4.1). The Fetcher reports this
// via TLSInfo.Valid. Pass iff Valid == true (Req 4.2); HTTP-only, connection
// failure, and timeout all map to fail with a reason (Req 4.3, 4.4, 4.5).
func (c HTTPSCheck) Evaluate(ctx context.Context, target Target, f port.Fetcher) []valueobject.Result {
	tls, err := f.FetchTLS(ctx, target.Host, httpsTimeout)
	if err != nil || tls == nil {
		// Could not establish an HTTPS connection or timed out (Req 4.5).
		return []valueobject.Result{result(httpsTitle, valueobject.StatusFail,
			"ไม่สามารถเชื่อมต่อผ่าน HTTPS ได้ (การเชื่อมต่อล้มเหลวหรือหมดเวลา)")}
	}

	if tls.Valid {
		return []valueobject.Result{result(httpsTitle, valueobject.StatusPass,
			"เว็บไซต์ให้บริการผ่าน HTTPS ด้วยใบรับรอง TLS ที่ถูกต้อง")}
	}

	// TLS invalid or HTTP-only — report the detected reason (Req 4.3, 4.4).
	msg := "เว็บไซต์ไม่ผ่านการตรวจสอบ HTTPS: ใบรับรอง TLS ไม่ถูกต้องหรือให้บริการผ่าน HTTP เท่านั้น"
	if tls.Reason != "" {
		msg = "เว็บไซต์ไม่ผ่านการตรวจสอบ HTTPS: " + tls.Reason
	}
	return []valueobject.Result{result(httpsTitle, valueobject.StatusFail, msg)}
}
