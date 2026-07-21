"use client";

// useAudit — audit request state machine for the dashboard
// (Req 1.4, 1.5, 9.5, 14.3, 14.4).
//
// Owns loading/result/error state and remembers the last submitted URL so the
// user can retry with the same URL after an error (Req 14.4).

import { useCallback, useRef, useState } from "react";
import type { AuditResponse, AuditError } from "@mono-repo/shared-types";
import { runAudit as runAuditRequest, toAuditError } from "@/services/auditService";

export interface UseAuditResult {
  loading: boolean;
  result: AuditResponse | null;
  error: AuditError | null;
  runAudit: (url: string) => Promise<void>;
  retry: () => Promise<void>;
}

export function useAudit(): UseAuditResult {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<AuditResponse | null>(null);
  const [error, setError] = useState<AuditError | null>(null);

  // Last submitted URL, so retry() can re-send without re-entering it.
  const lastUrlRef = useRef<string | null>(null);

  const runAudit = useCallback(async (url: string): Promise<void> => {
    // Remember the URL first so a retry works even if the request throws.
    lastUrlRef.current = url;
    setLoading(true);
    setError(null);
    try {
      const response = await runAuditRequest(url);
      setResult(response);
    } catch (err) {
      // Map to a friendly AuditError and clear any stale result (Req 9.5, 14.3).
      setError(toAuditError(err));
      setResult(null);
    } finally {
      setLoading(false);
    }
  }, []);

  const retry = useCallback(async (): Promise<void> => {
    if (lastUrlRef.current === null) return;
    await runAudit(lastUrlRef.current);
  }, [runAudit]);

  return { loading, result, error, runAudit, retry };
}

export default useAudit;
