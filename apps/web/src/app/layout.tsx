import type { Metadata } from "next";
import { Almarai, Instrument_Serif } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/theme/ThemeProvider";
import { LanguageProvider } from "@/i18n/LanguageProvider";

// Fonts for the Prisma landing page (root "/"). Exposed as CSS variables on
// <html> so they cascade everywhere; AdReady pages keep the system font
// because nothing applies var(--font-almarai) there.
const almarai = Almarai({
  subsets: ["arabic", "latin"],
  weight: ["300", "400", "700", "800"],
  variable: "--font-almarai",
  display: "swap",
});

const instrumentSerif = Instrument_Serif({
  subsets: ["latin"],
  weight: "400",
  style: "italic",
  variable: "--font-instrument-serif",
  display: "swap",
});

export const metadata: Metadata = {
  title: "Prisma — Creative Studio",
  description:
    "Prisma is a worldwide network of visual artists, filmmakers and storytellers.",
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
      lang="en"
      suppressHydrationWarning
      className={`${almarai.variable} ${instrumentSerif.variable}`}
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
