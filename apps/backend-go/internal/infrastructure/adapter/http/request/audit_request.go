// Package request นิยามโครงสร้างคำขอ (request DTO) ที่ผูกกับ body ของ HTTP endpoint
// ในเลเยอร์ adapter — แยก concern การ bind/validate รูปแบบคำขอออกจาก handler
package request

// RunAuditRequest คือ body ของคำขอ POST /api/audit
//
// ฟิลด์ URL เป็นค่าที่ผู้ใช้ส่งมาเพื่อตรวจสอบความพร้อม AdSense; ใช้ binding:"required"
// เพื่อให้ gin ปฏิเสธคำขอที่ไม่มีฟิลด์ url (→ handler ตอบ 400) การ normalize/SSRF guard
// ทำที่ backend (URL_Validator) เสมอ ไม่พึ่งพาการตรวจฝั่งหน้าเว็บ — Req 1.1, 8.1
type RunAuditRequest struct {
	URL string `json:"url" binding:"required"`
}
