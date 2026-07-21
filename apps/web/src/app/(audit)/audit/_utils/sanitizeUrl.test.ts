import { describe, it, expect } from "vitest";
import fc from "fast-check";
import { sanitizeUrl } from "./sanitizeUrl";

// Feature: adready-checker, Property 1: URL sanitization strips control
// characters and trims leading/trailing whitespace idempotently; the length
// used for the 2048-char limit is counted AFTER sanitization.
// Validates: Requirements 1.6

// Generator that produces "messy" strings including control characters
// (U+0000–U+001F, U+007F–U+009F), whitespace, and assorted Unicode so the
// property exercises the stripping/trimming behaviour.
const messyString = fc
  .array(fc.integer({ min: 0, max: 0x2fff }), { maxLength: 128 })
  .map((codePoints) =>
    codePoints.map((cp) => String.fromCodePoint(cp)).join(""),
  );

const CONTROL_CHARS = /\p{Cc}/u;

describe("sanitizeUrl (Property 1)", () => {
  it("removes control chars, trims, and is idempotent", () => {
    fc.assert(
      fc.property(messyString, (raw) => {
        const once = sanitizeUrl(raw);

        // No control characters remain anywhere.
        expect(CONTROL_CHARS.test(once)).toBe(false);

        // No leading/trailing whitespace remains.
        expect(once).toBe(once.trim());

        // Idempotent: applying again changes nothing.
        expect(sanitizeUrl(once)).toBe(once);
      }),
      { numRuns: 200 },
    );
  });

  it("counts length after sanitization (leading/trailing padding ignored)", () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 50 }),
        fc.integer({ min: 0, max: 50 }),
        fc.integer({ min: 0, max: 40 }),
        (leftPad, rightPad, coreLen) => {
          const core = "a".repeat(coreLen);
          // Pad with spaces and control chars that must be stripped/trimmed.
          const raw = " ".repeat(leftPad) + "\t\n" + core + "\r" + " ".repeat(rightPad);
          const sanitized = sanitizeUrl(raw);
          // The tab/newline/carriage-return are control chars (stripped),
          // surrounding spaces are trimmed → only the core remains.
          expect([...sanitized].length).toBe(coreLen);
        },
      ),
      { numRuns: 200 },
    );
  });
});
