// Package server ประกอบและ bootstrap HTTP server (Gin) ตามการออกแบบ — Req 9.6, 9.7, 14.7
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"monorepo/backend-go/internal/config"
	"monorepo/backend-go/internal/container"
	"monorepo/backend-go/internal/infrastructure/adapter/http/routes"
)

// HTTPServer ห่อ *http.Server พร้อมเมท็อดสำหรับ start/shutdown
type HTTPServer struct {
	server *http.Server
}

// NewHTTPServer สร้าง Gin router + http.Server ตามการตั้งค่าจริง:
//   - gin.SetMode(ReleaseMode) เมื่อ ENVIRONMENT == "production" มิฉะนั้น DebugMode
//   - SetTrustedProxies(cfg.TrustedProxies) เพื่อจำกัดว่า gin จะเชื่อ X-Forwarded-For
//     จาก peer ที่อยู่ในรายการ trusted เท่านั้น (anti-spoof สำหรับ c.ClientIP) — Req 9.6, 9.7
//   - WriteTimeout = cfg.HTTPWriteTimeout (default 60s) ต้องมากกว่า worst-case audit
//     runtime มิฉะนั้นการตอบกลับที่ใช้เวลานานจะถูกตัดกลางคัน — Req 14.7
func NewHTTPServer(cfg *config.Config, cnt *container.Container) (*HTTPServer, error) {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}

	routes.RegisterRoutes(router, cfg, cnt)

	return &HTTPServer{
		server: &http.Server{
			Addr:           ":" + cfg.HTTPPort,
			Handler:        router,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   cfg.HTTPWriteTimeout,
			MaxHeaderBytes: 1 << 20, // 1 MiB
		},
	}, nil
}

// Start เริ่มรับการเชื่อมต่อ (blocking จนกว่า server ปิดหรือเกิด error)
// http.ErrServerClosed ถือเป็นการปิดปกติ ผู้เรียกควรจัดการแยกจาก error อื่น
func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown ปิด server แบบ graceful ภายใน deadline ของ ctx ที่ส่งเข้ามา
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
