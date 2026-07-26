# Project & Feature Context

## Stack
- **Monorepo**: pnpm workspaces + Turborepo (`pnpm@10.17.1`, `turbo@^2.3.3`). ทั้ง `apps/web` และ `apps/backend-go` เป็น workspace package → `turbo run dev/build/lint/test` คุมทั้งคู่
- **FE** `apps/web`: Next.js 16.0.7 (App Router) · React 19.2 · TypeScript · Tailwind v4 · axios · next-themes · **framer-motion** (แอนิเมชัน landing) · lucide-react
- **BE** `apps/backend-go`: Go 1.25 · Gin v1.11 · Hexagonal + CQRS · **stateless (ไม่มี DB/GORM)** · zap · `golang.org/x/net` (HTML parse)
- **Shared** `packages/shared-types`: สัญญาข้อมูล TS แหล่งเดียว (`@mono-repo/shared-types`)
- **ฟอนต์**: **Anuphan** (Thai+Latin, primary, var `--font-sans`) + **Noto Serif Thai** (accent, var `--font-serif-thai`) โหลดผ่าน `next/font/google` ใน `src/app/layout.tsx`
- **i18n**: TH/EN client-side (`src/i18n/LanguageProvider.tsx`), default = TH
- **Go module path**: `monorepo/backend-go`
- **Test**: FE = vitest + fast-check · BE = `testing/quick` (stdlib). Property test รัน ≥100 iterations

## Features (สรุป)
1. **Landing page** (`/`) — การตลาด AdReady, วิดีโอพื้นหลัง, framer-motion, สองภาษา
2. กรอก/validate URL (FE UX + BE re-validate เสมอ)
3. SSRF + path-traversal + DNS-rebinding guard (BE)
4. ตรวจ 4 กลุ่ม: ads.txt · HTTPS/TLS · robots.txt · compliance links (ไทย/อังกฤษ)
5. คำนวณ `auditScore` (round-half-up, ถ่วงน้ำหนักเท่ากัน; empty → `null`)
6. Rate limiting 60/60s ต่อ IP + `Retry-After` (429) + trusted proxy
7. แสดงคะแนน (score bands) + รายการผล (grid) + ปุ่ม "แก้ปัญหานี้"
8. หน้า `/services/fix-adsense` (2 แพ็กเกจ, CTA placeholder ไม่มี payment จริง)
9. Error handling: unreachable → 400, internal → 500, per-check panic isolation
10. **i18n TH/EN** สลับจาก nav · **SEO** (metadata/robots/sitemap/OG/JSON-LD) · **branding** (logo วงกลม + favicon)

## โครงสร้าง & "งานอยู่ path ไหน"

### Frontend — `apps/web/src/`
| งาน | Path |
|---|---|
| **Landing page** (`/`) ประกอบ Hero/About/Features | `app/page.tsx` |
| Landing components (Hero/About/Features + แอนิเมชัน) | `app/prisma/_components/` (`Hero`, `About`, `Features`, `WordsPullUp`, `WordsPullUpMultiStyle`, `AnimatedLetter`) |
| หน้า dashboard ตรวจสอบ (`/audit`) | `app/(audit)/audit/page.tsx` |
| ฟอร์ม/คะแนน/รายการผล/ error state | `app/(audit)/audit/_components/` (`AuditForm`, `ScoreDisplay`, `ResultsList`, `ResultItem`, `ErrorState`) |
| hooks (state + countdown) | `app/(audit)/audit/_hooks/` (`useAudit`, `useCountdown`) |
| validate/sanitize URL | `app/(audit)/audit/_utils/` (`validateUrl`, `sanitizeUrl`) |
| เรียก API (axios client) | `src/services/auditService.ts` |
| Proxy BFF → Go backend | `app/api/audit/route.ts` |
| หน้าบริการช่วยเหลือ (`/services/fix-adsense`) | `app/services/fix-adsense/page.tsx` (+ `layout.tsx` = SEO metadata) |
| **i18n** (dictionary TH/EN + `useLang`/`t`) | `src/i18n/LanguageProvider.tsx` |
| nav (pill ลอย + logo + ปุ่ม TH/EN) | `src/components/layout/SiteNav.tsx` |
| shell layout ต่อกลุ่ม (Header/SiteNav) | `app/(audit)/layout.tsx`, `app/services/layout.tsx`, root `src/app/layout.tsx` |
| shared UI (Card/Button/Badge) + `cx` | `src/components/ui/`, `src/lib/cx.ts` |
| design tokens + noise utilities | `src/app/globals.css` |
| **SEO**: robots / sitemap | `app/robots.ts`, `app/sitemap.ts` |
| **SEO**: OG image | `app/opengraph-image.tsx` |
| **SEO**: JSON-LD (WebSite + SoftwareApplication) | ใน `app/page.tsx` |
| favicon / apple icon (วงกลม จาก logo) | `app/icon.tsx`, `app/apple-icon.tsx` |
| assets: วิดีโอ hero, โลโก้ | `public/hero.mp4`, `public/logo.png` |
| env FE (ต้องมีเอง — ดูหมายเหตุ) | `apps/web/.env.local` |

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
`packages/shared-types/src/index.ts` — `AuditStatus`, `AuditResult`, `AuditRequest`, `AuditResponse{auditScore:number|null; results}`, `AuditError` — ห้ามประกาศซ้ำใน FE/BE

## i18n (TH/EN)
- `LanguageProvider` (context, default `th`, เก็บ localStorage ผ่าน `useSyncExternalStore`) wrap ไว้ที่ root layout
- ใช้งาน: `const { t, tList, lang, setLang } = useLang()` → `t("key")`, `tList("key")` (คืน array)
- เพิ่มข้อความ: แก้ dictionary `th` + `en` ใน `LanguageProvider.tsx` (key เดียวกันทั้งสองภาษา)
- **สำคัญ**: string ที่ vitest อ้างอิง (audit.* / fix.*) ต้องคงค่าไทยเป๊ะ และ default ต้องเป็น `th` (component ที่เทสต์ render โดยไม่มี provider จะ fallback เป็นไทย)
- ข้อความผลลัพธ์ (`message`) มาจาก Go backend เป็นไทย (dynamic) — ยังไม่แปล EN

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
- **SEO base URL**: ตอน deploy ตั้ง `NEXT_PUBLIC_SITE_URL=https://โดเมนจริง` — robots/sitemap/OG/JSON-LD ใช้ค่านี้ (fallback = `http://localhost:3000`)
- **พอร์ต**: FE :3000, BE :8080 ต้องตรงกันระหว่าง `NEXT_PUBLIC_GO_API_URL` กับ backend `HTTP_PORT`
- **favicon/logo**: เปลี่ยนโลโก้แค่ทับ `public/logo.png` แล้ว favicon (`icon.tsx`/`apple-icon.tsx`) อัปเดตตามอัตโนมัติ (crop วงกลม); browser cache หนัก ให้ hard refresh
- FE เรียก backend ผ่าน proxy same-origin `/api/audit` เท่านั้น (ไม่เรียกตรง)
- audit domain เป็น stateless — ห้ามเพิ่ม DB/GORM ในโดเมนนี้
- **ฟอนต์ต้องรองรับไทย**: ใช้ Anuphan/Noto Serif Thai (อย่าใช้ Almarai — ไม่มี glyph ไทย); การ reveal ต่ออักขระใน About ตัดแบบ **grapheme** (`Intl.Segmenter`) กันสระ/วรรณยุกต์ไทยแตก
