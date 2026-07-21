"use client";

// useCountdown — live 1s countdown for the rate-limit retry timer (Req 9.5).
//
// Given an initial number of seconds, counts down to 0 once per second. Handles
// null/non-positive inputs as 0, and clears the interval on unmount or when the
// initial value changes (leak-safe).

import { useEffect, useState } from "react";

// Clamp any input to a non-negative integer starting point.
function normalize(value: number | null): number {
  if (value === null || !Number.isFinite(value) || value <= 0) {
    return 0;
  }
  return Math.floor(value);
}

/**
 * useCountdown returns the remaining seconds, decrementing every second until
 * it reaches 0.
 */
export function useCountdown(initialSeconds: number | null): number {
  const start = normalize(initialSeconds);

  const [remaining, setRemaining] = useState<number>(start);
  const [prevStart, setPrevStart] = useState<number>(start);

  // Reset when the starting value changes by adjusting state during render
  // instead of synchronously inside an effect. This is the idiomatic React
  // pattern for deriving state from changing inputs, and the guard prevents an
  // infinite render loop.
  if (prevStart !== start) {
    setPrevStart(start);
    setRemaining(start);
  }

  useEffect(() => {
    // Nothing to count down.
    if (start <= 0) {
      return;
    }

    const interval = setInterval(() => {
      setRemaining((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    // Clear on unmount or when the starting value changes.
    return () => clearInterval(interval);
  }, [start]);

  return remaining;
}

export default useCountdown;
