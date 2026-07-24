"use client";

import { useScroll } from "framer-motion";
import { useRef } from "react";
import AnimatedLetter from "./AnimatedLetter";
import WordsPullUpMultiStyle from "./WordsPullUpMultiStyle";
import { useLang } from "@/i18n/LanguageProvider";

// Split text into grapheme clusters (locale-aware when Intl.Segmenter is
// available) so Thai combining marks stay attached to their base character.
// Falls back to Array.from (code points) on environments without Segmenter.
function splitGraphemes(text: string, lang: string): string[] {
  if (typeof Intl !== "undefined" && "Segmenter" in Intl) {
    const segmenter = new Intl.Segmenter(lang, { granularity: "grapheme" });
    return Array.from(segmenter.segment(text), (s) => s.segment);
  }
  return Array.from(text);
}

export default function About() {
  const { t, tList, lang } = useLang();
  const ref = useRef<HTMLParagraphElement>(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start 0.8", "end 0.2"],
  });

  const headingSegments = tList("landing.about.heading");
  // Split into grapheme clusters so Thai combining marks (vowels/tone marks)
  // stay attached to their base character instead of becoming separate spans.
  const body = t("landing.about.body");
  const chars = splitGraphemes(body, lang);

  return (
    <section className="bg-black px-4 md:px-6 py-24 md:py-32">
      <div className="bg-[#101010] rounded-2xl md:rounded-[2rem] max-w-6xl mx-auto text-center px-6 py-16 md:py-24">
        {/* Crawlable section H2 — the visible heading below is animated per-word,
            so a real keyword-rich H2 is provided for SEO. */}
        <h2 className="sr-only">{t("landing.seo.about.h2")}</h2>
        <span className="text-primary text-[10px] sm:text-xs uppercase tracking-widest">
          {t("landing.about.label")}
        </span>

        <WordsPullUpMultiStyle
          segments={[
            { text: headingSegments[0] ?? "", className: "font-normal" },
            {
              text: headingSegments[1] ?? "",
              className: "font-serif text-primary",
            },
            { text: headingSegments[2] ?? "", className: "font-normal" },
          ]}
          className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl xl:text-7xl max-w-3xl mx-auto leading-tight sm:leading-tight mt-6 text-[#E1E0CC]"
        />

        <p
          ref={ref}
          className="text-[#DEDBC8] text-xs sm:text-sm md:text-base max-w-2xl mx-auto mt-10 leading-relaxed"
          style={{ whiteSpace: "pre-wrap" }}
        >
          {chars.map((char, index) => (
            <AnimatedLetter
              key={index}
              char={char}
              index={index}
              totalChars={chars.length}
              progress={scrollYProgress}
            />
          ))}
        </p>
      </div>
    </section>
  );
}
