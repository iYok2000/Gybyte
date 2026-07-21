"use client";

// ResultsList — renders the audit result cards in order (Req 11.1, 11.6).
//
// Preserves the exact order of the `results` array (deterministic engine
// order). A responsive grid (row-major) keeps that order while fitting more
// cards above the fold and reducing vertical scroll. When the array is empty,
// shows a flat "no results" card with no Get Help button (Req 11.6).

import type { AuditResult } from "@mono-repo/shared-types";
import { useLang } from "@/i18n/LanguageProvider";
import { ResultItem } from "./ResultItem";

export interface ResultsListProps {
  results: AuditResult[];
}

export function ResultsList({ results }: ResultsListProps) {
  const { t } = useLang();

  // Empty results: informative flat card, no Get Help button (Req 11.6).
  if (results.length === 0) {
    return (
      <div className="rounded-xl border border-white/10 bg-[#101010] p-4">
        <p className="py-4 text-center text-gray-400">
          {t("audit.result.empty")}
        </p>
      </div>
    );
  }

  return (
    <ul className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {results.map((result, index) => (
        // Order-preserving key; index disambiguates duplicate titles (Req 11.1).
        <li key={`${result.title}-${index}`} className="h-full">
          <ResultItem result={result} />
        </li>
      ))}
    </ul>
  );
}

export default ResultsList;
