package errors

import "fmt"

// DomainError คืออินเทอร์เฟซร่วมของ error เชิงโดเมนทุกชนิด
// รองรับการแม็พเป็น HTTP status และ response envelope โดยไม่ leak รายละเอียดภายใน
type DomainError interface {
	error
	Code() string    // รหัส error สำหรับ client/ตัวแม็พ status
	Message() string // ข้อความปลอดภัยสำหรับส่งกลับ client (ไม่รวม internal cause)
}

// NotFoundError — ทรัพยากรที่ระบุไม่พบ → 404
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string { return e.Message() }
func (e *NotFoundError) Code() string  { return CodeNotFound }
func (e *NotFoundError) Message() string {
	return fmt.Sprintf("%s with id %s not found", e.Resource, e.ID)
}

// ValidationError — อินพุตไม่ผ่านการตรวจสอบ → 400
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string { return e.Message() }
func (e *ValidationError) Code() string  { return CodeValidation }
func (e *ValidationError) Message() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Msg)
	}
	return e.Msg
}

// ConflictError — สถานะขัดแย้ง (เช่น ซ้ำซ้อน) → 409
type ConflictError struct {
	Resource string
	Msg      string
}

func (e *ConflictError) Error() string { return e.Message() }
func (e *ConflictError) Code() string  { return CodeConflict }
func (e *ConflictError) Message() string {
	if e.Resource != "" {
		return fmt.Sprintf("%s: %s", e.Resource, e.Msg)
	}
	return e.Msg
}

// InternalError — ข้อผิดพลาดภายในที่ไม่คาดคิด → 500
// Message() คืนเฉพาะ Msg (ข้อความทั่วไป) และไม่รวม cause เพื่อกัน leak รายละเอียดภายใน
// ส่วน cause จริงเข้าถึงได้ผ่าน Unwrap() สำหรับ logging ภายในเท่านั้น
type InternalError struct {
	Msg string
	Err error
}

func (e *InternalError) Error() string   { return e.Message() }
func (e *InternalError) Code() string    { return CodeInternal }
func (e *InternalError) Message() string { return e.Msg }
func (e *InternalError) Unwrap() error   { return e.Err }

// BadRequestError — คำขอไม่ถูกต้อง/ปลายทางไม่อนุญาต → 400
type BadRequestError struct {
	Msg string
}

func (e *BadRequestError) Error() string   { return e.Message() }
func (e *BadRequestError) Code() string    { return CodeBadRequest }
func (e *BadRequestError) Message() string { return e.Msg }

// --- Constructors ---

// NewValidationError สร้าง ValidationError ที่ระบุชื่อฟิลด์
func NewValidationError(field, msg string) *ValidationError {
	return &ValidationError{Field: field, Msg: msg}
}

// NewValidationErrorSimple สร้าง ValidationError โดยไม่ระบุฟิลด์
func NewValidationErrorSimple(msg string) *ValidationError {
	return &ValidationError{Msg: msg}
}

// NewBadRequestError สร้าง BadRequestError
func NewBadRequestError(msg string) *BadRequestError {
	return &BadRequestError{Msg: msg}
}

// NewInternalError ห่อ cause จริงไว้ภายใน แต่เปิดเผยเฉพาะข้อความทั่วไป (msg)
func NewInternalError(msg string, err error) *InternalError {
	return &InternalError{Msg: msg, Err: err}
}
