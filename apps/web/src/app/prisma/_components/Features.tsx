"use client";

import { motion, useInView } from "framer-motion";
import { ArrowRight, Check } from "lucide-react";
import Link from "next/link";
import { useRef, type ReactNode } from "react";
import WordsPullUpMultiStyle from "./WordsPullUpMultiStyle";
import { useLang } from "@/i18n/LanguageProvider";

const CARD_EASE = [0.22, 1, 0.36, 1] as const;

function AnimatedCard({
  index,
  children,
}: {
  index: number;
  children: ReactNode;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const isInView = useInView(ref, { once: true, margin: "-100px" });
  return (
    <motion.div
      ref={ref}
      className="rounded-2xl overflow-hidden h-full flex flex-col"
      initial={{ opacity: 0, scale: 0.95 }}
      animate={isInView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.95 }}
      transition={{ duration: 0.6, delay: index * 0.15, ease: CARD_EASE }}
    >
      {children}
    </motion.div>
  );
}

interface InfoCard {
  icon: string;
  number: string;
  title: string;
  items: string[];
}

// Icon sources kept as-is; copy is localized via the i18n dictionary.
const CARD_ICONS = [
  "https://images.higgs.ai/?default=1&output=webp&url=https%3A%2F%2Fd8j0ntlcm91z4.cloudfront.net%2Fuser_38xzZboKViGWJOttwIXH07lWA1P%2Fhf_20260405_171918_4a5edc79-d78f-4637-ac8b-53c43c220606.png&w=1280&q=85",
  "https://images.higgs.ai/?default=1&output=webp&url=https%3A%2F%2Fd8j0ntlcm91z4.cloudfront.net%2Fuser_38xzZboKViGWJOttwIXH07lWA1P%2Fhf_20260405_171741_ed9845ab-f5b2-4018-8ce7-07cc01823522.png&w=1280&q=85",
  "https://images.higgs.ai/?default=1&output=webp&url=https%3A%2F%2Fd8j0ntlcm91z4.cloudfront.net%2Fuser_38xzZboKViGWJOttwIXH07lWA1P%2Fhf_20260405_171809_f56666dc-c099-4778-ad82-9ad4f209567b.png&w=1280&q=85",
];

export default function Features() {
  const { t, tList } = useLang();

  const INFO_CARDS: InfoCard[] = [
    {
      icon: CARD_ICONS[0],
      number: "01",
      title: t("landing.features.card2.title"),
      items: tList("landing.features.card2.items"),
    },
    {
      icon: CARD_ICONS[1],
      number: "02",
      title: t("landing.features.card3.title"),
      items: tList("landing.features.card3.items"),
    },
    {
      icon: CARD_ICONS[2],
      number: "03",
      title: t("landing.features.card4.title"),
      items: tList("landing.features.card4.items"),
    },
  ];

  return (
    <section className="min-h-screen bg-black relative px-4 md:px-6 py-24">
      <div className="absolute inset-0 bg-noise opacity-[0.15] pointer-events-none" />

      <div className="relative z-10 max-w-7xl mx-auto">
        {/* Crawlable section H2 — the visible header below is animated per-word. */}
        <h2 className="sr-only">{t("landing.seo.features.h2")}</h2>
        <div className="text-xl sm:text-2xl md:text-3xl lg:text-4xl font-normal">
          <WordsPullUpMultiStyle
            segments={[
              {
                text: t("landing.features.header1"),
                className: "",
              },
            ]}
            className="text-[#E1E0CC]"
          />
          <br />
          <WordsPullUpMultiStyle
            segments={[
              {
                text: t("landing.features.header2"),
                className: "text-gray-500",
              },
            ]}
            delay={0.3}
          />
        </div>
      </div>

      <div className="relative z-10 max-w-7xl mx-auto mt-16 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-2 md:gap-1 lg:h-[480px]">
        <AnimatedCard index={0}>
          <div className="relative h-64 lg:h-full">
            <video
              src="https://d8j0ntlcm91z4.cloudfront.net/user_38xzZboKViGWJOttwIXH07lWA1P/hf_20260406_133058_0504132a-0cf3-4450-a370-8ea3b05c95d4.mp4"
              autoPlay
              loop
              muted
              playsInline
              className="absolute inset-0 w-full h-full object-cover"
            />
            <div className="absolute inset-x-0 bottom-0 h-1/2 bg-gradient-to-t from-black/80 to-transparent pointer-events-none" />
            <span
              className="absolute bottom-4 left-4 text-lg font-medium"
              style={{ color: "#E1E0CC" }}
            >
              {t("landing.features.card1.caption")}
            </span>
          </div>
        </AnimatedCard>

        {INFO_CARDS.map((card, i) => (
          <AnimatedCard key={card.number} index={i + 1}>
            <div className="bg-[#212121] p-6 flex flex-col h-64 lg:h-full">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img
                src={card.icon}
                className="w-10 h-10 sm:w-12 sm:h-12 rounded object-cover"
                alt=""
              />
              <div className="mt-4">
                <span className="text-gray-500 text-xs">{card.number}</span>
                <h3
                  className="text-lg font-medium"
                  style={{ color: "#E1E0CC" }}
                >
                  {card.title}
                </h3>
              </div>
              <ul className="mt-4 space-y-2">
                {card.items.map((item) => (
                  <li
                    key={item}
                    className="flex items-start gap-2 text-gray-400 text-xs sm:text-sm"
                  >
                    <Check className="w-4 h-4 text-primary shrink-0 mt-0.5" />
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
              <Link
                href="/services/fix-adsense"
                className="mt-auto inline-flex items-center gap-1 text-primary text-sm"
              >
                {t("landing.features.learnMore")}
                <ArrowRight className="w-4 h-4 -rotate-45" />
              </Link>
            </div>
          </AnimatedCard>
        ))}
      </div>
    </section>
  );
}
