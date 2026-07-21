// Package container ประกอบ (wire) dependency ของแต่ละโดเมนเข้าด้วยกัน
// และถือ handler ที่ใช้ร่วมกันทั้งแอป — Req 15.6
package container

import (
	"monorepo/backend-go/internal/application/audit/query"
	"monorepo/backend-go/internal/application/audit/validation"
	"monorepo/backend-go/internal/config"
	"monorepo/backend-go/internal/core/domain/audit/service"
	outboundhttp "monorepo/backend-go/internal/infrastructure/adapter/outbound/http"
)

// Container ถือ dependency/handler ที่ wire ไว้แล้วสำหรับใช้ในเลเยอร์ HTTP
type Container struct {
	cfg *config.Config

	// RunAuditHandler เป็น handler เดียวของโดเมน audit (stateless — ไม่มี DB/GORM)
	RunAuditHandler *query.RunAuditHandler
}

// New สร้าง Container โดยรับ config ที่โหลดแล้ว และ wire dependency ของทุกโดเมน
//
// การ wire สาย AdReady (stateless): urlValidator ถูกแชร์ระหว่าง SafeHTTPClient
// (redirect + dial-time SSRF guard) และ RunAuditHandler (input validation + SSRF
// guard) เพื่อให้ใช้กฎการกรองชุดเดียวกันตลอดทั้ง flow — Req 9.1, 15.2
func New(cfg *config.Config) *Container {
	// URL validator เดียวที่แชร์ทั้ง outbound client และ query handler
	urlValidator := validation.NewURLValidator()
	// SafeHTTPClient implement port.Fetcher โดยบังคับ SSRF guard ระดับ dial
	safeHTTPClient := outboundhttp.NewSafeHTTPClient(urlValidator)
	// engine ประกอบชุด check คงที่และรันแบบ concurrent
	auditEngine := service.NewEngine(safeHTTPClient)
	// query handler orchestrate: validate → SSRF guard → engine.Run → map response
	runAuditHandler := query.NewRunAuditHandler(urlValidator, auditEngine)

	return &Container{
		cfg:             cfg,
		RunAuditHandler: runAuditHandler,
	}
}

// Close ปล่อยทรัพยากรที่ Container ถือครอง (เช่น การเชื่อมต่อฐานข้อมูลของโดเมนอื่น)
// โดเมน audit เป็น stateless จึงไม่ผูกกับทรัพยากรที่ต้องปิดในที่นี้ — Req 15.5
func (c *Container) Close() error {
	return nil
}
