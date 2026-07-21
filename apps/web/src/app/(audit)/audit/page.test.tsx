import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import type { AuditResult } from "@mono-repo/shared-types";
import { AuditForm } from "./_components/AuditForm";
import { ScoreDisplay } from "./_components/ScoreDisplay";
import { ResultsList } from "./_components/ResultsList";

// Unit tests for audit form + result rendering.
// Requirements: 1.1, 1.4, 10.1, 10.2, 10.3, 10.4, 11.2, 11.4, 11.6

afterEach(() => {
  cleanup();
});

describe("AuditForm (Req 1.1, 1.4)", () => {
  it("shows the URL input and submit button together", () => {
    render(<AuditForm onSubmit={vi.fn()} />);
    expect(
      screen.getByLabelText("URL ของเว็บไซต์ที่ต้องการตรวจสอบ"),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "ตรวจสอบ" })).toBeInTheDocument();
  });

  it("disables the button and shows loading text while loading (Req 1.4)", () => {
    render(<AuditForm onSubmit={vi.fn()} loading />);
    const button = screen.getByRole("button", { name: "กำลังตรวจสอบ..." });
    expect(button).toBeDisabled();
  });

  it("submits the sanitized URL when input is valid", () => {
    const onSubmit = vi.fn();
    const { container } = render(<AuditForm onSubmit={onSubmit} />);
    const input = screen.getByLabelText("URL ของเว็บไซต์ที่ต้องการตรวจสอบ");
    fireEvent.change(input, { target: { value: "https://example.com" } });
    fireEvent.submit(container.querySelector("form")!);
    expect(onSubmit).toHaveBeenCalledWith("https://example.com");
  });

  it("rejects invalid input with a message and does not submit (Req 1.3)", () => {
    const onSubmit = vi.fn();
    const { container } = render(<AuditForm onSubmit={onSubmit} />);
    const input = screen.getByLabelText("URL ของเว็บไซต์ที่ต้องการตรวจสอบ");
    fireEvent.change(input, { target: { value: "ftp://example.com" } });
    fireEvent.submit(container.querySelector("form")!);
    expect(onSubmit).not.toHaveBeenCalled();
    expect(screen.getByRole("alert")).toBeInTheDocument();
    // Input preserved.
    expect((input as HTMLInputElement).value).toBe("ftp://example.com");
  });
});

describe("ScoreDisplay (Req 10.1–10.4)", () => {
  it("shows a loading spinner and no score while loading (Req 10.4)", () => {
    const { container } = render(<ScoreDisplay loading />);
    expect(screen.getByRole("status")).toBeInTheDocument();
    expect(container.querySelector(".text-7xl")).toBeNull();
  });

  it("renders the score as the largest element (Req 10.1, 10.2)", () => {
    const { container } = render(<ScoreDisplay score={73} />);
    const big = container.querySelector(".text-7xl");
    expect(big).not.toBeNull();
    expect(big?.textContent).toBe("73");
  });

  it("colors the score by band (Req 10.7)", () => {
    const { container: low } = render(<ScoreDisplay score={30} />);
    expect(low.querySelector(".text-7xl")?.className).toContain("--error");
    cleanup();
    const { container: mid } = render(<ScoreDisplay score={70} />);
    expect(mid.querySelector(".text-7xl")?.className).toContain("--warning");
    cleanup();
    const { container: high } = render(<ScoreDisplay score={95} />);
    expect(high.querySelector(".text-7xl")?.className).toContain("--success");
  });
});

describe("ResultsList / ResultItem (Req 11.2, 11.4, 11.6)", () => {
  const failResult: AuditResult = {
    title: "ads.txt",
    status: "fail",
    message: "ไม่พบไฟล์ ads.txt",
  };
  const passResult: AuditResult = {
    title: "HTTPS",
    status: "pass",
    message: "ใช้ HTTPS ถูกต้อง",
  };

  it("shows pass/fail badges to distinguish status (Req 11.2)", () => {
    render(<ResultsList results={[passResult, failResult]} />);
    expect(screen.getByText("ผ่าน")).toBeInTheDocument();
    expect(screen.getByText("ไม่ผ่าน")).toBeInTheDocument();
  });

  it("links the Get Help button to /services/fix-adsense on fail (Req 11.4)", () => {
    const { container } = render(<ResultsList results={[failResult]} />);
    const link = container.querySelector('a[href="/services/fix-adsense"]');
    expect(link).not.toBeNull();
    expect(screen.getByRole("button", { name: "แก้ปัญหานี้" })).toBeInTheDocument();
  });

  it("shows no Get Help button for passing checks (Req 11.5)", () => {
    render(<ResultsList results={[passResult]} />);
    expect(
      screen.queryByRole("button", { name: "แก้ปัญหานี้" }),
    ).toBeNull();
  });

  it("shows an empty message and no button when results are empty (Req 11.6)", () => {
    render(<ResultsList results={[]} />);
    expect(screen.getByText("ไม่มีผลลัพธ์")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "แก้ปัญหานี้" }),
    ).toBeNull();
  });
});
