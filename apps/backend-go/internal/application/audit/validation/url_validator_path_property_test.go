package validation

import (
	"context"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"testing/quick"
)

// newGETRequest เป็น helper สร้าง *http.Request สำหรับทดสอบ redirect guard
func newGETRequest(t *testing.T, rawURL string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, nil)
	if err != nil {
		t.Fatalf("failed to build request for %q: %v", rawURL, err)
	}
	return req
}

// Feature: adready-checker, Property 4
//
// Property 4: ประกอบพาธทรัพยากรจาก origin + พาธคงที่เท่านั้น (กัน path traversal)
// Validates: Requirements 2.6

// dangerousPaths ตัวอย่างพาธที่ผู้ใช้ควบคุมได้ซึ่งอาจเปิดช่องให้เกิด path traversal
var dangerousPaths = []string{
	"/../../etc/passwd",
	"/foo/../../bar",
	"/%2e%2e/%2e%2e/secret",
	"/legit/path",
	"/..;/admin",
	"/deep/nested/../../../../root",
	"",
	"/",
}

var dangerousQueries = []string{
	"a=1&b=../../x",
	"redirect=/../../evil",
	"",
	"q=%2e%2e%2f",
}

var dangerousFragments = []string{
	"section",
	"../../frag",
	"",
}

var fixedPaths = []string{adsTxtPath, robotsTxtPath, homepagePath}

// randomHost สร้าง hostname ที่ถูกต้องแบบสุ่มจาก rng
func randomHost(rng *rand.Rand) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	labels := 1 + rng.Intn(3)
	parts := make([]string, labels)
	for i := range parts {
		n := 1 + rng.Intn(8)
		var sb strings.Builder
		for j := 0; j < n; j++ {
			sb.WriteByte(letters[rng.Intn(len(letters))])
		}
		parts[i] = sb.String()
	}
	return strings.Join(parts, ".") + ".com"
}

func TestProperty4_ResourcePathAssembly(t *testing.T) {
	// For any URL ที่ผู้ใช้ส่งมา (รวม path/query/fragment อันตราย):
	// URL ทรัพยากรที่ประกอบได้ต้องเท่ากับ origin + พาธคงที่เท่านั้น
	// และต้องไม่มีส่วนของพาธที่ผู้ใช้ควบคุมได้ (เช่น "..") หลุดเข้ามา
	assertion := func(seed int64, schemeSel bool, pSel, qSel, fSel uint8) bool {
		rng := rand.New(rand.NewSource(seed))

		scheme := "https"
		if schemeSel {
			scheme = "http"
		}
		host := randomHost(rng)
		userPath := dangerousPaths[int(pSel)%len(dangerousPaths)]
		userQuery := dangerousQueries[int(qSel)%len(dangerousQueries)]
		userFrag := dangerousFragments[int(fSel)%len(dangerousFragments)]

		u := &url.URL{
			Scheme:   scheme,
			Host:     host,
			Path:     userPath,
			RawQuery: userQuery,
			Fragment: userFrag,
		}

		expectedOrigin := scheme + "://" + host
		if OriginOf(u) != expectedOrigin {
			return false
		}

		for _, fp := range fixedPaths {
			got := ResourceURL(u, fp)
			want := expectedOrigin + fp

			// ต้องเท่ากับ origin + พาธคงที่พอดี
			if got != want {
				return false
			}
			// ต้องไม่มี path traversal token หลุดเข้ามา
			if strings.Contains(got, "..") {
				return false
			}
			// ต้องไม่มี query/fragment ของผู้ใช้ปนมา
			if userQuery != "" && strings.Contains(got, userQuery) {
				return false
			}
			if userFrag != "" && strings.Contains(got, "#"+userFrag) {
				return false
			}
		}
		return true
	}

	if err := quick.Check(assertion, &quick.Config{MaxCount: 300}); err != nil {
		t.Fatalf("Property 4 (resource path assembly) failed: %v", err)
	}
}
