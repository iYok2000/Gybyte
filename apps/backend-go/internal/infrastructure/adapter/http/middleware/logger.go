// Package middleware รวม Gin middleware ที่ใช้ร่วมกัน (logging, security headers,
// CORS, rate limiting) ตามแบบแผนเดียวกับ monorepo ต้นทาง — Req 15.6
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"monorepo/backend-go/pkg/logger"
)

// Logger คืน Gin middleware ที่บันทึก access log แบบ structured ด้วย zap
// เก็บ method/path/status/duration/ip โดยใช้ c.ClientIP() ซึ่งเคารพ SetTrustedProxies
// (ป้องกันการปลอม X-Forwarded-For จาก peer ที่ไม่อยู่ในรายการ trusted) — Req 15.6, 9.6
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// ประมวลผล request ต่อในสาย middleware ก่อน แล้วค่อยบันทึกผลลัพธ์
		c.Next()

		logger.L().Info("http_request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		)
	}
}
