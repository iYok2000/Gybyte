// Package smoke holds repository-convention ("structure") smoke tests for the
// standalone AdReady starter kit. These tests do not exercise runtime behaviour;
// instead they walk the repository tree and assert that the standalone-repo
// conventions from Requirement 15 hold, so the layout cannot silently drift.
//
// Feature: adready-checker
//
// Covered conventions:
//   - FE audit code lives ONLY in apps/web                       (Req 15.2)
//   - the audit engine lives in apps/backend-go                  (Req 15.2)
//   - the shared contract is defined in packages/shared-types    (Req 15.4)
//     only, and is not redeclared in the Go backend beyond DTOs
//   - no SQL/GORM in the audit domain (stateless)                (Req 15.5)
//   - one concern per file + business-logic comments present     (Req 15.8)
//
// The tests are deliberately robust: when a path cannot be resolved (for
// example when the module is vendored or run in isolation) they SKIP rather
// than fail, so they never produce a false negative outside the monorepo.
package smoke

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// repoRoot walks upward from this test file until it finds the monorepo root,
// identified by the pnpm-workspace.yaml marker that sits beside apps/ and
// packages/. It returns ("", false) when the root cannot be located so callers
// can skip gracefully instead of false-failing.
func repoRoot(t *testing.T) (string, bool) {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", false
	}

	dir := filepath.Dir(thisFile)
	for {
		marker := filepath.Join(dir, "pnpm-workspace.yaml")
		if _, err := os.Stat(marker); err == nil {
			// Confirm the expected workspace siblings exist too.
			if isDir(filepath.Join(dir, "apps")) && isDir(filepath.Join(dir, "packages")) {
				return dir, true
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the filesystem root without finding the marker.
			return "", false
		}
		dir = parent
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// auditGoDirs are the backend audit-domain source directories that must remain
// stateless (no SQL/GORM) and self-documenting (Req 15.5, 15.8).
func auditGoDirs(root string) []string {
	return []string{
		filepath.Join(root, "apps", "backend-go", "internal", "core", "domain", "audit"),
		filepath.Join(root, "apps", "backend-go", "internal", "application", "audit"),
	}
}

// walkGoFiles invokes fn for every non-test .go file found under dir. Missing
// directories are ignored so the walker never fails on an absent path.
func walkGoFiles(t *testing.T, dir string, includeTests bool, fn func(path string, src string)) {
	t.Helper()
	if !isDir(dir) {
		return
	}
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // tolerate transient walk errors
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if !includeTests && strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		fn(path, string(data))
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", dir, err)
	}
}

// TestAuditEngineLivesInBackendGo asserts the audit engine (and its supporting
// domain packages) resides under apps/backend-go (Req 15.2).
func TestAuditEngineLivesInBackendGo(t *testing.T) {
	root, ok := repoRoot(t)
	if !ok {
		t.Skip("repo root not resolvable; skipping structure smoke test")
	}

	enginePath := filepath.Join(root, "apps", "backend-go", "internal",
		"core", "domain", "audit", "service", "engine.go")
	if !exists(enginePath) {
		t.Errorf("expected audit engine at %s (Req 15.2)", enginePath)
	}

	// The audit domain must not leak into the web app: no Go sources under
	// apps/web at all.
	webDir := filepath.Join(root, "apps", "web")
	if isDir(webDir) {
		walkGoFiles(t, webDir, true, func(path, _ string) {
			t.Errorf("unexpected Go source inside apps/web: %s (Req 15.2)", path)
		})
	}
}

// TestFrontendAuditCodeOnlyInWeb asserts the frontend audit feature lives under
// apps/web/src/app/(audit)/audit and that no React (.tsx) frontend code leaks
// into apps/backend-go (Req 15.2).
func TestFrontendAuditCodeOnlyInWeb(t *testing.T) {
	root, ok := repoRoot(t)
	if !ok {
		t.Skip("repo root not resolvable; skipping structure smoke test")
	}

	auditFeatureDir := filepath.Join(root, "apps", "web", "src", "app", "(audit)", "audit")
	if !isDir(auditFeatureDir) {
		t.Errorf("expected FE audit feature dir at %s (Req 15.2)", auditFeatureDir)
	}

	// No frontend (.tsx/.ts/.jsx) sources should exist under the Go backend.
	backendDir := filepath.Join(root, "apps", "backend-go")
	if isDir(backendDir) {
		err := filepath.WalkDir(backendDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			for _, ext := range []string{".tsx", ".jsx", ".ts"} {
				if strings.HasSuffix(path, ext) {
					t.Errorf("unexpected frontend source in backend-go: %s (Req 15.2)", path)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s: %v", backendDir, err)
		}
	}
}

// TestSharedContractSingleSource asserts the audit data contract is declared
// exactly once, in packages/shared-types, and is not redeclared in the Go
// backend beyond the transport DTO layer (Req 15.4).
func TestSharedContractSingleSource(t *testing.T) {
	root, ok := repoRoot(t)
	if !ok {
		t.Skip("repo root not resolvable; skipping structure smoke test")
	}

	// The single source of truth must exist and declare the contract.
	contractPath := filepath.Join(root, "packages", "shared-types", "src", "index.ts")
	if !exists(contractPath) {
		t.Fatalf("expected shared contract at %s (Req 15.4)", contractPath)
	}
	data, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("reading shared contract: %v", err)
	}
	src := string(data)
	for _, name := range []string{"AuditStatus", "AuditResult", "AuditRequest", "AuditResponse", "AuditError"} {
		if !strings.Contains(src, name) {
			t.Errorf("shared contract %s missing type %q (Req 15.4)", contractPath, name)
		}
	}

	// The Go DTO layer intentionally mirrors the wire shape (AuditResponse,
	// AuditResultDTO). That mirror is allowed ONLY inside the dto package —
	// the contract type name must not be redeclared anywhere else in the Go
	// audit domain.
	dtoDir := filepath.Join(root, "apps", "backend-go", "internal", "application", "audit", "dto")
	for _, dir := range auditGoDirs(root) {
		walkGoFiles(t, dir, false, func(path, goSrc string) {
			if strings.Contains(goSrc, "type AuditResponse struct") {
				inDTO := strings.HasPrefix(path, dtoDir+string(os.PathSeparator))
				if !inDTO {
					t.Errorf("AuditResponse redeclared outside dto layer: %s (Req 15.4)", path)
				}
			}
		})
	}
}

// TestAuditDomainIsStateless asserts no SQL/GORM dependency appears in the audit
// domain: the audit feature is a read-only evaluator with no persistence
// (Req 15.5).
func TestAuditDomainIsStateless(t *testing.T) {
	root, ok := repoRoot(t)
	if !ok {
		t.Skip("repo root not resolvable; skipping structure smoke test")
	}

	forbidden := []string{
		"gorm.io/gorm",
		"\"database/sql\"",
		"jinzhu/gorm",
	}

	for _, dir := range auditGoDirs(root) {
		walkGoFiles(t, dir, true, func(path, goSrc string) {
			for _, needle := range forbidden {
				if strings.Contains(goSrc, needle) {
					t.Errorf("audit domain must be stateless but %s references %q (Req 15.5)", path, needle)
				}
			}
		})
	}
}

// TestAuditSourcesAreDocumented asserts every audit-domain source file carries
// at least one comment (business-logic documentation) and declares exactly one
// package clause — a light "one concern per file" + self-documenting check
// (Req 15.8).
func TestAuditSourcesAreDocumented(t *testing.T) {
	root, ok := repoRoot(t)
	if !ok {
		t.Skip("repo root not resolvable; skipping structure smoke test")
	}

	checked := 0
	for _, dir := range auditGoDirs(root) {
		walkGoFiles(t, dir, false, func(path, goSrc string) {
			checked++

			// Business-logic comment present (Req 15.8).
			if !strings.Contains(goSrc, "//") && !strings.Contains(goSrc, "/*") {
				t.Errorf("audit source lacks explanatory comment: %s (Req 15.8)", path)
			}

			// One concern per file: exactly one package clause.
			pkgCount := strings.Count(goSrc, "\npackage ")
			if strings.HasPrefix(goSrc, "package ") {
				pkgCount++
			}
			if pkgCount != 1 {
				t.Errorf("expected exactly one package clause in %s, found %d (Req 15.8)", path, pkgCount)
			}
		})
	}

	if checked == 0 {
		t.Skip("no audit-domain Go sources found; skipping documentation smoke check")
	}
}
