"use client";

import { motion, useTransform, type MotionValue } from "framer-motion";

interface AnimatedLetterProps {
  char: string;
  index: number;
  totalChars: number;
  progress: MotionValue<number>;
}

export default function AnimatedLetter({
  char,
  index,
  totalChars,
  progress,
}: AnimatedLetterProps) {
  const charProgress = index / totalChars;
  const opacity = useTransform(
    progress,
    [charProgress - 0.1, charProgress + 0.05],
    [0.2, 1],
  );

  return (
    <motion.span style={{ opacity, whiteSpace: "pre" }}>
      {char === " " ? "\u00A0" : char}
    </motion.span>
  );
}
