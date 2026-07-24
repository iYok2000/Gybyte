import type { Metadata } from "next";
import { Anuphan, Instrument_Serif, Noto_Serif_Thai } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/theme/ThemeProvider";
import { LanguageProvider } from "@/i18n/LanguageProvider";

// Primary UI font — Anuphan covers both Thai and Latin so the landing and
// AdReady/audit pages render consistently in either language. Exposed as
// --font-sans on <html> so it cascades everywhere.
const anuphan = Anuphan({
  subsets: ["thai", "latin"],
  weight: ["300", "400", "500", "600", "700"],
  variable: "--font-sans",
  display: "swap",
});

// Serif accent used by the About heading. Noto Serif Thai supports Thai
// glyphs (Instrument Serif does not), so it must be first in the serif stack.
const notoSerifThai = Noto_Serif_Thai({
  subsets: ["thai", "latin"],
  weight: ["400", "600"],
  variable: "--font-serif-thai",
  display: "swap",
});

// Kept as a Latin-only serif fallback appended after the Thai-capable serif.
const instrumentSerif = Instrument_Serif({
  subsets: ["latin"],
  weight: "400",
  style: "italic",
  variable: "--font-instrument-serif",
  display: "swap",
});

export const metadata: Metadata = {
  metadataBase: new URL(
    process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000",
  ),
  title: {
    default: "AdReady — ตรวจสอบความพร้อม Google AdSense ฟรี",
    template: "%s | AdReady",
  },
  description:
    "เครื่องมือตรวจสอบความพร้อม AdSense ฟรี เช็กว่าเว็บผ่าน AdSense ไหม — ตรวจ ads.txt, HTTPS, robots.txt และหน้า Privacy, Contact, About พร้อมคะแนน 0–100 และวิธีแก้เพื่อสมัคร AdSense ให้ผ่าน",
  keywords: [
    "ตรวจสอบความพร้อม AdSense",
    "เช็กเว็บผ่าน AdSense ไหม",
    "สมัคร AdSense ให้ผ่าน",
    "ตรวจ ads.txt",
    "เครื่องมือตรวจ AdSense ฟรี",
    "ทำเว็บให้พร้อม AdSense",
    "AdSense readiness checker",
    "check AdSense eligibility",
    "get approved for Google AdSense",
    "free ads.txt checker",
    "website AdSense audit tool",
  ],
  openGraph: {
    title: "AdReady — ตรวจสอบความพร้อม Google AdSense ฟรี",
    description:
      "เช็กว่าเว็บผ่าน AdSense ไหม ด้วยการตรวจ ads.txt, HTTPS, robots.txt และหน้า Privacy, Contact, About พร้อมคะแนนความพร้อม 0–100 และวิธีแก้เพื่อสมัคร AdSense ให้ผ่าน",
    type: "website",
    locale: "th_TH",
    siteName: "AdReady",
    url: "/",
  },
  twitter: {
    card: "summary_large_image",
    title: "AdReady — ตรวจสอบความพร้อม Google AdSense ฟรี",
    description:
      "เครื่องมือตรวจ AdSense ฟรี เช็ก ads.txt, HTTPS, robots.txt และหน้า compliance พร้อมคะแนนความพร้อมและวิธีแก้เพื่อสมัคร AdSense ให้ผ่าน",
  },
  robots: { index: true, follow: true },
  alternates: { canonical: "/" },
};

// Minimal root shell: html/body + theme context only. Per-area chrome (the
// AdReady Header/BodyLayout) lives in group layouts so the cinematic Prisma
// landing at "/" renders full-bleed without the AdReady header.
export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="th"
      suppressHydrationWarning
      className={`${anuphan.variable} ${notoSerifThai.variable} ${instrumentSerif.variable}`}
    >
      <body className="antialiased">
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <LanguageProvider>{children}</LanguageProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
