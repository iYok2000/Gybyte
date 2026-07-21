package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/quick"
	"time"

	"github.com/gin-gonic/gin"
)

// TestRateLimiterProperty13 ตรวจสอบคุณสมบัติของ audit rate limiter
//
// Feature: adready-checker, Property 13
// Property 13: Rate limiter อนุญาต min(N,60) คำขอต่อกรอบ 60 วินาที;
// เมื่อถูกปฏิเสธ retryAfter เป็นจำนวนเต็มบวก ≤ 60; และรีเซ็ตตัวนับเมื่อครบกรอบ
// Validates: Requirements 9.1, 9.2, 9.3, 9.4
func TestRateLimiterProperty13(t *testing.T) {
	property := func(seed uint8) bool {
		rl := NewRateLimiter(60, 60*time.Second)
		n := int(seed) // จำนวนคำขอ 0..255 ภายในกรอบเดียว
		const ip = "203.0.113.7"

		allowedCount := 0
		for i := 0; i < n; i++ {
			allowed, retryAfter := rl.allowWithRetry(ip)
			if allowed {
				allowedCount++
				// คำขอที่ผ่านต้องไม่มี retryAfter (Req 9.2)
				if retryAfter != 0 {
					return false
				}
			} else {
				// คำขอที่ถูกปฏิเสธต้องมี retryAfter เป็นจำนวนเต็มบวก ≤ 60 (Req 9.3)
				if retryAfter < 1 || retryAfter > 60 {
					return false
				}
			}
		}

		// จำนวนที่อนุญาตต้องเท่ากับ min(N, 60) พอดี (Req 9.1)
		expected := n
		if expected > 60 {
			expected = 60
		}
		if allowedCount != expected {
			return false
		}

		// จำลองการพ้นกรอบเวลา → ตัวนับต้องรีเซ็ตและอนุญาตคำขอได้อีกครั้ง (Req 9.4)
		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if exists {
			v.lastSeen = time.Now().Add(-61 * time.Second)
		}
		rl.mu.Unlock()
		if exists {
			allowed, retryAfter := rl.allowWithRetry(ip)
			if !allowed || retryAfter != 0 {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 200}); err != nil {
		t.Errorf("Property 13 failed: %v", err)
	}
}

// TestLimitGinWithRetryDoesNotCallEngineOn429 ยืนยันว่าเมื่อเกินโควตา
// middleware ตอบ 429 พร้อม header Retry-After + body ที่มี retryAfter และ abort
// โดยไม่เรียก handler ปลายทาง (ไม่เรียก engine) — Req 9.2, 9.3
func TestLimitGinWithRetryDoesNotCallEngineOn429(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// limiter ความจุ 1 คำขอต่อกรอบ เพื่อให้คำขอที่สองถูกปฏิเสธแน่นอน
	rl := NewRateLimiter(1, 60*time.Second)

	engineCalls := 0
	router := gin.New()
	router.POST("/api/audit", rl.LimitGinWithRetry(), func(c *gin.Context) {
		engineCalls++
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	doRequest := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/api/audit", nil)
		req.RemoteAddr = "198.51.100.9:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	// คำขอแรกผ่าน → เรียก engine
	if w := doRequest(); w.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", w.Code)
	}

	// คำขอที่สองถูกปฏิเสธ → 429, ไม่เรียก engine
	w := doRequest()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", w.Code)
	}
	if got := w.Header().Get("Retry-After"); got == "" {
		t.Error("expected Retry-After header to be set on 429")
	}
	if engineCalls != 1 {
		t.Errorf("engine should be called exactly once (only the allowed request), got %d", engineCalls)
	}

	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code       string `json:"code"`
			Message    string `json:"message"`
			RetryAfter int    `json:"retryAfter"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse 429 body: %v", err)
	}
	if body.Success {
		t.Error("expected success=false on 429 body")
	}
	if body.Error.Code != "RATE_LIMIT_EXCEEDED" {
		t.Errorf("expected error code RATE_LIMIT_EXCEEDED, got %q", body.Error.Code)
	}
	if body.Error.RetryAfter < 1 || body.Error.RetryAfter > 60 {
		t.Errorf("expected retryAfter in [1,60], got %d", body.Error.RetryAfter)
	}
}
