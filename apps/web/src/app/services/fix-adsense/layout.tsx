import type { Metadata } from "next";

// Server-component layout for /services/fix-adsense so the page (a client
// component) can still expose SEO metadata. Nests inside the services group
// layout (SiteNav shell) and just renders its children.
export const metadata: Metadata = {
  title: "บริการแก้ปัญหา AdSense",
  description:
    "ไม่ผ่านเกณฑ์ AdSense? เราช่วยทำเว็บให้พร้อม AdSense — แก้ ads.txt, เปิด HTTPS, จัดการ robots.txt และเพิ่มหน้า Privacy, Contact, About เพื่อสมัคร AdSense ให้ผ่าน",
  openGraph: {
    title: "บริการแก้ปัญหา AdSense | AdReady",
    description:
      "บริการแก้ปัญหาความพร้อม AdSense: แก้ ads.txt, HTTPS, robots.txt และหน้า compliance ให้ครบ เพื่อทำเว็บให้พร้อม AdSense และสมัครให้ผ่าน",
    type: "website",
    locale: "th_TH",
    siteName: "AdReady",
    url: "/services/fix-adsense",
  },
  alternates: { canonical: "/services/fix-adsense" },
};

export default function FixAdSenseLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return <>{children}</>;
}
