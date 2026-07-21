package middleware

import "github.com/gin-gonic/gin"

// SecurityHeadersGin ตั้งค่า HTTP security headers มาตรฐานให้ทุก response
// เพื่อลดความเสี่ยง clickjacking, MIME sniffing, XSS และควบคุมสิทธิ์ของเบราว์เซอร์
// ใช้แบบแผนเดียวกับ monorepo ต้นทาง — Req 15.6
func SecurityHeadersGin() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		// กัน clickjacking: ห้ามฝังหน้าใน frame/iframe
		h.Set("X-Frame-Options", "DENY")
		// กันเบราว์เซอร์เดา content-type ผิด (MIME sniffing)
		h.Set("X-Content-Type-Options", "nosniff")
		// เปิด XSS filter ของเบราว์เซอร์รุ่นเก่าและบล็อกเมื่อพบ
		h.Set("X-XSS-Protection", "1; mode=block")
		// บังคับ HTTPS สำหรับการเชื่อมต่อในอนาคต (รวมทุก subdomain)
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		// จำกัดแหล่งทรัพยากรที่โหลดได้ตาม Content Security Policy
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'")
		// จำกัดข้อมูล referrer ที่ส่งออกไปยัง origin อื่น
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// ปิดการเข้าถึงฟีเจอร์เบราว์เซอร์ที่ไม่จำเป็น
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}
