// Next.js proxy BFF for the audit endpoint (Req 9.3, 9.5, 14.3, 15.1, 15.2).
//
// This route forwards audit requests to the Go backend verbatim. It performs
// NO audit logic itself (Req 15.2). Its only responsibilities are:
//  - forward the caller's real IP so backend rate limiting works per-IP,
//  - relay the upstream status/body/Retry-After header unchanged,
//  - translate connection failures/timeouts into a 502 error envelope.

import { NextResponse, type NextRequest } from "next/server";

// Backend base URL (trailing slashes stripped) + fixed audit endpoint.
const GO_API_URL = (
  process.env.NEXT_PUBLIC_GO_API_URL || "http://localhost:9000"
).replace(/\/+$/, "");
const AUDIT_ENDPOINT = `${GO_API_URL}/api/audit`;

// Upstream request budget. Matches the backend/client 30s audit window.
const UPSTREAM_TIMEOUT_MS = 30_000;

/**
 * extractClientIp resolves the caller's IP for per-IP rate limiting (Req 9.1):
 * the first entry of X-Forwarded-For, then X-Real-IP.
 */
function extractClientIp(req: NextRequest): string | null {
  const xff = req.headers.get("x-forwarded-for");
  if (xff) {
    const first = xff.split(",")[0]?.trim();
    if (first) return first;
  }
  const realIp = req.headers.get("x-real-ip");
  if (realIp) {
    const trimmed = realIp.trim();
    if (trimmed) return trimmed;
  }
  return null;
}

export async function POST(req: NextRequest): Promise<Response> {
  // Read the raw body and forward it untouched.
  const bodyText = await req.text();

  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), UPSTREAM_TIMEOUT_MS);

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  const clientIp = extractClientIp(req);
  if (clientIp) {
    // Forward the resolved client IP so the backend rate limiter counts the
    // real caller (Req 9.1).
    headers["X-Forwarded-For"] = clientIp;
  }

  try {
    const upstream = await fetch(AUDIT_ENDPOINT, {
      method: "POST",
      headers,
      body: bodyText,
      cache: "no-store",
      signal: controller.signal,
    });

    // Relay the upstream response verbatim: status + body, preserving the
    // content type and the Retry-After header when present (Req 9.3, 9.5).
    const responseBody = await upstream.text();
    const responseHeaders = new Headers();

    const contentType = upstream.headers.get("content-type");
    responseHeaders.set(
      "content-type",
      contentType ?? "application/json",
    );

    const retryAfter = upstream.headers.get("retry-after");
    if (retryAfter) {
      responseHeaders.set("retry-after", retryAfter);
    }

    return new NextResponse(responseBody, {
      status: upstream.status,
      headers: responseHeaders,
    });
  } catch (err) {
    // Timeout (AbortController) vs. connection failure → 502 with a generic
    // envelope that leaks no internal detail (Req 14.3).
    const isTimeout = err instanceof Error && err.name === "AbortError";
    const code = isTimeout ? "UPSTREAM_TIMEOUT" : "UPSTREAM_UNAVAILABLE";
    const message = isTimeout
      ? "การตรวจสอบใช้เวลานานเกินกำหนด กรุณาลองใหม่อีกครั้ง"
      : "ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ตรวจสอบได้ กรุณาลองใหม่ภายหลัง";

    return NextResponse.json(
      { success: false, error: { code, message } },
      { status: 502 },
    );
  } finally {
    clearTimeout(timeout);
  }
}
