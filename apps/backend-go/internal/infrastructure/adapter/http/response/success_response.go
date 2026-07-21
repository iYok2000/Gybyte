package response

// NewSuccessResponse สร้าง envelope สำหรับผลลัพธ์ที่สำเร็จพร้อม payload
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}
