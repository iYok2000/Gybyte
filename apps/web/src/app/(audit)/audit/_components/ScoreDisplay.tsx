"use client";

// ScoreDisplay — prominent audit score presentation (Req 8.6, 10.1–10.5, 10.7).
//
// Evaluation order (important):
//   1. loading            → spinner, never a score (Req 10.4)
//   2. null / undefined    → indeterminate message (not a passing score) via
//                            role="status" (Req 8.6, 10.5)
//   3. NaN / out-of-range / non-integer → invalid message via role="alert",
//                            the bad value is never rendered (Req 10.5)
//   4. valid 0–100 integer → large number, colored by Score_Band
//                            (Req 10.1, 10.2, 10.7)
//
// Static copy is localized via useLang() and defaults to Thai when rendered
// without a LanguageProvider (isolated tests).

import { useLang } from "@/i18n/LanguageProvider";

export interface ScoreDisplayProps {
  /** auditScore from the API: an integer 0–100, or null when indeterminate. */
  score?: number | null;
  /** Whether an audit request is currently in flight. */
  loading?: boolean;
}

export type ScoreBand = "low" | "mid" | "high";

// A valid score is an integer within [0, 100] inclusive.
export function isValidScore(value: number): boolean {
  return Number.isInteger(value) && value >= 0 && value <= 100;
}

// Score_Band for a score, or null when the score is not a valid integer 0–100.
export function scoreBand(score: number | null | undefined): ScoreBand | null {
  if (score === null || score === undefined || !isValidScore(score)) {
    return null;
  }
  if (score < 50) return "low";
  if (score < 90) return "mid";
  return "high";
}

// Score_Band → color token class (Req 10.7). Keeps the "--error"/"--warning"/
// "--success" tokens the tests assert on.
const BAND_CLASS: Record<ScoreBand, string> = {
  low: "text-(--error)",
  mid: "text-(--warning)",
  high: "text-(--success)",
};

// Surface shared by every state of the score box.
const SURFACE =
  "rounded-2xl border border-white/10 bg-[#101010] p-6";

export function ScoreDisplay({ score, loading = false }: ScoreDisplayProps) {
  const { t } = useLang();

  // 1. Loading: show a spinner, never a score (Req 10.4).
  if (loading) {
    return (
      <div className={SURFACE}>
        <div
          role="status"
          aria-live="polite"
          className="flex items-center justify-center gap-3 py-6"
        >
          <span
            aria-hidden="true"
            className="h-6 w-6 animate-spin rounded-full border-2 border-primary border-t-transparent"
          />
          <span className="text-gray-400">{t("audit.score.loading")}</span>
        </div>
      </div>
    );
  }

  // 2. Indeterminate: null (no evaluable items) or undefined (not provided).
  //    Shown as a status, never presented as a passing score (Req 8.6, 10.5).
  if (score === null || score === undefined) {
    return (
      <div className={SURFACE}>
        <p role="status" className="py-6 text-center text-lg text-gray-400">
          {t("audit.score.indeterminate")}
        </p>
      </div>
    );
  }

  // 3. Invalid: NaN, out of range, or non-integer. Surface an error and never
  //    render the bad value (Req 10.5).
  if (!isValidScore(score)) {
    return (
      <div className={SURFACE}>
        <p role="alert" className="py-6 text-center text-lg text-(--error)">
          {t("audit.score.invalid")}
        </p>
      </div>
    );
  }

  // 4. Valid score: the number is the first and largest element in the
  //    results area (Req 10.1), colored by its band (Req 10.7).
  const band = scoreBand(score) as ScoreBand;
  const bandClass = BAND_CLASS[band];
  return (
    <div className={SURFACE}>
      <div className="flex flex-col items-center gap-1 py-2">
        <span className={`text-7xl font-bold leading-none ${bandClass}`}>
          {score}
        </span>
        <span className={`text-sm font-medium ${bandClass}`}>
          {t(`audit.band.${band}`)}
        </span>
        <span className="text-xs text-gray-500">
          {t("audit.score.caption")}
        </span>
      </div>
    </div>
  );
}

export default ScoreDisplay;
