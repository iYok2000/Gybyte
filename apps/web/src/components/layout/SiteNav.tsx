"use client";

import Link from "next/link";
import { useState } from "react";
import { useLang } from "@/i18n/LanguageProvider";

// Shared floating-pill navigation used on both the Prisma landing ("/") and
// the AdReady pages (/audit, /services/*). A black pill hangs from the top
// edge, centered horizontally, with warm-cream links.

const NAV_LINKS: { label: string; href: string }[] = [
  { label: "Our story", href: "/" },
  { label: "Collective", href: "#" },
  { label: "Workshops", href: "#" },
  { label: "Audit", href: "/audit" },
  { label: "Inquiries", href: "#" },
];

function NavLink({ label, href }: { label: string; href: string }) {
  const [hovered, setHovered] = useState(false);
  return (
    <Link
      href={href}
      className="text-[10px] sm:text-xs md:text-sm transition-colors whitespace-nowrap"
      style={{ color: hovered ? "#E1E0CC" : "rgba(225,224,204,0.8)" }}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      {label}
    </Link>
  );
}

// Small TH/EN pill: shows the language you'd switch TO (e.g. "EN" while Thai).
function LangToggle() {
  const { lang, setLang, t } = useLang();
  return (
    <button
      type="button"
      onClick={() => setLang(lang === "th" ? "en" : "th")}
      aria-label={t("nav.toggleAria")}
      className="rounded-full border border-white/20 px-2 py-0.5 text-[10px] font-medium tracking-wide text-[#E1E0CC] transition-colors hover:border-white/40 sm:text-xs"
    >
      {t("nav.toggle")}
    </button>
  );
}

export function SiteNav() {
  return (
    <nav className="fixed top-0 left-1/2 -translate-x-1/2 z-50">
      <div className="bg-black rounded-b-2xl md:rounded-b-3xl px-4 py-2 md:px-8">
        <div className="flex items-center gap-3 sm:gap-6 md:gap-12 lg:gap-14">
          {/* Brand logo — links back to the homepage. */}
          <Link href="/" aria-label="AdReady" className="shrink-0">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src="/logo.png"
              alt="AdReady"
              className="h-7 w-7 rounded-full object-cover sm:h-8 sm:w-8"
            />
          </Link>
          {NAV_LINKS.map((item) => (
            <NavLink key={item.label} label={item.label} href={item.href} />
          ))}
          <LangToggle />
        </div>
      </div>
    </nav>
  );
}

export default SiteNav;
