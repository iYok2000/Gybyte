// Audit API client (Req 1.2, 1.5, 9.5, 14.3, 15.3).
//
// The frontend never calls the Go backend directly. It talks to the
// same-origin Next.js proxy (`/api/audit`), which forwards to the backend.
// This module owns the axios client and maps transport failures into the
// shared AuditError contract so the UI can render friendly Thai messages
// without leaking internal details (Req 14.3).

import { isAxiosError } from "axios";
import type {
  AuditRequest,
  AuditResponse,
  AuditError,
} from "@mono-repo/shared-types";
import { createAxiosInstance } from "@/lib/axios";

// Client-side timeout: abort if the audit has not responded within 30s
// (Req 1.5). Must be >= the proxy/backend budget so the surfaced error is a
// real timeout rather than a premature client abort.
export const AUDIT_TIMEOUT_MS = 30_000;

// Same-origin instance (empty baseURL → relative requests hit the Next.js
// proxy). Override the shared 10s default with the 30s audit budget.
const auditApi = createAxiosInstance("");
auditApi.defaults.timeout = AUDIT_TIMEOUT_MS;

/**
 * runAudit sends the target URL to the audit proxy and returns the parsed
 * AuditResponse. Rejections propagate to the caller, which should map them via
 * toAuditError.
 */
export async function runAudit(url: string): Promise<AuditResponse> {
  const body: AuditRequest = { url };
  const response = await auditApi.post<AuditResponse>("/api/audit", body);
  return response.data;
}

/**
 * toAuditError maps any thrown value from runAudit into the AuditError
 * contract. Handles: non-axios errors, client timeouts (ECONNABORTED),
 * network failures (no response), and server error bodies.
 */
export function toAuditError(error: unknown): AuditError {
  // Non-axios / unexpected error.
  if (!isAxiosError(error)) {
    return {
      code: "UNKNOWN_ERROR",
      message: "เกิดข้อผิดพลาดที่ไม่คาดคิด กรุณาลองใหม่อีกครั้ง",
    };
  }

  // Client-side timeout (Req 1.5).
  if (error.code === "ECONNABORTED") {
    return {
      code: "TIMEOUT",
      message: "การตรวจสอบใช้เวลานานเกินกำหนด กรุณาลองใหม่อีกครั้ง",
    };
  }

  // No response reached the client (connection refused, DNS, offline, ...).
  if (!error.response) {
    return {
      code: "NETWORK_ERROR",
      message:
        "ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้ กรุณาตรวจสอบการเชื่อมต่อแล้วลองใหม่",
    };
  }

  // Server responded with an error status: read the structured error body,
  // falling back to safe defaults so no internal detail leaks (Req 14.3).
  const { status, data, headers } = error.response;
  const errObj = extractErrorObject(data);

  const code =
    typeof errObj.code === "string" && errObj.code.length > 0
      ? errObj.code
      : `HTTP_${status}`;

  const message =
    typeof errObj.message === "string" && errObj.message.length > 0
      ? errObj.message
      : "การตรวจสอบไม่สำเร็จ กรุณาลองใหม่อีกครั้ง";

  // retryAfter: prefer the body value, fall back to the Retry-After header
  // (Req 9.5).
  const retryAfter = parseRetryAfter(
    errObj.retryAfter,
    headers?.["retry-after"],
  );

  const result: AuditError = { code, message };
  if (retryAfter !== undefined) {
    result.retryAfter = retryAfter;
  }
  return result;
}

// Safely pull the `error` object out of a response body of unknown shape.
function extractErrorObject(data: unknown): {
  code?: unknown;
  message?: unknown;
  retryAfter?: unknown;
} {
  if (data && typeof data === "object" && "error" in data) {
    const err = (data as { error?: unknown }).error;
    if (err && typeof err === "object") {
      return err as { code?: unknown; message?: unknown; retryAfter?: unknown };
    }
  }
  return {};
}

// Parse a positive-integer retry-after from a body value first, then a header.
function parseRetryAfter(
  bodyValue: unknown,
  headerValue: unknown,
): number | undefined {
  if (
    typeof bodyValue === "number" &&
    Number.isFinite(bodyValue) &&
    bodyValue > 0
  ) {
    return Math.floor(bodyValue);
  }
  if (typeof bodyValue === "string") {
    const n = Number.parseInt(bodyValue, 10);
    if (Number.isFinite(n) && n > 0) return n;
  }
  if (typeof headerValue === "string") {
    const n = Number.parseInt(headerValue, 10);
    if (Number.isFinite(n) && n > 0) return n;
  }
  return undefined;
}
