// Package response นิยาม response envelope ร่วม + การแม็พ error → HTTP status — Req 15.6
package response

// ErrorBody เป็นโครงข้อผิดพลาดที่ปลอดภัยต่อการส่งกลับ client (ไม่มีรายละเอียดภายใน)
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Response คือ envelope มาตรฐานของทุก HTTP response
// success=false → มี error; success=true → มี data (อ็อพชัน)
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

// NewErrorResponse สร้าง Response ที่แสดงข้อผิดพลาดด้วย code/message ที่กำหนด
func NewErrorResponse(code, message string) Response {
	return Response{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	}
}
