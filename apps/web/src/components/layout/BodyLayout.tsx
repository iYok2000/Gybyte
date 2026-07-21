import type { ReactNode } from "react";

// Shared body wrapper that constrains page content to a consistent width.
// Every page renders through this shell so layout stays uniform (Req 12.5).
export function BodyLayout({ children }: { children: ReactNode }) {
  return (
    <main className="mx-auto w-full max-w-5xl flex-1 px-4 py-8">{children}</main>
  );
}
