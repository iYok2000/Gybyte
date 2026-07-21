// Frontend URL validation for the audit form (Req 1.3, 1.6, 2.5).
//
// This is UX-only: it gives the user immediate, specific feedback before a
// request is sent. The backend (URL_Validator) always re-validates and is the
// authoritative SSRF/format guard — never rely on this check for security.

import { sanitizeUrl } from "./sanitizeUrl";

// Maximum accepted URL length, counted AFTER normalization (Req 1.3, 2.1).
export const MAX_URL_LENGTH = 2048;

// Specific rejection reasons — one per failure case so the UI can surface a
// targeted message (Req 1.3).
export type ValidateUrlReason =
  | "empty"
  | "too-long"
  | "invalid-format"
  | "bad-scheme";

export interface ValidateUrlResult {
  valid: boolean;
  reason?: ValidateUrlReason;
}

// Only http/https schemes are allowed (Req 2.1, 2.2). Compared against the
// URL API's `protocol`, which includes the trailing colon.
const ALLOWED_PROTOCOLS = new Set(["http:", "https:"]);

/**
 * validateUrl normalizes the raw input, then checks (in order): non-empty,
 * length within MAX_URL_LENGTH, parseable URL format, and an allowed scheme.
 * Returns { valid: true } on success, otherwise { valid: false, reason } with
 * the specific failing case.
 */
export function validateUrl(raw: string): ValidateUrlResult {
  const sanitized = sanitizeUrl(raw);

  // Empty (or whitespace/control-only) input.
  if (sanitized.length === 0) {
    return { valid: false, reason: "empty" };
  }

  // Length is counted by Unicode code points (runes) to match the backend.
  if ([...sanitized].length > MAX_URL_LENGTH) {
    return { valid: false, reason: "too-long" };
  }

  // Format check via the WHATWG URL parser.
  let parsed: URL;
  try {
    parsed = new URL(sanitized);
  } catch {
    return { valid: false, reason: "invalid-format" };
  }

  // Scheme must be http or https.
  if (!ALLOWED_PROTOCOLS.has(parsed.protocol.toLowerCase())) {
    return { valid: false, reason: "bad-scheme" };
  }

  return { valid: true };
}

export default validateUrl;
