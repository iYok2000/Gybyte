// Package errors นิยาม domain error types + error codes ที่ใช้ร่วมทุกเลเยอร์ — Req 15.6
package errors

// Error codes มาตรฐานที่แม็พไปยัง HTTP status ใน response.HTTPStatusFromError
const (
	CodeNotFound     = "NOT_FOUND"
	CodeValidation   = "VALIDATION_ERROR"
	CodeConflict     = "CONFLICT"
	CodeInternal     = "INTERNAL_ERROR"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeBadRequest   = "BAD_REQUEST"
)
