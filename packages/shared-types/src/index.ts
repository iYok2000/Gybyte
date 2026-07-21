/**
 * @mono-repo/shared-types
 *
 * สัญญาข้อมูล (data contract) เดียวที่ใช้ร่วมกันระหว่าง FE (apps/web) และ BE (apps/backend-go)
 * ของฟีเจอร์ AdReady (AdSense Compliance Auditor)
 *
 * นิยาม type เหล่านี้เป็น "single source of truth" — ห้ามประกาศซ้ำในฝั่ง FE หรือ BE
 * (Req 8.1, 8.2, 15.3, 15.4)
 */

/**
 * สถานะของ Audit_Result หนึ่งรายการ — จำกัดเฉพาะ "pass" หรือ "fail" เท่านั้น (Req 8.3)
 */
export type AuditStatus = "pass" | "fail";

/**
 * ผลลัพธ์ของ Audit_Check หนึ่งรายการ (Req 8.2)
 * - title: ชื่อสิ่งที่ตรวจสอบ (เช่น "ads.txt", "HTTPS")
 * - status: pass หรือ fail
 * - message: ข้อความอธิบายผล; สำหรับสถานะ fail ต้องไม่ว่างเปล่าและบอกวิธีแก้ (Req 8.4)
 */
export interface AuditResult {
  title: string;
  status: AuditStatus;
  message: string;
}

/**
 * คำขอตรวจสอบที่ FE ส่งไปยัง Audit_API — มีเพียง URL ของเว็บไซต์เป้าหมาย (Req 1.1, 8.1)
 */
export interface AuditRequest {
  url: string;
}

/**
 * การตอบกลับสำเร็จจาก Audit_API (Req 8.1, 8.5, 8.6)
 * - auditScore: จำนวนเต็ม 0–100; เป็น null เมื่อประเมินไม่ได้ (indeterminate — ไม่มีรายการที่ประเมินได้)
 * - results: อาร์เรย์ของ Audit_Result (คงลำดับ deterministic ตามที่ engine กำหนด)
 */
export interface AuditResponse {
  auditScore: number | null;
  results: AuditResult[];
}

/**
 * โครงสร้างข้อผิดพลาดที่ FE ใช้ตีความ (Req 9.3, 14.3)
 * - code: รหัสข้อผิดพลาด (เช่น "RATE_LIMIT_EXCEEDED", "VALIDATION_ERROR")
 * - message: ข้อความสำหรับแสดงผู้ใช้
 * - retryAfter: จำนวนวินาทีที่ต้องรอก่อนลองใหม่ (มีเฉพาะกรณี HTTP 429)
 */
export interface AuditError {
  code: string;
  message: string;
  retryAfter?: number;
}
