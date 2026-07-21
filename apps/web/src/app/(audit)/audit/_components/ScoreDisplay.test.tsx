import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import fc from "fast-check";
import { ScoreDisplay } from "./ScoreDisplay";

// Feature: adready-checker, Property 14: the FE score display shows only
// integer scores 0–100; null/undefined render an indeterminate message
// ("ไม่สามารถประเมินคะแนนได้"), and NaN/out-of-range/non-integer values render
// an error ("คะแนนไม่ถูกต้อง") without ever showing the bad value.
// Validates: Requirements 8.6, 10.5

afterEach(() => {
  cleanup();
});

// Valid: integers within [0, 100].
const validScore = fc.integer({ min: 0, max: 100 });

// Invalid: out of range, non-integer, or NaN.
const invalidScore = fc.oneof(
  fc.integer({ min: 101, max: 100_000 }),
  fc.integer({ min: -100_000, max: -1 }),
  fc
    .float({ min: Math.fround(0), max: Math.fround(100), noNaN: true })
    .filter((n) => !Number.isInteger(n)),
  fc.constant(Number.NaN),
);

describe("ScoreDisplay (Property 14)", () => {
  it("renders exactly the score for any valid integer 0–100", () => {
    fc.assert(
      fc.property(validScore, (score) => {
        cleanup();
        render(<ScoreDisplay score={score} />);

        // The score number is shown, and no error is raised.
        expect(screen.getByText(String(score))).toBeInTheDocument();
        expect(screen.queryByRole("alert")).toBeNull();
      }),
      { numRuns: 200 },
    );
  });

  it("shows the indeterminate status for null/undefined without a number", () => {
    fc.assert(
      fc.property(fc.constantFrom(null, undefined), (score) => {
        cleanup();
        render(<ScoreDisplay score={score} />);

        const status = screen.getByRole("status");
        // Indeterminate message, and never a numeric value nor an error.
        expect(status.textContent ?? "").toMatch(/ประเมิน/);
        expect(status.textContent ?? "").not.toMatch(/\d/);
        expect(screen.queryByRole("alert")).toBeNull();
      }),
      { numRuns: 100 },
    );
  });

  it("shows an error for NaN/out-of-range/non-integer and never the value", () => {
    fc.assert(
      fc.property(invalidScore, (score) => {
        cleanup();
        render(<ScoreDisplay score={score} />);

        const alert = screen.getByRole("alert");
        expect(alert.textContent).toBe("คะแนนไม่ถูกต้อง");

        // The invalid value itself must not appear anywhere.
        if (!Number.isNaN(score)) {
          expect(screen.queryByText(String(score))).toBeNull();
        }
      }),
      { numRuns: 200 },
    );
  });
});
