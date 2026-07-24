"use client";

import { motion } from "framer-motion";
import { ArrowRight } from "lucide-react";
import Link from "next/link";
import WordsPullUp from "./WordsPullUp";
import { SiteNav } from "@/components/layout/SiteNav";
import { useLang } from "@/i18n/LanguageProvider";

const EASE = [0.16, 1, 0.3, 1] as const;

export default function Hero() {
  const { t } = useLang();
  return (
    <section className="h-screen p-4 md:p-6">
      {/* Single crawlable H1 for SEO. The big animated "AdReady" wordmark below
          is decorative, so the real, keyword-rich heading is visually hidden. */}
      <h1 className="sr-only">{t("landing.seo.h1")}</h1>
      <div className="relative w-full h-full rounded-2xl md:rounded-[2rem] overflow-hidden">
        <video
          src="/hero.mp4"
          autoPlay
          loop
          muted
          playsInline
          className="absolute inset-0 w-full h-full object-cover"
        />
        <div className="absolute inset-0 noise-overlay opacity-[0.7] mix-blend-overlay pointer-events-none" />
        <div className="absolute inset-0 bg-gradient-to-b from-black/30 via-transparent to-black/60 pointer-events-none" />

        <SiteNav />

        <div className="absolute bottom-0 left-0 right-0 p-6 md:p-10">
          <div className="grid grid-cols-12 gap-4 items-end">
            <div className="col-span-12 md:col-span-8">
              <WordsPullUp
                text="AdReady"
                showAsterisk
                className="text-[17vw] sm:text-[16vw] md:text-[15vw] lg:text-[14vw] xl:text-[13vw] font-medium leading-[0.85] tracking-[-0.07em] text-[#E1E0CC] whitespace-nowrap max-w-full"
              />
            </div>
            <div className="col-span-12 md:col-span-4">
              <motion.p
                className="text-primary/70 text-xs sm:text-sm md:text-base"
                style={{ lineHeight: 1.2 }}
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, delay: 0.5, ease: EASE }}
              >
                {t("landing.hero.description")}
              </motion.p>
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.6, delay: 0.7, ease: EASE }}
                className="mt-6"
              >
                <Link
                  href="/audit"
                  className="group inline-flex items-center gap-2 hover:gap-3 bg-primary text-black rounded-full pl-5 pr-1.5 py-1.5 font-medium text-sm sm:text-base transition-all"
                >
                  <span>{t("landing.hero.cta")}</span>
                  <span className="bg-black rounded-full w-9 h-9 sm:w-10 sm:h-10 flex items-center justify-center transition-transform group-hover:scale-110">
                    <ArrowRight className="w-4 h-4 text-primary" />
                  </span>
                </Link>
              </motion.div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
