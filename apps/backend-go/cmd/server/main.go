// Command server เป็น entrypoint ของ backend: โหลด config, init logger,
// wire container และรัน HTTP server พร้อม graceful shutdown
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"monorepo/backend-go/internal/config"
	"monorepo/backend-go/internal/container"
	httpserver "monorepo/backend-go/internal/infrastructure/adapter/http/server"
	"monorepo/backend-go/pkg/logger"
)

func main() {
	// โหลด env จากไฟล์ .env ที่รากเรโป (best-effort — ไม่ fatal ถ้าไม่มีไฟล์
	// เพราะใน production ค่ามักถูกตั้งผ่าน environment โดยตรง)
	_ = godotenv.Load("../../.env")

	cfg := config.Load()

	logger.Init(cfg.Environment)
	defer logger.Sync()

	// ต้องเปิดอย่างน้อยหนึ่ง server (HTTP หรือ gRPC) มิฉะนั้นไม่มีอะไรให้รัน
	if !cfg.EnableHTTP && !cfg.EnableGRPC {
		log.Fatal("no server enabled: set ENABLE_HTTP or ENABLE_GRPC to true")
	}

	cnt := container.New(cfg)
	defer func() {
		if err := cnt.Close(); err != nil {
			logger.L().Error("failed to close container", zap.Error(err))
		}
	}()

	// srv ถูกอ้างถึงตอน graceful shutdown; เก็บไว้เพื่อเรียก Shutdown ได้
	var srv *httpserver.HTTPServer

	if cfg.EnableHTTP {
		var err error
		srv, err = httpserver.NewHTTPServer(cfg, cnt)
		if err != nil {
			log.Fatalf("failed to create HTTP server: %v", err)
		}

		// รัน HTTP server ใน goroutine เพื่อไม่บล็อกสาย main ที่รอสัญญาณปิด
		go func() {
			logger.L().Info("HTTP server starting", zap.String("port", cfg.HTTPPort))
			if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("HTTP server error: %v", err)
			}
		}()
	}

	// รอสัญญาณ SIGINT/SIGTERM เพื่อเริ่ม graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.L().Info("shutting down server...")

	// ให้เวลา in-flight request เสร็จภายใน 5 วินาที
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if srv != nil {
		if err := srv.Shutdown(ctx); err != nil {
			logger.L().Error("HTTP server forced to shutdown", zap.Error(err))
		}
	}

	logger.L().Info("server stopped")
}
