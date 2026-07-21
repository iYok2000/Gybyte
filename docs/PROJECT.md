# Project & Feature Context

## Stack
- **Monorepo**: pnpm workspaces + Turborepo (`pnpm@10.17.1`, `turbo@^2.3.3`)
- **FE** `apps/web`: Next.js 16.0.7 (App Router) · React 19.2 · TypeScript · Tailwind v4 · axios · next-themes
- **BE** `apps/backend-go`: Go 1.25 · Gin v1.11 · Hexagonal + CQRS · **stateless (ไม่มี DB/GORM)** · zap · `golang.org/x/net` (HTML parse)
- **Shared** `packages/shared-types`: สัญญาข้อมูล TS แหล่งเดียว (`@mono-repo/shared-types`)
- **Go module path**: `monorepo/backend-go`
- **Test**: FE = vitest + fast-check · BE = `testing/quick` (stdlib). Property test รัน ≥100 iterations

## Features (สรุป)
1. กรอก/validate URL (FE UX + BE re-validate เสมอ)
2. SSRF + path-traversal + DNS-rebinding guard (BE)
3. ตรวจ 4 กลุ่ม: ads.txt · HTTPS/TLS · robots.txt · compliance links (ไทย/อังกฤษ)
4. คำนวณ `auditScore` (round-half-up, ถ่วงน้ำหนักเท่ากัน; empty → `null`)
5. Rate limiting 60/60s ต่อ IP + `Retry-After` (429) + trusted proxy
6. แสดงคะแนน (score bands) + รายการผล + ปุ่ม "แก้ปัญหานี้"
7. หน้า `/services/fix-adsense` (2 แพ็กเกจ, CTA placeholder ไม่มี payment จริง)
8. Error handling: unreachable → 400, internal → 500, per-check panic isolation

## โครงสร้าง & "งานอยู่ path ไหน"

### Frontend — `apps/web/src/`
| งาน | Path |
|---|---|
| หน้า dashboard (ประกอบทุกอย่าง) | `app/(audit)/audit/page.tsx` |
| ฟอร์ม/คะแนน/รายการผล/ error state | `app/(audit)/audit/_components/` (`AuditForm`, `ScoreDisplay`, `ResultsList`, `ResultItem`, `ErrorState`) |
| hooks (state + countdown) | `app/(audit)/audit/_hooks/` (`useAudit`, `useCountdown`) |
| validate/sanitize URL | `app/(audit)/audit/_utils/` (`validateUrl`, `sanitizeUrl`) |
| เรียก API (axios client) | `src/services/auditService.ts` |
| Proxy BFF → Go backend | `app/api/audit/route.ts` |
| หน้าบริการช่วยเหลือ | `app/services/fix-adsense/page.tsx` |
| root redirect → /audit | `app/page.tsx` |
| shared UI (Card/Button/Badge) | `src/components/ui/` |
| layout/header/nav | `src/app/layout.tsx`, `src/components/layout/` |
| design tokens | `src/app/globals.css` |
| env (ต้องมีเอง — ดูหมายเหตุ) | `apps/web/.env.local` |

### Backend — `apps/backend-go/internal/`
| งาน | Path |
|---|---|
| domain: value objects (Status/Result) | `core/domain/audit/valueobject/` |
| domain: outbound port (Fetcher) | `core/domain/audit/port/` |
| domain: check strategies + Target | `core/domain/audit/check/` (`ads_txt`, `https`, `robots_txt`, `compliance_links`) |
| domain: score | `core/domain/audit/score/` |
| domain: engine (orchestration) | `core/domain/audit/service/engine.go` |
| app: DTO (response schema) | `application/audit/dto/` |
| app: query handler | `application/audit/query/run_audit.go` |
| app: URL validator (SSRF) | `application/audit/validation/url_validator.go` |
| infra: HTTP handler/request | `infrastructure/adapter/http/{handler,request}/` |
| infra: routes/middleware | `infrastructure/adapter/http/{routes,middleware}/` (rate limiter, CORS, security, logger) |
| infra: response envelope | `infrastructure/adapter/http/response/` |
| infra: server bootstrap | `infrastructure/adapter/http/server/http_server.go`, `cmd/server/main.go` |
| infra: SafeHTTPClient (outbound) | `infrastructure/adapter/outbound/http/safe_client.go` |
| config / logger / errors | `internal/config/`, `pkg/logger/`, `pkg/errors/` |
| DI wiring | `internal/container/container.go` |

### Shared contract
`packages/shared-types/src/index.ts` — `AuditStatus`, `AuditResult`, `AuditRequest`, `AuditResponse{auditScore:number|null; results}`, `AuditError`
ห้ามประกาศซ้ำใน FE/BE

## คำสั่ง (Makefile / pnpm)
| คำสั่ง | ทำอะไร |
|---|---|
| `make setup` | `pnpm install` |
| `make rundev` | รัน FE + BE พร้อมกัน (`turbo run dev`) |
| `make web` / `make go` | รันเฉพาะ FE / BE |
| `make build` | `turbo run build` (next build + `go build`) |
| `make lint` | eslint + `go vet` |
| `make test` | vitest + `go test` |

## หมายเหตุสำคัญ (gotchas)
- **env ของ FE**: Next.js อ่าน env จาก `apps/web/` เท่านั้น ไม่อ่าน `.env` ที่ root → ต้องมี `apps/web/.env.local` ตั้ง `NEXT_PUBLIC_GO_API_URL=http://localhost:8080` (แก้ปัญหา proxy fallback ไป :9000) เปลี่ยนแล้วต้อง restart dev server
- **env ของ BE**: อ่าน `../../.env` (root) ผ่าน godotenv; ต้องมี `JWT_SECRET` (≥32 ตัว) มิฉะนั้น `log.Fatal`
- **พอร์ต**: FE :3000, BE :8080 ต้องตรงกันระหว่าง `NEXT_PUBLIC_GO_API_URL` กับ backend `HTTP_PORT`
- FE เรียก backend ผ่าน proxy same-origin `/api/audit` เท่านั้น (ไม่เรียกตรง)
- audit domain เป็น stateless — ห้ามเพิ่ม DB/GORM ในโดเมนนี้
