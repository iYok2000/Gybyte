// URL sanitization for the audit form (Req 1.6).
//
// Frontend validation is UX-only; the backend re-validates every URL. This
// helper normalizes raw user input so the format/length checks in validateUrl
// operate on a clean value.

// Matches every Unicode control character (general category Cc), e.g. NUL,
// tab, newline, carriage return, DEL and the C1 range.
const CONTROL_CHARS = /\p{Cc}/gu;

/**
 * sanitizeUrl strips control characters anywhere in the string and trims
 * leading/trailing whitespace. Normalization happens BEFORE any length
 * counting so the 2048-character limit is measured against the clean value
 * (Req 1.6).
 *
 * The operation is idempotent: sanitizeUrl(sanitizeUrl(x)) === sanitizeUrl(x).
 */
export function sanitizeUrl(raw: string): string {
  return raw.replace(CONTROL_CHARS, "").trim();
}

export default sanitizeUrl;
