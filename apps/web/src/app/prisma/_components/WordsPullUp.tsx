"use client";

import { motion, useInView } from "framer-motion";
import { useRef } from "react";

interface WordsPullUpProps {
  text: string;
  className?: string;
  showAsterisk?: boolean;
  delay?: number;
}

const EASE = [0.16, 1, 0.3, 1] as const;

export default function WordsPullUp({
  text,
  className,
  showAsterisk = false,
  delay = 0,
}: WordsPullUpProps) {
  const ref = useRef<HTMLSpanElement>(null);
  const isInView = useInView(ref, { once: true });
  const words = text.split(" ");

  return (
    <span ref={ref} className={`inline-flex flex-wrap ${className ?? ""}`}>
      {words.map((word, index) => {
        const isLast = index === words.length - 1;
        const renderWord = () => {
          if (showAsterisk && isLast) {
            const lastAsteriskIndex = word.lastIndexOf("a");
            if (lastAsteriskIndex !== -1) {
              const before = word.slice(0, lastAsteriskIndex);
              const after = word.slice(lastAsteriskIndex + 1);
              return (
                <span className="relative">
                  {before}a{after}
                  <span className="absolute top-[0.65em] -right-[0.3em] text-[0.31em]">
                    *
                  </span>
                </span>
              );
            }
          }
          return word;
        };

        return (
          <motion.span
            key={index}
            className="inline-block"
            initial={{ opacity: 0, y: 20 }}
            animate={isInView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
            transition={{
              duration: 0.5,
              delay: index * 0.08 + delay,
              ease: EASE,
            }}
          >
            {renderWord()}
            {index < words.length - 1 ? "\u00A0" : null}
          </motion.span>
        );
      })}
    </span>
  );
}
