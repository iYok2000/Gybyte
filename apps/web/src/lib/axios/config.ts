// Central axios configuration (Req 15.6).
// Base URLs come from public env vars with safe localhost fallbacks so the
// app works out-of-the-box in local development. GO_API points at the Go
// backend; the audit feature actually talks to the same-origin Next.js proxy,
// but this shared config mirrors the source monorepo conventions 1:1.

export const API_CONFIG = {
  // Node/BFF API (used by other domains). Kept for parity with the source repo.
  NODE_API: process.env.NEXT_PUBLIC_NODE_API_URL || "http://localhost:3001",
  // Go backend base URL.
  GO_API: process.env.NEXT_PUBLIC_GO_API_URL || "http://localhost:9000",
  // Default request timeout in milliseconds.
  TIMEOUT: 10000,
} as const;

export type ApiConfig = typeof API_CONFIG;

export default API_CONFIG;
