import type { Metadata } from "next";
import { SiteNav } from "@/components/layout/SiteNav";
import { BodyLayout } from "@/components/layout/BodyLayout";

export const metadata: Metadata = {
  title: "AdReady — AdSense Compliance Auditor",
  description:
    "ตรวจสอบความพร้อมของเว็บไซต์สำหรับ Google AdSense: ads.txt, HTTPS, robots.txt และลิงก์ compliance",
};

// AdReady shell — shared floating-pill SiteNav (same as the Prisma landing) +
// constrained BodyLayout (Req 12.5, 13.1). pt-16 clears the fixed nav pill.
export default function AdReadyLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div
      className="flex min-h-screen flex-col bg-black pt-16 text-[#E1E0CC]"
      style={{
        fontFamily:
          "var(--font-almarai), -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
      }}
    >
      <SiteNav />
      <BodyLayout>{children}</BodyLayout>
    </div>
  );
}
