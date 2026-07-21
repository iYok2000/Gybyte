"use client";

import { useScroll } from "framer-motion";
import { useRef } from "react";
import AnimatedLetter from "./AnimatedLetter";
import WordsPullUpMultiStyle from "./WordsPullUpMultiStyle";

const BODY_TEXT =
  "Over the last seven years, I have worked with Parallax, a Berlin-based production house that crafts cinema, series, and Noir Studio in Paris. Together, we have created work that has earned international acclaim at several major festivals.";

export default function About() {
  const ref = useRef<HTMLParagraphElement>(null);
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start 0.8", "end 0.2"],
  });

  const chars = BODY_TEXT.split("");

  return (
    <section className="bg-black px-4 md:px-6 py-24 md:py-32">
      <div className="bg-[#101010] rounded-2xl md:rounded-[2rem] max-w-6xl mx-auto text-center px-6 py-16 md:py-24">
        <span className="text-primary text-[10px] sm:text-xs uppercase tracking-widest">
          Visual arts
        </span>

        <WordsPullUpMultiStyle
          segments={[
            { text: "I am Marcus Chen,", className: "font-normal" },
            { text: "a self-taught director.", className: "italic font-serif" },
            {
              text: "I have skills in color grading, visual effects, and narrative design.",
              className: "font-normal",
            },
          ]}
          className="text-3xl sm:text-4xl md:text-5xl lg:text-6xl xl:text-7xl max-w-3xl mx-auto leading-[0.95] sm:leading-[0.9] mt-6 text-[#E1E0CC]"
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
