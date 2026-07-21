// Package validation ตรวจสอบและกรองความถูกต้อง/ความปลอดภัยของ URL ที่ผู้ใช้ส่งเข้ามา
// ป้องกัน SSRF (Server-Side Request Forgery) และ path traversal ก่อนดำเนินการร้องขอออกภายนอก
// — Requirements 1.6, 2.1–2.8
package validation

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	apperrors "monorepo/backend-go/pkg/errors"
)

// ค่าคงที่เชิงนโยบายของ URL_Validator (ถอดจาก design.md — ห้ามพึ่งพา FE validation)
const (
	// MaxURLLength ความยาวสูงสุดของ URL ที่ยอมรับ นับเป็นจำนวน rune (ไม่ใช่ byte) — Req 2.1
	MaxURLLength = 2048
	// dnsResolveTimeout งบเวลาสูงสุดสำหรับการแปลงค่าชื่อโฮสต์ (DNS) — Req 2.8
	dnsResolveTimeout = 5 * time.Second
	// maxRedirects จำนวน redirect hop สูงสุดที่อนุญาตให้ติดตาม — Req 2.7
	maxRedirects = 10
)

// พาธคงที่ที่กำหนดล่วงหน้าสำหรับประกอบ URL ทรัพยากร (กัน path traversal — Req 2.6)
// ต้องสอดคล้องกับพาธที่ใช้ในเลเยอร์ check (ads.txt / robots.txt / homepage)
const (
	adsTxtPath    = "/ads.txt"
	robotsTxtPath = "/robots.txt"
	homepagePath  = "/"
)

// allowedSchemes สคีมที่อนุญาตเท่านั้น — Req 2.1, 2.2
var allowedSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

// blockedHostnames ชื่อโฮสต์ที่ปฏิเสธทันทีโดยไม่ต้อง resolve — Req 2.4
// ครอบคลุม localhost และ metadata endpoint ของผู้ให้บริการคลาวด์
var blockedHostnames = map[string]bool{
	"localhost":                true,
	"metadata.google.internal": true,
	"metadata":                 true,
}

// URLValidator เป็น stateless validator ใช้ร่วมระหว่าง input validation และ dial-time SSRF guard
type URLValidator struct{}

// NewURLValidator สร้างอินสแตนซ์ของ URLValidator
func NewURLValidator() *URLValidator {
	return &URLValidator{}
}

// ValidateAndNormalize ตัดอักขระควบคุม + trim → ปฏิเสธค่าว่าง → นับ rune ≤ 2048
// → parse → บังคับสคีม http/https (lowercase) → ต้องมี host; error เป็น *ValidationError (→ 400)
// — Requirements 1.6, 2.1, 2.2, 2.5
func (v *URLValidator) ValidateAndNormalize(raw string) (*url.URL, error) {
	// 1) ลบอักขระควบคุมและตัดช่องว่างหัวท้ายก่อนตรวจสอบและก่อนนับความยาว — Req 1.6
	cleaned := strings.TrimSpace(stripControlChars(raw))

	// 2) ปฏิเสธค่าว่างเปล่า
	if cleaned == "" {
		return nil, apperrors.NewValidationError("url", "URL ต้องไม่ว่างเปล่า")
	}

	// 3) นับความยาวเป็นจำนวน "rune" (ไม่ใช่ byte) เทียบกับขีดจำกัด 2048 — Req 2.1, 2.2
	if utf8.RuneCountInString(cleaned) > MaxURLLength {
		return nil, apperrors.NewValidationError("url", "URL ยาวเกินขีดจำกัด 2048 อักขระ")
	}

	// 4) parse รูปแบบ URL
	parsed, err := url.Parse(cleaned)
	if err != nil {
		return nil, apperrors.NewValidationError("url", "รูปแบบ URL ไม่ถูกต้อง")
	}

	// 5) บังคับสคีมอยู่ในชุดที่อนุญาต (เปรียบเทียบแบบ lowercase) — Req 2.2
	scheme := strings.ToLower(parsed.Scheme)
	if !allowedSchemes[scheme] {
		return nil, apperrors.NewValidationError("url", "รองรับเฉพาะ URL ที่ใช้สคีม http หรือ https")
	}
	parsed.Scheme = scheme

	// 6) ต้องมีชื่อโฮสต์ (กัน "http:///path" หรือ URL ที่ไม่มี host)
	if parsed.Host == "" || parsed.Hostname() == "" {
		return nil, apperrors.NewValidationError("url", "URL ต้องระบุชื่อโฮสต์")
	}

	return parsed, nil
}

// stripControlChars ลบอักขระควบคุม (unicode.IsControl) ทั้งหมดออกจากสตริง — Req 1.6
func stripControlChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1 // drop rune
		}
		return r
	}, s)
}

// GuardSSRF ตรวจสอบก่อนดำเนินการร้องขอออกภายนอก:
// ปฏิเสธ blocked hostname → ถ้าเป็น IP literal ตรวจตรง → มิฉะนั้น resolve ภายใน 5s แล้วตรวจ "ทุก" IP
// error เป็น *BadRequestError (→ 400) — Requirements 2.3, 2.4, 2.8
func (v *URLValidator) GuardSSRF(ctx context.Context, u *url.URL) error {
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return apperrors.NewBadRequestError("ปลายทางไม่ได้รับอนุญาต")
	}

	// ปฏิเสธชื่อโฮสต์ต้องห้ามทันที (localhost / cloud metadata) — Req 2.4
	if blockedHostnames[host] {
		return apperrors.NewBadRequestError("ปลายทางไม่ได้รับอนุญาต")
	}

	// ถ้า host เป็น IP literal อยู่แล้ว ตรวจโดยตรง (ไม่ต้อง resolve)
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return apperrors.NewBadRequestError("ปลายทางเป็นที่อยู่ภายในที่ไม่ได้รับอนุญาต")
		}
		return nil
	}

	// มิฉะนั้น resolve ชื่อโฮสต์ภายในงบเวลา 5s แล้วตรวจทุก IP ที่ได้ — Req 2.3, 2.8
	resolveCtx, cancel := context.WithTimeout(ctx, dnsResolveTimeout)
	defer cancel()

	var resolver net.Resolver
	addrs, err := resolver.LookupIPAddr(resolveCtx, host)
	if err != nil {
		// resolve ล้มเหลว/หมดเวลา → ปฏิเสธโดยไม่ยิงออกภายนอก — Req 2.8
		return apperrors.NewBadRequestError("ไม่สามารถแปลงค่าชื่อโฮสต์ได้ภายในเวลาที่กำหนด")
	}

	ips := make([]net.IP, 0, len(addrs))
	for _, a := range addrs {
		ips = append(ips, a.IP)
	}
	return v.GuardResolvedIPs(ips)
}

// GuardResolvedIPs ปฏิเสธถ้ามี IP ใด ๆ ที่ถูกจำแนกว่าเป็นที่อยู่ภายใน — Requirements 2.3, 2.4
func (v *URLValidator) GuardResolvedIPs(ips []net.IP) error {
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return apperrors.NewBadRequestError("ปลายทางเป็นที่อยู่ภายในที่ไม่ได้รับอนุญาต")
		}
	}
	return nil
}

// NewRedirectGuard คืน callback สำหรับ http.Client.CheckRedirect:
// จำกัดจำนวน redirect + re-validate scheme + GuardSSRF ทุก hop — Requirement 2.7
func (v *URLValidator) NewRedirectGuard() func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		// cap จำนวน redirect hop
		if len(via) >= maxRedirects {
			return apperrors.NewBadRequestError("จำนวนการเปลี่ยนเส้นทาง (redirect) เกินขีดจำกัด")
		}
		// re-validate สคีมของปลายทาง redirect
		if !allowedSchemes[strings.ToLower(req.URL.Scheme)] {
			return apperrors.NewBadRequestError("ปลายทางของการเปลี่ยนเส้นทางไม่ได้รับอนุญาต")
		}
		// ตรวจ SSRF ปลายทาง redirect ด้วยเกณฑ์เดียวกันก่อนติดตาม
		if err := v.GuardSSRF(req.Context(), req.URL); err != nil {
			return err
		}
		return nil
	}
}

// --- IP classification (SSRF) ---

var (
	cloudMetadataIPv4 = net.ParseIP("169.254.169.254")
	cloudMetadataIPv6 = net.ParseIP("fd00:ec2::254")
)

// reservedExtraCIDRs ช่วงที่อยู่สงวนเพิ่มเติมที่ net.IP ไม่มี helper ให้ตรงตัว
// CGNAT (100.64.0.0/10), TEST-NET-1/2/3 และ class-E (240.0.0.0/4)
var reservedExtraCIDRs = mustParseCIDRs(
	"100.64.0.0/10",   // Carrier-Grade NAT (RFC 6598)
	"192.0.2.0/24",    // TEST-NET-1 (RFC 5737)
	"198.51.100.0/24", // TEST-NET-2 (RFC 5737)
	"203.0.113.0/24",  // TEST-NET-3 (RFC 5737)
	"240.0.0.0/4",     // reserved / class-E (RFC 1112)
)

func mustParseCIDRs(cidrs ...string) []*net.IPNet {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		if _, n, err := net.ParseCIDR(c); err == nil {
			nets = append(nets, n)
		}
	}
	return nets
}

// isBlockedIP จำแนกว่า IP (ทั้ง IPv4 และ IPv6) เป็นปลายทางภายในที่ไม่อนุญาตหรือไม่ — Req 2.3, 2.4
func isBlockedIP(ip net.IP) bool {
	// nil = แปลงค่าไม่ได้ → บล็อกเพื่อความปลอดภัย
	if ip == nil {
		return true
	}
	// loopback / private / link-local / unspecified / multicast — Req 2.3
	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() ||
		ip.IsMulticast() {
		return true
	}
	// cloud metadata endpoint — Req 2.4
	if isCloudMetadata(ip) {
		return true
	}
	// ช่วงสงวนเพิ่มเติม (CGNAT / TEST-NET / class-E) — Req 2.3
	return isReservedExtra(ip)
}

func isCloudMetadata(ip net.IP) bool {
	return ip.Equal(cloudMetadataIPv4) || ip.Equal(cloudMetadataIPv6)
}

func isReservedExtra(ip net.IP) bool {
	for _, n := range reservedExtraCIDRs {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// --- Resource path assembly (กัน path traversal — Req 2.6) ---

// OriginOf คืน origin ของ URL ในรูป "scheme://host" โดยทิ้ง path/query/fragment ที่ผู้ใช้ควบคุมได้
func OriginOf(u *url.URL) string {
	return u.Scheme + "://" + u.Host
}

// ResourceURL ประกอบ URL ทรัพยากรจาก origin ต่อด้วย "พาธคงที่" ที่กำหนดล่วงหน้าเท่านั้น
// ไม่นำส่วนของพาธที่ผู้ใช้ควบคุมได้มาต่อ — กัน path traversal (Req 2.6)
func ResourceURL(u *url.URL, fixedPath string) string {
	return OriginOf(u) + fixedPath
}
