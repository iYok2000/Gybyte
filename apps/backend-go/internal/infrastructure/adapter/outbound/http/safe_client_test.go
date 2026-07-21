// Integration tests สำหรับ SafeHTTPClient ด้วย httptest local server
// ยืนยัน: redirect-guard, TLS จริง, per-request timeout และเพดาน body 5 MiB
// — Requirements 2.7, 4.1, 4.5, 14.2
package http

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"monorepo/backend-go/internal/application/audit/validation"
)

// newTestClient สร้าง SafeHTTPClient แล้วผ่อนปรน dial-time IP guard ให้เชื่อมต่อ
// httptest บน loopback ได้ (โปรดักชันยังคงบล็อก loopback ผ่าน validator.GuardResolvedIPs)
// หมายเหตุ: redirect guard ยังใช้ validator.GuardSSRF จริง จึงยังบล็อก loopback ที่ปลาย redirect
func newTestClient() *SafeHTTPClient {
	c := NewSafeHTTPClient(validation.NewURLValidator())
	c.guardIPs = func(_ []net.IP) error { return nil }
	return c
}

// TestFetch_BodyCappedAt5MiB ยืนยันว่า Fetch อ่าน body ไม่เกิน maxResponseBodyBytes (5 MiB)
// แม้ต้นทางจะส่งข้อมูลใหญ่กว่านั้น — Req 14.2
func TestFetch_BodyCappedAt5MiB(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// ส่ง 6 MiB — เกินเพดาน 5 MiB
		_, _ = w.Write(make([]byte, 6<<20))
	}))
	defer srv.Close()

	client := newTestClient()
	res, err := client.Fetch(context.Background(), http.MethodGet, srv.URL, 10*time.Second)
	if err != nil {
		t.Fatalf("Fetch returned error: %v", err)
	}
	if len(res.Body) != maxResponseBodyBytes {
		t.Fatalf("expected body capped at %d bytes, got %d", maxResponseBodyBytes, len(res.Body))
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}
}

// TestFetch_PerRequestTimeout ยืนยันว่า timeout ต่อคำขอทำงานผ่าน context
// เมื่อต้นทางตอบช้ากว่ากำหนด Fetch ต้องคืน error — Req 4.5, 14.2
func TestFetch_PerRequestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	client := newTestClient()
	_, err := client.Fetch(context.Background(), http.MethodGet, srv.URL, 30*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

// TestFetch_RedirectGuardBlocksInternal ยืนยันว่า redirect guard บล็อกการติดตาม
// redirect ไปยังปลายทางภายใน (loopback) — Req 2.7
func TestFetch_RedirectGuardBlocksInternal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// redirect ไปยัง loopback ซึ่ง GuardSSRF ต้องปฏิเสธก่อนติดตาม
		http.Redirect(w, r, "http://127.0.0.1:9/internal", http.StatusFound)
	}))
	defer srv.Close()

	client := newTestClient()
	_, err := client.Fetch(context.Background(), http.MethodGet, srv.URL, 10*time.Second)
	if err == nil {
		t.Fatal("expected redirect to internal address to be blocked, got nil error")
	}
}

// TestFetchTLS_RealHandshakeSelfSigned ยืนยัน TLS handshake จริงต่อ httptest TLS server
// ที่ใช้ใบรับรอง self-signed → TLSInfo.Valid == false พร้อมเหตุผล (chain ไม่น่าเชื่อถือ) — Req 4.1
func TestFetchTLS_RealHandshakeSelfSigned(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// srv.URL = https://127.0.0.1:PORT — ตัด scheme ออกให้เหลือ host:port
	host := srv.Listener.Addr().String()

	client := newTestClient()
	info, err := client.FetchTLS(context.Background(), host, 10*time.Second)
	if err != nil {
		t.Fatalf("FetchTLS returned dial error: %v", err)
	}
	if info.Valid {
		t.Fatal("expected self-signed cert to be invalid, got Valid=true")
	}
	if info.Reason == "" {
		t.Fatal("expected a non-empty reason for invalid TLS")
	}
}
