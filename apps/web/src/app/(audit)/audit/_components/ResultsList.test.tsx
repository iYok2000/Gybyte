import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import fc from "fast-check";
import type { AuditResult } from "@mono-repo/shared-types";
import { ResultsList } from "./ResultsList";

// Feature: adready-checker, Property 15: ResultsList preserves the order of the
// results array, and the number of "แก้ปัญหานี้" (Get Help) buttons equals the
// number of failing items exactly (one per fail, none for pass).
// Validates: Requirements 11.1, 11.3, 11.5

afterEach(() => {
  cleanup();
});

// Titles use a whitespace-free ASCII charset so DOM textContent comparison is
// exact and never collides with the Thai button label.
const safeTitle = fc
  .array(fc.constantFrom(..."abcdefABCDEF0123456789".split("")), {
    minLength: 1,
    maxLength: 10,
  })
  .map((chars) => chars.join(""));

const resultArb: fc.Arbitrary<AuditResult> = fc.record({
  title: safeTitle,
  status: fc.constantFrom<"pass" | "fail">("pass", "fail"),
  message: fc.string({ maxLength: 40 }),
});

const resultsArb = fc.array(resultArb, { maxLength: 12 });

describe("ResultsList (Property 15)", () => {
  it("preserves order and shows one Get Help button per failing item", () => {
    fc.assert(
      fc.property(resultsArb, (results) => {
        cleanup();
        const { container } = render(<ResultsList results={results} />);

        // Order preserved: rendered titles match the input order exactly.
        const renderedTitles = Array.from(
          container.querySelectorAll("h3"),
        ).map((h) => h.textContent);
        expect(renderedTitles).toEqual(results.map((r) => r.title));

        // Get Help button count == number of failing items (Req 11.3, 11.5).
        const failCount = results.filter((r) => r.status === "fail").length;
        const buttons = screen.queryAllByRole("button", {
          name: "แก้ปัญหานี้",
        });
        expect(buttons).toHaveLength(failCount);
      }),
      { numRuns: 200 },
    );
  });
});
