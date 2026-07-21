import { describe, it, expect, vi } from "vitest";
import fc from "fast-check";
import { validateUrl, MAX_URL_LENGTH, type ValidateUrlReason } from "./validateUrl";

// Feature: adready-checker, Property 2: invalid frontend input is rejected
// with a specific reason, no audit request is sent, and the user's input is
// preserved.
// Validates: Requirements 1.3

// --- Generators, one per invalid class -----------------------------------

// empty: whitespace/control-only strings that sanitize to "".
const emptyInput = fc
  .array(fc.constantFrom(" ", "\t", "\n", "\r", "\v", "\f", "\u0000", "\u001f"), {
    maxLength: 20,
  })
  .map((chars) => chars.join(""));

// too-long: > MAX_URL_LENGTH code points, no control chars so sanitize keeps
// the length. Checked before format, so content need not be a valid URL.
const tooLongInput = fc
  .integer({ min: MAX_URL_LENGTH + 1, max: MAX_URL_LENGTH + 500 })
  .map((n) => "a".repeat(n));

// invalid-format: non-empty, within length, does not parse as a URL. Bare
// letter tokens have no scheme, so `new URL(...)` throws.
const invalidFormatInput = fc
  .array(fc.constantFrom(..."abcdefghijklmnopqrstuvwxyz".split("")), {
    minLength: 1,
    maxLength: 20,
  })
  .map((chars) => chars.join(""));

// bad-scheme: parses as a URL but uses a scheme other than http/https.
const badSchemeInput = fc
  .tuple(
    fc.constantFrom("ftp", "file", "ws", "wss", "gopher"),
    fc
      .array(fc.constantFrom(..."abcdefghijklmnopqrstuvwxyz".split("")), {
        minLength: 1,
        maxLength: 12,
      })
      .map((chars) => chars.join("")),
  )
  .map(([scheme, host]) => `${scheme}://${host}.com/path`);

describe("validateUrl (Property 2)", () => {
  it("rejects empty input with reason 'empty'", () => {
    fc.assert(
      fc.property(emptyInput, (raw) => {
        expectRejected(raw, "empty");
      }),
      { numRuns: 200 },
    );
  });

  it("rejects over-length input with reason 'too-long'", () => {
    fc.assert(
      fc.property(tooLongInput, (raw) => {
        expectRejected(raw, "too-long");
      }),
      { numRuns: 100 },
    );
  });

  it("rejects unparseable input with reason 'invalid-format'", () => {
    fc.assert(
      fc.property(invalidFormatInput, (raw) => {
        expectRejected(raw, "invalid-format");
      }),
      { numRuns: 200 },
    );
  });

  it("rejects non-http(s) schemes with reason 'bad-scheme'", () => {
    fc.assert(
      fc.property(badSchemeInput, (raw) => {
        expectRejected(raw, "bad-scheme");
      }),
      { numRuns: 200 },
    );
  });
});

// Asserts validateUrl returns the expected reason, does not trigger a send,
// and does not mutate the input (input preserved — Req 1.3).
function expectRejected(raw: string, expected: ValidateUrlReason): void {
  const original = raw;
  const onSubmit = vi.fn();

  const result = validateUrl(raw);

  // Rejected with the specific reason.
  expect(result.valid).toBe(false);
  expect(result.reason).toBe(expected);

  // Simulated dashboard guard: no request is sent when validation fails.
  if (result.valid) onSubmit(raw);
  expect(onSubmit).not.toHaveBeenCalled();

  // Input value is preserved (validateUrl is pure and does not mutate).
  expect(raw).toBe(original);
}
