// Package routes ลงทะเบียน global middleware และ endpoint ทั้งหมดของ HTTP server
package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"monorepo/backend-go/internal/config"
	"monorepo/backend-go/internal/container"
	"monorepo/backend-go/internal/infrastructure/adapter/http/handler"
	"monorepo/backend-go/internal/infrastructure/adapter/http/middleware"
)

// RegisterRoutes ติดตั้ง global middleware ตามลำดับ Logger → SecurityHeaders → CORS
// จากนั้นลงทะเบียน health endpoints และ group /api
//
// สร้าง rate limiter สามตัวตามการใช้งาน:
//   - loginRateLimiter (5/15m)  สำหรับ endpoint ล็อกอินของโดเมนอื่น
//   - apiRateLimiter   (100/1m) สำหรับ API ทั่วไป
//   - auditRateLimiter (60/60s) สำหรับ endpoint audit (+Retry-After) — Req 9.1
//
// หมายเหตุ: endpoint audit (POST /api/audit) จะถูกลงทะเบียนในงาน 15.3
func RegisterRoutes(router *gin.Engine, cfg *config.Config, cnt *container.Container) {
	// global middleware — ลำดับสำคัญ: log ก่อน, ตั้ง security headers, แล้วจัดการ CORS
	router.Use(middleware.Logger())
	router.Use(middleware.SecurityHeadersGin())
	router.Use(middleware.CORS(cfg))

	// rate limiter สำหรับแต่ละกลุ่มการใช้งาน (สร้างครั้งเดียวต่อ instance)
	_ = middleware.NewLoginRateLimiter() // 5 / 15m  (สงวนไว้สำหรับโดเมน auth)
	_ = middleware.NewAPIRateLimiter()   // 100 / 1m (สงวนไว้สำหรับ API ทั่วไป)
	auditRateLimiter := middleware.NewAuditRateLimiter()

	// health endpoints — ใช้ตรวจสุขภาพ service
	router.GET("/", healthHandler)
	router.GET("/health", healthHandler)
	router.GET("/ping", healthHandler)

	// group /api — endpoint audit สาธารณะ จำกัดอัตรา 60/60s ต่อ IP (+Retry-After)
	// ก่อนเรียก engine; 429 จะ abort โดยไม่เรียก handler ปลายทาง — Req 9.1
	api := router.Group("/api")
	auditHandler := handler.NewAuditHandler(cnt.RunAuditHandler)
	api.POST("/audit", auditRateLimiter.LimitGinWithRetry(), auditHandler.RunAudit)
}

// healthHandler ตอบสถานะพร้อมใช้งานของ service
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
