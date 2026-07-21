"use client";

// Fix_AdSense_Page — help/services landing page (Req 12.1–12.5, 12.7).
//
// Destination of the "แก้ปัญหานี้" (Get Help) buttons. Presents a heading and
// explanatory content plus two Service_Packages (DIY guide + done-for-you)
// each with a prominent price and a CTA. CTAs point at a placeholder contact
// anchor only — there is no payment/checkout processing (Req 12.7).
//
// Dark cinematic palette consistent with the Prisma landing. Static copy is
// localized via useLang() and defaults to Thai when rendered without a
// provider (isolated tests). Package brand names and prices are neutral and
// kept as-is across languages.

import Link from "next/link";
import { useLang } from "@/i18n/LanguageProvider";

// Placeholder contact destination — no real payment flow is wired (Req 12.7).
const CONTACT_URL = "#contact";

interface ServicePackage {
  name: string;
  price: string;
  descKey: string;
  featuresKey: string;
  ctaKey: string;
  recommended: boolean;
}

// The two offered packages (Req 12.3). Names/prices are brand-neutral.
const PACKAGES: ServicePackage[] = [
  {
    name: "DIY Guide",
    price: "฿290",
    descKey: "fix.pkg.diy.desc",
    featuresKey: "fix.pkg.diy.features",
    ctaKey: "fix.pkg.diy.cta",
    recommended: false,
  },
  {
    name: "Done-For-You",
    price: "฿1,990",
    descKey: "fix.pkg.dfy.desc",
    featuresKey: "fix.pkg.dfy.features",
    ctaKey: "fix.pkg.dfy.cta",
    recommended: true,
  },
];

export default function FixAdSensePage() {
  const { t, tList } = useLang();

  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-6">
      <header className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold text-[#E1E0CC] sm:text-3xl">
          {t("fix.title")}
        </h1>
        <p className="text-gray-400">{t("fix.intro")}</p>
      </header>

      {/* At least one explanatory content section (Req 12.2). */}
      <section className="flex flex-col gap-2 rounded-2xl border border-white/10 bg-[#101010] p-5">
        <h2 className="text-lg font-semibold text-[#E1E0CC]">
          {t("fix.help.title")}
        </h2>
        <p className="text-sm text-gray-400">{t("fix.help.body")}</p>
      </section>

      {/* Two Service_Packages with prominent prices and CTAs (Req 12.3, 12.4). */}
      <section className="grid gap-4 sm:grid-cols-2">
        {PACKAGES.map((pkg) => (
          <div
            key={pkg.name}
            className={`flex flex-col gap-3 rounded-2xl border bg-[#212121] p-5 ${
              pkg.recommended ? "border-primary/60" : "border-white/10"
            }`}
          >
            <div className="flex items-center justify-between gap-2">
              <h2 className="text-lg font-semibold text-[#E1E0CC]">
                {pkg.name}
              </h2>
              {pkg.recommended && (
                <span className="rounded-full bg-primary/15 px-2 py-0.5 text-xs font-medium text-primary">
                  {t("fix.recommended")}
                </span>
              )}
            </div>

            {/* Prominent price (Req 12.3). */}
            <p className="text-4xl font-bold text-primary">{pkg.price}</p>

            <p className="text-sm text-gray-400">{t(pkg.descKey)}</p>

            <ul className="flex flex-col gap-1 text-sm text-[#DEDBC8]">
              {tList(pkg.featuresKey).map((feature) => (
                <li key={feature}>• {feature}</li>
              ))}
            </ul>

            {/* CTA → placeholder contact destination, no payment (Req 12.7). */}
            <Link href={CONTACT_URL} className="mt-auto pt-2">
              <span
                className={`block w-full rounded-full px-4 py-2 text-center font-medium transition-colors ${
                  pkg.recommended
                    ? "bg-primary text-black hover:bg-[#E1E0CC]"
                    : "border border-white/15 text-[#E1E0CC] hover:border-white/40"
                }`}
              >
                {t(pkg.ctaKey)}
              </span>
            </Link>
          </div>
        ))}
      </section>

      {/* Secondary link back to the audit dashboard. */}
      <div>
        <Link
          href="/audit"
          className="text-primary underline underline-offset-2 hover:text-[#E1E0CC]"
        >
          {t("fix.back")}
        </Link>
      </div>
    </div>
  );
}
