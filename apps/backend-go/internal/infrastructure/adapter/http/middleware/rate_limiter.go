package middleware

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// visitor เก็บสถานะ token bucket ของผู้เรียกแต่ละราย (ระบุด้วย IP)
type visitor struct {
	tokens   float64   // จำนวน token ที่เหลือในถัง
	lastSeen time.Time // เวลาเข้าใช้งานล่าสุด (ใช้สำหรับ refill + cleanup)
}

// RateLimiter เป็น token bucket ต่อ IP แบบ in-memory
//
// - rate   = จำนวนคำขอสูงสุดที่อนุญาตภายในหนึ่ง window (ความจุถัง)
// - window = ความยาวของกรอบเวลาที่ใช้ refill token จนเต็มถัง
//
// มี goroutine เบื้องหลัง (cleanupVisitors) ลบ visitor ที่ไม่ active เกิน 3 นาที
// เพื่อกันหน่วยความจำโตไม่จำกัด — Req 9.1, 9.4
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     float64
	window   time.Duration
}

// NewRateLimiter สร้าง rate limiter พร้อม token bucket และเริ่ม goroutine cleanup
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     float64(rate),
		window:   window,
	}
	go rl.cleanupVisitors()
	return rl
}

// cleanupVisitors รันทุก 1 นาที ลบ visitor ที่ไม่ได้ใช้งานเกิน 3 นาที
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// refillLocked เติม token ตามเวลาที่ผ่านไปนับจาก lastSeen (ต้องถือ lock อยู่แล้ว)
// อัตราการเติม = rate token ต่อหนึ่ง window; ถังไม่เกินความจุ rate
func (rl *RateLimiter) refillLocked(v *visitor, now time.Time) {
	elapsed := now.Sub(v.lastSeen)
	if elapsed <= 0 {
		return
	}
	refill := (float64(elapsed) / float64(rl.window)) * rl.rate
	v.tokens = math.Min(rl.rate, v.tokens+refill)
}

// allow ตรวจว่าผู้เรียก ip นี้ยังส่งคำขอได้หรือไม่ (token bucket แบบ refill ต่อเนื่อง)
// คืน true พร้อมหัก 1 token เมื่ออนุญาต; false เมื่อ token ไม่พอ — แบบแผนเดิม
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, ok := rl.visitors[ip]
	if !ok {
		// ผู้เรียกใหม่: เริ่มถังเต็มแล้วหัก 1 token สำหรับคำขอนี้
		rl.visitors[ip] = &visitor{tokens: rl.rate - 1, lastSeen: now}
		return true
	}

	rl.refillLocked(v, now)
	v.lastSeen = now
	if v.tokens >= 1 {
		v.tokens--
		return true
	}
	return false
}

// Limit เป็น middleware สำหรับ net/http stdlib (แบบแผนเดิม)
// ใช้ getClientIP (X-Forwarded-For → X-Real-IP → RemoteAddr) เพื่อจำแนกผู้เรียก
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !rl.allow(ip) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// LimitGin เป็น middleware สำหรับ Gin (แบบแผนเดิม) ใช้ c.ClientIP()
func (rl *RateLimiter) LimitGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.allow(c.ClientIP()) {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}

// NewLoginRateLimiter จำกัด 5 คำขอ ต่อ 15 นาที (สำหรับ endpoint ล็อกอิน) — แบบแผนเดิม
func NewLoginRateLimiter() *RateLimiter {
	return NewRateLimiter(5, 15*time.Minute)
}

// NewAPIRateLimiter จำกัด 100 คำขอ ต่อ 1 นาที (สำหรับ API ทั่วไป) — แบบแผนเดิม
func NewAPIRateLimiter() *RateLimiter {
	return NewRateLimiter(100, time.Minute)
}

// --- ส่วนเพิ่มเติมสำหรับ AdReady audit (additive; ไม่แตะ allow/LimitGin เดิม) ---

// NewAuditRateLimiter จำกัด 60 คำขอ ต่อ 60 วินาที ต่อผู้เรียกแต่ละราย — Req 9.1
func NewAuditRateLimiter() *RateLimiter {
	return NewRateLimiter(60, 60*time.Second)
}

// allowWithRetry ทำงานแบบ fixed-window ต่อ IP สำหรับ endpoint audit:
//   - refill ถังจนเต็มเมื่อพ้นกรอบเวลา (elapsed >= window) แล้วรีเซ็ตตัวนับ — Req 9.4
//   - เมื่ออนุญาต: หัก 1 token, retryAfter = 0
//   - เมื่อถูกปฏิเสธ: คำนวณ retryAfter = ceil(เวลาที่เหลือจนพ้นกรอบ) หน่วยวินาที
//     แล้ว clamp ให้อยู่ในช่วง 1..window(วินาที) — Req 9.3
func (rl *RateLimiter) allowWithRetry(ip string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowSecs := int(math.Ceil(rl.window.Seconds()))

	v, ok := rl.visitors[ip]
	if !ok {
		// ผู้เรียกใหม่: เริ่มกรอบใหม่แบบถังเต็มแล้วหัก 1 token
		rl.visitors[ip] = &visitor{tokens: rl.rate - 1, lastSeen: now}
		return true, 0
	}

	// พ้นกรอบเวลาแล้ว → รีเซ็ตถังให้เต็มและเริ่มกรอบใหม่ (Req 9.4)
	elapsed := now.Sub(v.lastSeen)
	if elapsed >= rl.window {
		v.tokens = rl.rate
		v.lastSeen = now
	}

	if v.tokens >= 1 {
		v.tokens--
		// อยู่ในกรอบเดิม: อัปเดต lastSeen เฉพาะเมื่อเป็นการเริ่มกรอบใหม่เท่านั้น
		// เพื่อให้กรอบเวลานับจากคำขอแรกของกรอบ (fixed window)
		return true, 0
	}

	// ถูกปฏิเสธ: คำนวณเวลาที่ต้องรอจนพ้นกรอบปัจจุบัน
	remaining := rl.window - now.Sub(v.lastSeen)
	retryAfter := int(math.Ceil(remaining.Seconds()))
	if retryAfter < 1 {
		retryAfter = 1
	}
	if retryAfter > windowSecs {
		retryAfter = windowSecs
	}
	return false, retryAfter
}

// LimitGinWithRetry เป็น middleware สำหรับ Gin ที่ใช้กับ endpoint audit
// เมื่อเกินโควตาจะตอบ HTTP 429 พร้อม header Retry-After และ body ที่มี retryAfter
// แล้ว abort โดยไม่ส่งต่อไปยัง handler (ไม่เรียก engine) — Req 9.2, 9.3
//
// ใช้ c.ClientIP() ซึ่งเคารพ SetTrustedProxies (รองรับทั้ง IPv4/IPv6 และกันการปลอม IP) — Req 9.6
func (rl *RateLimiter) LimitGinWithRetry() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowed, retryAfter := rl.allowWithRetry(c.ClientIP())
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":       "RATE_LIMIT_EXCEEDED",
					"message":    "ส่งคำขอบ่อยเกินไป โปรดลองใหม่อีกครั้งในภายหลัง",
					"retryAfter": retryAfter,
				},
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// getClientIP หาที่อยู่ IP ของผู้เรียกสำหรับ stdlib path:
// X-Forwarded-For (ตัวแรก) → X-Real-IP → RemoteAddr — แบบแผนเดิม
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For อาจมีหลาย IP คั่นด้วย ","; ใช้ตัวแรก (ต้นทางจริง)
		if idx := strings.IndexByte(xff, ','); idx >= 0 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	// RemoteAddr เป็นรูป host:port — ตัด port ออก
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}
