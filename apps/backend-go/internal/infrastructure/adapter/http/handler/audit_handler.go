// Package handler ประกอบด้วย HTTP handler ของเลเยอร์ adapter ที่รับคำขอ, เรียก
// application query/command แล้วแปลงผล/ข้อผิดพลาดเป็น HTTP response envelope
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	auditquery "monorepo/backend-go/internal/application/audit/query"
	"monorepo/backend-go/internal/infrastructure/adapter/http/request"
	"monorepo/backend-go/internal/infrastructure/adapter/http/response"
)

// AuditHandler จัดการ endpoint audit โดย delegate ตรรกะทั้งหมดไปยัง RunAuditHandler
// (application query) — handler ทำหน้าที่เพียง bind คำขอและ map ผล/error → HTTP
type AuditHandler struct {
	runAuditHandler *auditquery.RunAuditHandler
}

// NewAuditHandler สร้าง AuditHandler โดยรับ RunAuditHandler ที่ wire แล้วจาก container
func NewAuditHandler(runAuditHandler *auditquery.RunAuditHandler) *AuditHandler {
	return &AuditHandler{runAuditHandler: runAuditHandler}
}

// RunAudit จัดการ POST /api/audit:
//
//  1. bind JSON body → RunAuditRequest; ถ้า bind ล้มเหลว (เช่น ไม่มีฟิลด์ url)
//     ตอบ 400 VALIDATION_ERROR โดยไม่เรียก application layer — Req 8.1, 8.7
//  2. เรียก RunAuditHandler.Handle; ถ้าเกิด error map เป็น HTTP status + envelope
//     ที่ปลอดภัย (ไม่ leak รายละเอียดภายใน) ผ่าน response helpers — Req 8.7, 14.3, 14.5
//  3. สำเร็จ → 200 พร้อม payload { auditScore, results }
func (h *AuditHandler) RunAudit(c *gin.Context) {
	var req request.RunAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.NewErrorResponse("VALIDATION_ERROR", "Invalid request body"))
		return
	}

	result, err := h.runAuditHandler.Handle(c.Request.Context(), req.URL)
	if err != nil {
		c.JSON(response.HTTPStatusFromError(err), response.ErrorResponseFromError(err))
		return
	}

	c.JSON(http.StatusOK, result)
}
