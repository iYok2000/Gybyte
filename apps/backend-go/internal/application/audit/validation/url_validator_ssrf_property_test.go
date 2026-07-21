package validation

import (
	"math/rand"
	"net"
	"testing"
	"testing/quick"
)

// Feature: adready-checker, Property 3
//
// Property 3: การจำแนก IP/host แบบ SSRF ปฏิเสธปลายทางภายในทั้ง IPv4 และ IPv6
// ก่อนออกภายนอก และใช้ซ้ำกับ redirect
// Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.7

// ipCategory ระบุประเภทของ IP ที่จะสร้างในการทดสอบ พร้อมความคาดหมายว่าถูกบล็อกหรือไม่
type ipCategory int

const (
	catLoopbackV4 ipCategory = iota
	catPrivate10
	catPrivate172
	catPrivate192
	catLinkLocalV4
	catCGNAT
	catTestNet
	catClassE
	catUnspecifiedV4
	catMulticastV4
	catMetadataV4
	catLoopbackV6
	catPrivateV6
	catLinkLocalV6
	catMetadataV6
	catPublicV4
	catPublicV6
	numCategories
)

// fakeFetcher นับจำนวนครั้งที่มีการ "ยิงออกภายนอก" เพื่อยืนยันว่าไม่ถูกเรียกเมื่อถูกบล็อก
type fakeFetcher struct{ calls int }

func (f *fakeFetcher) fetch() { f.calls++ }

// publicV4Samples IP สาธารณะที่รับรองว่าไม่อยู่ในช่วงภายในใด ๆ
var publicV4Samples = []net.IP{
	net.ParseIP("8.8.8.8"),
	net.ParseIP("1.1.1.1"),
	net.ParseIP("93.184.216.34"),
	net.ParseIP("208.67.222.222"),
	net.ParseIP("151.101.1.140"),
}

// publicV6Samples IPv6 สาธารณะ (global unicast) ที่รับรองว่าไม่อยู่ในช่วงภายใน
var publicV6Samples = []net.IP{
	net.ParseIP("2606:4700:4700::1111"),
	net.ParseIP("2001:4860:4860::8888"),
	net.ParseIP("2620:fe::fe"),
}

// buildIP สร้าง IP ตามประเภทที่กำหนดโดยใช้ค่าสุ่มจาก rng และคืนความคาดหมายว่าถูกบล็อกหรือไม่
func buildIP(cat ipCategory, rng *rand.Rand) (net.IP, bool) {
	b := func() byte { return byte(rng.Intn(256)) }
	switch cat {
	case catLoopbackV4:
		return net.IPv4(127, b(), b(), b()), true
	case catPrivate10:
		return net.IPv4(10, b(), b(), b()), true
	case catPrivate172:
		return net.IPv4(172, byte(16+rng.Intn(16)), b(), b()), true
	case catPrivate192:
		return net.IPv4(192, 168, b(), b()), true
	case catLinkLocalV4:
		return net.IPv4(169, 254, b(), b()), true
	case catCGNAT:
		return net.IPv4(100, byte(64+rng.Intn(64)), b(), b()), true
	case catTestNet:
		nets := [][3]byte{{192, 0, 2}, {198, 51, 100}, {203, 0, 113}}
		p := nets[rng.Intn(len(nets))]
		return net.IPv4(p[0], p[1], p[2], b()), true
	case catClassE:
		return net.IPv4(byte(240+rng.Intn(16)), b(), b(), b()), true
	case catUnspecifiedV4:
		return net.IPv4(0, 0, 0, 0), true
	case catMulticastV4:
		return net.IPv4(byte(224+rng.Intn(16)), b(), b(), b()), true
	case catMetadataV4:
		return net.ParseIP("169.254.169.254"), true
	case catLoopbackV6:
		return net.ParseIP("::1"), true
	case catPrivateV6:
		// fd00::/8 (unique local) — IsPrivate == true
		ip := make(net.IP, net.IPv6len)
		ip[0] = 0xfd
		for i := 1; i < net.IPv6len; i++ {
			ip[i] = b()
		}
		return ip, true
	case catLinkLocalV6:
		// fe80::/10
		ip := make(net.IP, net.IPv6len)
		ip[0] = 0xfe
		ip[1] = 0x80
		return ip, true
	case catMetadataV6:
		return net.ParseIP("fd00:ec2::254"), true
	case catPublicV4:
		return publicV4Samples[rng.Intn(len(publicV4Samples))], false
	case catPublicV6:
		return publicV6Samples[rng.Intn(len(publicV6Samples))], false
	default:
		return net.ParseIP("8.8.8.8"), false
	}
}

func TestProperty3_SSRFClassification(t *testing.T) {
	v := NewURLValidator()

	// สำหรับทุกประเภท IP: การจำแนกต้องตรงกับความคาดหมาย และเมื่อถูกบล็อก
	// ต้องไม่มีการยิงออกภายนอก (fakeFetcher ไม่ถูกเรียก)
	assertion := func(seed int64, catSel uint8) bool {
		rng := rand.New(rand.NewSource(seed))
		cat := ipCategory(int(catSel) % int(numCategories))
		ip, expectBlocked := buildIP(cat, rng)

		// isBlockedIP ต้องจำแนกตรงกับความคาดหมาย
		if isBlockedIP(ip) != expectBlocked {
			return false
		}

		// GuardResolvedIPs ต้องคืน error ก็ต่อเมื่อถูกบล็อก
		err := v.GuardResolvedIPs([]net.IP{ip})
		if (err != nil) != expectBlocked {
			return false
		}

		// จำลอง pipeline: ยิงออกภายนอกเฉพาะเมื่อ guard ผ่านเท่านั้น
		fetcher := &fakeFetcher{}
		if err == nil {
			fetcher.fetch()
		}
		if expectBlocked && fetcher.calls != 0 {
			return false // ถูกบล็อกแต่กลับยิงออกภายนอก — ผิด
		}
		if !expectBlocked && fetcher.calls != 1 {
			return false
		}
		return true
	}

	if err := quick.Check(assertion, &quick.Config{MaxCount: 500}); err != nil {
		t.Fatalf("Property 3 (SSRF classification) failed: %v", err)
	}
}

// Feature: adready-checker, Property 3
// ยืนยันเสริม: redirect guard ใช้เกณฑ์ SSRF เดียวกัน — ปฏิเสธ IP literal ภายในทุก hop
func TestProperty3_RedirectGuardReusesSSRF(t *testing.T) {
	v := NewURLValidator()
	guard := v.NewRedirectGuard()

	assertion := func(seed int64, catSel uint8) bool {
		rng := rand.New(rand.NewSource(seed))
		cat := ipCategory(int(catSel) % int(numCategories))
		ip, expectBlocked := buildIP(cat, rng)

		host := ip.String()
		if ip.To4() == nil {
			host = "[" + ip.String() + "]"
		}
		req := newGETRequest(t, "https://"+host+"/")
		err := guard(req, nil)

		// redirect ไปยังปลายทางภายในต้องถูกปฏิเสธ (error != nil) เช่นเดียวกับ GuardSSRF
		return (err != nil) == expectBlocked
	}

	if err := quick.Check(assertion, &quick.Config{MaxCount: 300}); err != nil {
		t.Fatalf("Property 3 (redirect guard reuse) failed: %v", err)
	}
}
