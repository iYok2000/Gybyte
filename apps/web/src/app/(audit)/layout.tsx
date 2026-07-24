import type { Metadata } from "next";
import { SiteNav } from "@/components/layout/SiteNav";
import { BodyLayout } from "@/components/layout/BodyLayout";

export const metadata: Metadata = {
  title: "ตรวจสอบเว็บไซต์",
  description:
    "รันการตรวจสอบความพร้อม AdSense ให้เว็บไซต์ของคุณ — ตรวจ ads.txt, HTTPS, robots.txt และหน้า Privacy, Contact, About แล้วรับคะแนนความพร้อม 0–100 พร้อมรายการที่ต้องแก้เพื่อสมัคร AdSense ให้ผ่าน",
  openGraph: {
    title: "ตรวจสอบเว็บไซต์ | AdReady",
    description:
      "รันการตรวจสอบความพร้อม AdSense: ตรวจ ads.txt, HTTPS, robots.txt และหน้า compliance พร้อมคะแนน 0–100 และวิธีแก้เพื่อทำเว็บให้พร้อม AdSense",
    type: "website",
    locale: "th_TH",
    siteName: "AdReady",
    url: "/audit",
  },
  alternates: { canonical: "/audit" },
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
          "var(--font-sans), -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
      }}
    >
      <SiteNav />
      <BodyLayout>{children}</BodyLayout>
    </div>
  );
}
