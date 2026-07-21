"use client";

// Audit_Dashboard page — wires the form, request state, and result views
// (Req 1.2, 9.5, 10.6, 14.3, 14.4).
//
// The result area is an aria-live="polite" region so screen readers announce
// the score/results/errors as they arrive. On error it renders ErrorState
// (with rate-limit countdown + same-URL retry); otherwise it shows a compact
// top summary (big score + band + passed count) followed by the ordered
// results grid, so the important information stays above the fold.

import { AuditForm } from "./_components/AuditForm";
import { ScoreDisplay, scoreBand } from "./_components/ScoreDisplay";
import { ResultsList } from "./_components/ResultsList";
import { ErrorState } from "./_components/ErrorState";
import { useAudit } from "./_hooks/useAudit";
import { useLang } from "@/i18n/LanguageProvider";

export default function AuditPage() {
  const { loading, result, error, runAudit, retry } = useAudit();
  const { t } = useLang();

  const band = result ? scoreBand(result.auditScore) : null;
  const total = result?.results.length ?? 0;
  const passed = result?.results.filter((r) => r.status === "pass").length ?? 0;

  return (
    <div className="flex flex-col gap-6">
      <header className="flex flex-col gap-2">
        <h1 className="text-2xl font-bold text-[#E1E0CC] sm:text-3xl">
          {t("audit.page.title")}
        </h1>
        <p className="text-gray-400">{t("audit.page.subtitle")}</p>
      </header>

      <AuditForm onSubmit={runAudit} loading={loading} />

      {/* Live region: announces score, results, and errors as they update. */}
      <section aria-live="polite" className="flex flex-col gap-4">
        {loading && <ScoreDisplay loading />}

        {!loading && error && <ErrorState error={error} onRetry={retry} />}

        {!loading && !error && result && (
          <>
            {/* Compact top summary: big score beside a short readiness line. */}
            <div className="grid gap-4 md:grid-cols-[auto_1fr] md:items-center">
              <ScoreDisplay score={result.auditScore} />
              <div className="flex flex-col justify-center gap-2 rounded-2xl border border-white/10 bg-[#101010] p-6">
                {band && (
                  <span className="text-xl font-semibold text-[#E1E0CC]">
                    {t(`audit.band.${band}`)}
                  </span>
                )}
                <span className="text-2xl font-bold text-primary">
                  {t("audit.page.passed", { passed, total })}
                </span>
              </div>
            </div>

            <ResultsList results={result.results} />
          </>
        )}
      </section>
    </div>
  );
}
