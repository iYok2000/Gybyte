package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"monorepo/backend-go/internal/config"
)

// CORS คืน Gin middleware ที่อนุญาต cross-origin request เฉพาะ origin ที่ตรงกับ
// รายการใน cfg.CorsAllowedOrigins เท่านั้น (allowlist) — Req 15.6
//
// พฤติกรรม:
//   - สะท้อน Access-Control-Allow-Origin เฉพาะเมื่อ origin ของคำขออยู่ใน allowlist
//   - อนุญาต credentials, กำหนด methods/headers และ cache preflight ไว้ 3600 วินาที
//   - ตอบ 204 ทันทีสำหรับ preflight (OPTIONS) โดยไม่ส่งต่อไปยัง handler
func CORS(cfg *config.Config) gin.HandlerFunc {
	// สร้าง set ของ origin ที่อนุญาตไว้ล่วงหน้าเพื่อ lookup แบบ O(1)
	allowed := make(map[string]struct{}, len(cfg.CorsAllowedOrigins))
	for _, o := range cfg.CorsAllowedOrigins {
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				h := c.Writer.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Access-Control-Allow-Credentials", "true")
				h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
				h.Set("Access-Control-Max-Age", "3600")
			}
		}

		// ตอบ preflight ทันทีโดยไม่ส่งต่อไปยัง handler ปลายทาง
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
