"use client";

// ResultItem — a single audit check card (Req 11.1, 11.2, 11.3, 11.4, 11.5).
//
// Compact by design so many results fit above the fold: a colored badge lets
// the user tell pass from fail without reading the message (Req 11.2), and
// failing checks get a small inline "แก้ปัญหานี้" (Get Help) link/button to the
// Fix AdSense page; passing checks get no button (Req 11.3–11.5).
//
// The check `title` and `message` come from the backend and are left as-is;
// only the FE-owned badge/button labels are localized.

import Link from "next/link";
import type { AuditResult } from "@mono-repo/shared-types";
import { useLang } from "@/i18n/LanguageProvider";

export interface ResultItemProps {
  result: AuditResult;
}

export function ResultItem({ result }: ResultItemProps) {
  const { t } = useLang();
  const isFail = result.status === "fail";

  return (
    <div className="flex h-full flex-col gap-2 rounded-xl border border-white/10 bg-[#212121] p-4">
      <div className="flex items-center gap-2">
        <span
          className={`inline-flex shrink-0 items-center rounded-full px-2 py-0.5 text-xs font-medium ${
            isFail
              ? "bg-(--error)/15 text-(--error)"
              : "bg-(--success)/15 text-(--success)"
          }`}
        >
          {isFail ? t("audit.result.fail") : t("audit.result.pass")}
        </span>
        <h3 className="truncate font-medium text-[#E1E0CC]">{result.title}</h3>
      </div>

      {result.message && (
        <p className="text-sm leading-snug text-gray-400">{result.message}</p>
      )}

      {/* Get Help link only for failing checks (Req 11.3, 11.4, 11.5). */}
      {isFail && (
        <Link
          href="/services/fix-adsense"
          className="mt-auto inline-flex w-fit"
        >
          <button
            type="button"
            className="rounded-full border border-primary/40 px-3 py-1 text-xs font-medium text-primary transition-colors hover:bg-primary hover:text-black"
          >
            {t("audit.result.getHelp")}
          </button>
        </Link>
      )}
    </div>
  );
}

export default ResultItem;
