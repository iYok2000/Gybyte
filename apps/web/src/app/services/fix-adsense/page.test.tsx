import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import FixAdSensePage from "./page";

// Unit tests for the Fix_AdSense_Page.
// Requirements: 12.1, 12.2, 12.3, 12.4, 12.5

afterEach(() => {
  cleanup();
});

describe("Fix_AdSense_Page (Req 12.1–12.5)", () => {
  it("renders a main heading and explanatory content (Req 12.2)", () => {
    render(<FixAdSensePage />);
    expect(
      screen.getByRole("heading", {
        level: 1,
        name: "บริการแก้ปัญหาความพร้อม AdSense",
      }),
    ).toBeInTheDocument();
    expect(screen.getByText("เราช่วยอะไรได้บ้าง")).toBeInTheDocument();
  });

  it("presents both service packages with prominent prices (Req 12.3)", () => {
    render(<FixAdSensePage />);
    expect(screen.getByText("DIY Guide")).toBeInTheDocument();
    expect(screen.getByText("Done-For-You")).toBeInTheDocument();
    expect(screen.getByText("฿290")).toBeInTheDocument();
    expect(screen.getByText("฿1,990")).toBeInTheDocument();
  });

  it("renders a CTA per package pointing at the placeholder contact (Req 12.4, 12.7)", () => {
    const { container } = render(<FixAdSensePage />);
    const ctaLinks = container.querySelectorAll('a[href="#contact"]');
    expect(ctaLinks).toHaveLength(2);
  });

  it("provides a link back to the audit dashboard", () => {
    const { container } = render(<FixAdSensePage />);
    expect(container.querySelector('a[href="/audit"]')).not.toBeNull();
  });
});
