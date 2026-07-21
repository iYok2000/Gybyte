// Package config โหลดและถือค่าคอนฟิกของแอปพลิเคชันจาก environment variables
// ใช้ร่วมกันทุกเลเยอร์ (server bootstrap, middleware, container) — Req 15.6
package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config รวมค่าคอนฟิกทั้งหมดที่ถอดจาก environment variables
type Config struct {
	EnableHTTP         bool
	EnableGRPC         bool
	HTTPPort           string
	GRPCPort           string
	ServiceName        string
	Environment        string
	FrontendURL        string
	DatabaseURL        string
	CorsAllowedOrigins []string
	JWTSecret          string
	HTTPWriteTimeout   time.Duration
	TrustedProxies     []string
}

// ค่า default สำหรับ TRUSTED_PROXIES ครอบคลุมช่วง private/loopback/link-local
// (IPv4 + IPv6) เพื่อให้ gin เชื่อ X-Forwarded-For เฉพาะจาก peer ภายในเท่านั้น — Req 9.7
const defaultTrustedProxies = "127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,::1/128,fc00::/7,169.254.0.0/16"

// Load อ่าน environment variables ทั้งหมด, ใส่ค่า default ตามการออกแบบ
// และบังคับ invariant ที่จำเป็น (JWT_SECRET ต้องไม่ว่าง) — Req 9.7, 14.7, 15.6
func Load() *Config {
	cfg := &Config{
		EnableHTTP:         getEnvBool("ENABLE_HTTP", true),
		EnableGRPC:         getEnvBool("ENABLE_GRPC", false),
		HTTPPort:           getEnv("HTTP_PORT", "8080"),
		GRPCPort:           getEnv("GRPC_PORT", "50051"),
		ServiceName:        getEnv("SERVICE_NAME", "backend-go"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		CorsAllowedOrigins: getEnvArray("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		// HTTP_WRITE_TIMEOUT ต้องมากกว่า worst-case audit runtime มิฉะนั้น
		// การตอบกลับที่ใช้เวลานานจะถูกตัดกลางคัน — Req 14.7
		HTTPWriteTimeout: getEnvDuration("HTTP_WRITE_TIMEOUT", 60*time.Second),
		TrustedProxies:   getEnvArray("TRUSTED_PROXIES", defaultTrustedProxies),
	}

	// JWT_SECRET เป็นค่าบังคับ (required) — หยุดการทำงานทันทีถ้าว่างเปล่า
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required but was not set")
	}
	// เตือน (ไม่ fatal) เมื่อ secret สั้นเกินไปเพื่อความปลอดภัยของการเซ็น token
	if len(cfg.JWTSecret) < 32 {
		log.Printf("warning: JWT_SECRET is shorter than 32 characters; use a longer secret in production")
	}

	return cfg
}

// getEnv คืนค่า env ที่ระบุ หรือ fallback เมื่อไม่ได้ตั้งค่า/ว่างเปล่า
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvBool แปลงค่า env เป็น bool ด้วย strconv.ParseBool; fallback เมื่อไม่ตั้ง/parse ไม่ได้
func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}

// getEnvDuration รับได้ทั้งเลขล้วน (ตีความเป็นวินาที) หรือ duration string ("90s", "1m30s")
func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	// เลขล้วน = จำนวนวินาที
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second
	}
	// มิฉะนั้น parse เป็น duration string
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	return fallback
}

// getEnvArray split ค่า env ด้วย "," แล้ว trim ช่องว่างและตัดรายการว่างทิ้ง
func getEnvArray(key, fallback string) []string {
	v := os.Getenv(key)
	if v == "" {
		v = fallback
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
