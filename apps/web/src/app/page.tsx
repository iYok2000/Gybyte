import type { Metadata } from "next";
import About from "./prisma/_components/About";
import Features from "./prisma/_components/Features";
import Hero from "./prisma/_components/Hero";

// Landing-specific metadata — strongest AdSense-readiness keywords for "/".
// Overrides the root title/description; inherits the rest from the root layout.
export const metadata: Metadata = {
  title: "ตรวจสอบความพร้อม AdSense ฟรี — เช็กเว็บผ่าน AdSense ไหม",
  description:
    "AdReady เครื่องมือตรวจ AdSense ฟรี เช็กว่าเว็บของคุณผ่าน AdSense ไหม — ตรวจ ads.txt, HTTPS, robots.txt และหน้า Privacy, Contact, About คลิกเดียวรู้คะแนน 0–100 พร้อมวิธีทำเว็บให้พร้อม AdSense",
  openGraph: {
    title: "ตรวจสอบความพร้อม AdSense ฟรี — เช็กเว็บผ่าน AdSense ไหม",
    description:
      "เช็กว่าเว็บผ่าน AdSense ไหมด้วยการตรวจ ads.txt, HTTPS, robots.txt และหน้า compliance พร้อมคะแนนความพร้อม 0–100 และวิธีสมัคร AdSense ให้ผ่าน",
    url: "/",
  },
  alternates: { canonical: "/" },
};

const siteUrl = process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000";

// JSON-LD structured data — helps Google understand AdReady as a free web app
// and enables richer search results.
const jsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "WebSite",
      "@id": `${siteUrl}/#website`,
      name: "AdReady",
      url: siteUrl,
      inLanguage: "th",
      description:
        "เครื่องมือตรวจสอบความพร้อม Google AdSense ฟรี เช็กว่าเว็บผ่าน AdSense ไหม",
    },
    {
      "@type": "SoftwareApplication",
      "@id": `${siteUrl}/#app`,
      name: "AdReady",
      url: siteUrl,
      applicationCategory: "BusinessApplication",
      operatingSystem: "Web",
      description:
        "AdReady is a free Google AdSense readiness checker: it audits ads.txt, HTTPS, robots.txt, and Privacy/Contact/About pages, then returns a 0–100 readiness score with the exact fixes to get approved.",
      offers: {
        "@type": "Offer",
        price: "0",
        priceCurrency: "THB",
      },
    },
  ],
};

// Root route ("/") — the Prisma creative-studio landing page.
// Anuphan (Thai+Latin) is the default font for this subtree.
export default function HomePage() {
  return (
    <main
      className="bg-black min-h-screen"
      style={{
        fontFamily:
          "var(--font-sans), -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
      }}
    >
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <Hero />
      <About />
      <Features />
    </main>
  );
}
