"use client";

import { motion, useInView } from "framer-motion";
import { useRef } from "react";

interface Segment {
  text: string;
  className?: string;
}

interface WordsPullUpMultiStyleProps {
  segments: Segment[];
  className?: string;
  delay?: number;
}

const EASE = [0.16, 1, 0.3, 1] as const;

export default function WordsPullUpMultiStyle({
  segments,
  className,
  delay = 0,
}: WordsPullUpMultiStyleProps) {
  const ref = useRef<HTMLSpanElement>(null);
  const isInView = useInView(ref, { once: true });

  const words = segments.flatMap((segment) =>
    segment.text
      .split(" ")
      .filter((word) => word.length > 0)
      .map((word) => ({ word, className: segment.className })),
  );

  return (
    <span
      ref={ref}
      className={`inline-flex flex-wrap justify-center ${className ?? ""}`}
    >
      {words.map(({ word, className: wordClassName }, index) => (
        <motion.span
          key={index}
          className={`inline-block ${wordClassName ?? ""}`}
          initial={{ opacity: 0, y: 20 }}
          animate={isInView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
          transition={{
            duration: 0.5,
            delay: index * 0.08 + delay,
            ease: EASE,
          }}
        >
          {word}
          {index < words.length - 1 ? "\u00A0" : null}
        </motion.span>
      ))}
    </span>
  );
}
