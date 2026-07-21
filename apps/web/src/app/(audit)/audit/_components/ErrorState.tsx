"use client";

// ErrorState — audit failure UI with retry (Req 9.5, 14.3, 14.4).
//
// Rate-limit (HTTP 429) errors show a "too many requests" message plus a live
// countdown, and keep the retry button disabled until the countdown reaches 0
// (Req 9.5). Other errors show a generic failure message. In every case the
// user can retry with the same URL without re-typing it (Req 14.4). Only the
// friendly message from the backend is shown — never internal jargon (Req 14.3).
//
// Static copy is localized via useLang(); the backend `error.message` is shown
// as returned.

import type { AuditError } from "@mono-repo/shared-types";
import { useLang } from "@/i18n/LanguageProvider";
import { useCountdown } from "../_hooks/useCountdown";

export interface ErrorStateProps {
  error: AuditError;
  onRetry: () => void;
}

// A rate-limit error is identified by its code (set by the backend/proxy).
function isRateLimitError(error: AuditError): boolean {
  return error.code === "RATE_LIMIT_EXCEEDED";
}

export function ErrorState({ error, onRetry }: ErrorStateProps) {
  const { t } = useLang();
  const rateLimited = isRateLimitError(error);

  // Hook must be called unconditionally; pass null when not rate-limited so it
  // stays at 0 (Req 9.5).
  const remaining = useCountdown(
    rateLimited && typeof error.retryAfter === "number"
      ? error.retryAfter
      : null,
  );

  const retryDisabled = rateLimited && remaining > 0;

  const heading = rateLimited
    ? t("audit.error.rateLimit")
    : t("audit.error.generic");

  return (
    <div className="rounded-2xl border border-white/10 bg-[#101010] p-6">
      <div className="flex flex-col items-center gap-3 py-2 text-center">
        <p className="text-lg font-medium text-(--error)">{heading}</p>
        <p className="text-sm text-gray-400">{error.message}</p>

        {rateLimited && remaining > 0 && (
          <p className="text-sm text-gray-400">
            {t("audit.error.wait", { n: remaining })}
          </p>
        )}

        <button
          type="button"
          onClick={onRetry}
          disabled={retryDisabled}
          className="inline-flex items-center justify-center rounded-full bg-primary px-6 py-2 font-medium text-black transition-colors hover:bg-[#E1E0CC] disabled:cursor-not-allowed disabled:opacity-60"
        >
          {t("audit.error.retry")}
        </button>
      </div>
    </div>
  );
}

export default ErrorState;
