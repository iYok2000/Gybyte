/**
 * Repository-convention ("structure") smoke tests for the standalone AdReady
 * starter kit — frontend side.
 *
 * Feature: adready-checker
 *
 * These tests do not exercise runtime behaviour. They walk the repository tree
 * and assert the standalone-repo conventions from Requirement 15 hold on the
 * frontend:
 *   - the audit feature lives under app/(audit)/audit (feature-scoped)  (Req 15.2)
 *   - the shared contract is imported from @mono-repo/shared-types and   (Req 15.4)
 *     is NOT redeclared anywhere in apps/web
 *
 * The tests skip gracefully when the repo root cannot be resolved, so they
 * never false-fail outside the monorepo layout.
 */

import { describe, it, expect } from "vitest";
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join, sep } from "node:path";

// Resolve the monorepo root by walking upward from this test file until the
// pnpm-workspace.yaml marker is found beside apps/ and packages/.
function findRepoRoot(): string | null {
  let dir = dirname(fileURLToPath(import.meta.url));
  // Guard against infinite loops on unexpected filesystems.
  for (let i = 0; i < 20; i++) {
    const marker = join(dir, "pnpm-workspace.yaml");
    if (
      existsSync(marker) &&
      existsSync(join(dir, "apps")) &&
      existsSync(join(dir, "packages"))
    ) {
      return dir;
    }
    const parent = dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }
  return null;
}

// Recursively collect files under dir whose name ends with one of exts.
function collectFiles(dir: string, exts: string[]): string[] {
  const out: string[] = [];
  if (!existsSync(dir)) return out;
  for (const entry of readdirSync(dir)) {
    if (entry === "node_modules" || entry === ".next") continue;
    const full = join(dir, entry);
    const info = statSync(full);
    if (info.isDirectory()) {
      out.push(...collectFiles(full, exts));
    } else if (exts.some((e) => full.endsWith(e))) {
      out.push(full);
    }
  }
  return out;
}

const root = findRepoRoot();

describe("AdReady standalone-repo structure (frontend)", () => {
  it.skipIf(!root)(
    "keeps the audit feature scoped under app/(audit)/audit (Req 15.2)",
    () => {
      const featureDir = join(
        root!,
        "apps",
        "web",
        "src",
        "app",
        "(audit)",
        "audit",
      );
      expect(existsSync(featureDir), `${featureDir} should exist`).toBe(true);
      // Feature-scoped sub-folders per Req 15.2.
      for (const sub of ["_components", "_hooks", "_utils"]) {
        expect(existsSync(join(featureDir, sub)), `${sub} should exist`).toBe(
          true,
        );
      }
      expect(existsSync(join(featureDir, "page.tsx"))).toBe(true);
    },
  );

  it.skipIf(!root)(
    "imports the audit contract from @mono-repo/shared-types (Req 15.4)",
    () => {
      const service = join(
        root!,
        "apps",
        "web",
        "src",
        "services",
        "auditService.ts",
      );
      expect(existsSync(service)).toBe(true);
      const src = readFileSync(service, "utf8");
      expect(src).toContain("@mono-repo/shared-types");
    },
  );

  it.skipIf(!root)(
    "does not redeclare the shared contract anywhere in apps/web (Req 15.4)",
    () => {
      const webSrc = join(root!, "apps", "web", "src");
      const files = collectFiles(webSrc, [".ts", ".tsx"]).filter(
        // The smoke test itself references the type names in comments; exclude it.
        (f) => !f.endsWith(`${sep}structure.smoke.test.ts`),
      );

      // Redeclaration = defining (not importing) a contract type. We flag
      // top-level `interface`/`type` declarations that reuse the contract
      // names owned by packages/shared-types.
      const contractNames = [
        "AuditStatus",
        "AuditResult",
        "AuditRequest",
        "AuditResponse",
        "AuditError",
      ];
      const offenders: string[] = [];
      for (const file of files) {
        const src = readFileSync(file, "utf8");
        for (const name of contractNames) {
          const declRe = new RegExp(
            `\\b(?:export\\s+)?(?:interface|type)\\s+${name}\\b`,
          );
          if (declRe.test(src)) {
            offenders.push(`${file} redeclares ${name}`);
          }
        }
      }
      expect(offenders, offenders.join("\n")).toEqual([]);
    },
  );
});
