import { describe, it, expect } from "vitest";
import fc from "fast-check";
import type { AuditResponse, AuditResult } from "@mono-repo/shared-types";

// Feature: adready-checker, Property 10: round-trip serialize/deserialize of
// AuditResponse on the TypeScript side preserves the data and the schema —
// auditScore stays an integer 0–100 or null, and every result keeps a status
// within {pass, fail}. A status outside that set is rejected by the guard.
// Validates: Requirements 8.1, 8.2, 8.3

// Schema guard mirroring the shared-types contract (Req 8.1, 8.2, 8.3).
function isAuditStatus(value: unknown): value is "pass" | "fail" {
  return value === "pass" || value === "fail";
}

function isAuditResponse(value: unknown): value is AuditResponse {
  if (value === null || typeof value !== "object") return false;
  const obj = value as { auditScore?: unknown; results?: unknown };

  // auditScore: integer within [0, 100], or null (indeterminate).
  const score = obj.auditScore;
  const scoreOk =
    score === null ||
    (typeof score === "number" &&
      Number.isInteger(score) &&
      score >= 0 &&
      score <= 100);
  if (!scoreOk) return false;

  // results: array of { title, status ∈ {pass,fail}, message }.
  if (!Array.isArray(obj.results)) return false;
  return obj.results.every((r) => {
    if (r === null || typeof r !== "object") return false;
    const item = r as { title?: unknown; status?: unknown; message?: unknown };
    return (
      typeof item.title === "string" &&
      typeof item.message === "string" &&
      isAuditStatus(item.status)
    );
  });
}

const resultArb: fc.Arbitrary<AuditResult> = fc.record({
  title: fc.string({ minLength: 1, maxLength: 200 }),
  status: fc.constantFrom<"pass" | "fail">("pass", "fail"),
  message: fc.string({ maxLength: 1000 }),
});

const responseArb: fc.Arbitrary<AuditResponse> = fc.record({
  auditScore: fc.oneof(fc.integer({ min: 0, max: 100 }), fc.constant(null)),
  results: fc.array(resultArb, { maxLength: 20 }),
});

describe("AuditResponse round-trip (Property 10)", () => {
  it("preserves data and schema across JSON serialize/deserialize", () => {
    fc.assert(
      fc.property(responseArb, (response) => {
        const roundTripped = JSON.parse(
          JSON.stringify(response),
        ) as AuditResponse;

        // Data is preserved exactly.
        expect(roundTripped).toEqual(response);
        // Schema still holds after the round-trip.
        expect(isAuditResponse(roundTripped)).toBe(true);
        // Field-level invariants (Req 8.1, 8.2).
        expect(roundTripped.auditScore).toEqual(response.auditScore);
        expect(roundTripped.results).toHaveLength(response.results.length);
      }),
      { numRuns: 200 },
    );
  });

  it("rejects a status outside {pass, fail}", () => {
    fc.assert(
      fc.property(
        fc.string().filter((s) => s !== "pass" && s !== "fail"),
        (badStatus) => {
          const invalid = {
            auditScore: 50,
            results: [{ title: "x", status: badStatus, message: "" }],
          };
          expect(isAuditResponse(invalid)).toBe(false);
        },
      ),
      { numRuns: 100 },
    );
  });
});
