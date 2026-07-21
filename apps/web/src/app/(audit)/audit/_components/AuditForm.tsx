"use client";

// AuditForm — URL entry form for the audit dashboard (Req 1.1, 1.3, 1.4, 1.6).
//
// Owns the input value so the user's text is preserved across failed
// validations (Req 1.3). Runs the UX-only validateUrl guard before submitting;
// the backend re-validates authoritatively. While a request is in flight the
// submit button is disabled and relabelled to prevent duplicate sends (Req 1.4).
//
// Static copy is localized via useLang(); it defaults to Thai (byte-identical
// to the original strings) when rendered without a LanguageProvider.

import { useState, type FormEvent } from "react";
import { sanitizeUrl } from "../_utils/sanitizeUrl";
import { validateUrl } from "../_utils/validateUrl";
import { useLang } from "@/i18n/LanguageProvider";

export interface AuditFormProps {
  /** Called with the sanitized URL when validation passes. */
  onSubmit: (url: string) => void;
  /** Whether an audit request is currently in flight. */
  loading?: boolean;
}

export function AuditForm({ onSubmit, loading = false }: AuditFormProps) {
  const { t } = useLang();

  // Controlled input — retains the raw text even after a rejected submit.
  const [value, setValue] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  function handleSubmit(event: FormEvent<HTMLFormElement>): void {
    // noValidate on the form means we own all validation here.
    event.preventDefault();

    const result = validateUrl(value);
    if (!result.valid && result.reason) {
      // Show the specific reason and do NOT send the request; input is kept.
      setErrorMessage(t(`audit.form.error.${result.reason}`));
      return;
    }

    setErrorMessage(null);
    // Submit the sanitized value so the backend receives the normalized URL.
    onSubmit(sanitizeUrl(value));
  }

  return (
    <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-3">
      {/* Input and submit button are visible together on screen (Req 1.1). */}
      <div className="flex flex-col gap-3 sm:flex-row">
        <input
          type="text"
          inputMode="url"
          name="url"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          maxLength={2048}
          disabled={loading}
          placeholder="https://example.com"
          aria-label={t("audit.form.inputLabel")}
          aria-invalid={errorMessage != null || undefined}
          aria-describedby={errorMessage ? "audit-url-error" : undefined}
          className="flex-1 rounded-full border border-white/10 bg-[#101010] px-5 py-2.5 text-[#E1E0CC] placeholder:text-gray-500 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary/40 disabled:opacity-60"
        />
        <button
          type="submit"
          disabled={loading}
          aria-busy={loading || undefined}
          className="inline-flex items-center justify-center gap-2 rounded-full bg-primary px-6 py-2.5 font-medium text-black transition-colors hover:bg-[#E1E0CC] disabled:cursor-not-allowed disabled:opacity-60"
        >
          {loading && (
            <span
              aria-hidden="true"
              className="h-4 w-4 animate-spin rounded-full border-2 border-black/40 border-t-transparent"
            />
          )}
          {loading ? t("audit.form.submitting") : t("audit.form.submit")}
        </button>
      </div>

      {errorMessage && (
        <p id="audit-url-error" role="alert" className="text-sm text-red-400">
          {errorMessage}
        </p>
      )}
    </form>
  );
}

export default AuditForm;
