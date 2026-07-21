package response

import (
	"net/http"

	apperrors "monorepo/backend-go/pkg/errors"
)

// HTTPStatusFromError แม็พ error เป็น HTTP status code ตามการออกแบบ:
//   - *NotFoundError   → 404
//   - *ValidationError → 400
//   - *ConflictError   → 409
//   - *BadRequestError → 400
//   - *InternalError   → 500
//   - DomainError อื่น → แม็พตาม Code()
//   - default          → 500
func HTTPStatusFromError(err error) int {
	switch e := err.(type) {
	case *apperrors.NotFoundError:
		return http.StatusNotFound
	case *apperrors.ValidationError:
		return http.StatusBadRequest
	case *apperrors.ConflictError:
		return http.StatusConflict
	case *apperrors.BadRequestError:
		return http.StatusBadRequest
	case *apperrors.InternalError:
		return http.StatusInternalServerError
	case apperrors.DomainError:
		return statusFromCode(e.Code())
	default:
		return http.StatusInternalServerError
	}
}

// statusFromCode แม็พ error code ทั่วไปของ DomainError ที่ไม่ตรงชนิดข้างต้น
func statusFromCode(code string) int {
	switch code {
	case apperrors.CodeNotFound:
		return http.StatusNotFound
	case apperrors.CodeValidation:
		return http.StatusBadRequest
	case apperrors.CodeConflict:
		return http.StatusConflict
	case apperrors.CodeBadRequest:
		return http.StatusBadRequest
	case apperrors.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperrors.CodeForbidden:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// ErrorResponseFromError สร้าง Response จาก error:
//   - ถ้าเป็น DomainError → ใช้ Code()/Message() ที่ปลอดภัย
//   - อื่น ๆ → คืน INTERNAL_ERROR + ข้อความทั่วไป (ไม่ leak รายละเอียดภายใน) — Req 14.5
func ErrorResponseFromError(err error) Response {
	if de, ok := err.(apperrors.DomainError); ok {
		return NewErrorResponse(de.Code(), de.Message())
	}
	return NewErrorResponse(apperrors.CodeInternal, "internal server error")
}
