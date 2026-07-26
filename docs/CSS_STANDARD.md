# CSS / UI Standard

Tailwind CSS **v4** + CSS custom-property design tokens. ไฟล์เดียว: `apps/web/src/app/globals.css`
(`@import "tailwindcss";` + tokens ใน `:root`/`.dark`). Dark mode = next-themes (`attribute="class"`)

## Fonts (รองรับ Thai+Latin)
โหลดผ่าน `next/font/google` ใน `src/app/layout.tsx`, เปิดเป็น CSS variable บน `<html>`:
| ฟอนต์ | บทบาท | variable | ใช้ผ่าน |
|---|---|---|---|
| **Anuphan** (300–700) | primary ทั้งเว็บ (landing + audit) | `--font-sans` | ตั้ง `fontFamily: var(--font-sans)` ที่ wrapper / `@theme` |
| **Noto Serif Thai** (400/600) | accent (หัวข้อ About) | `--font-serif-thai` | utility `font-serif` (globals map `--font-serif` → Thai serif ก่อน) |
| Instrument Serif (italic) | Latin fallback ของ serif | `--font-instrument-serif` | ต่อท้าย stack เท่านั้น |

> **ห้ามใช้ Almarai** (ไม่มี glyph ไทย). การ reveal ต่ออักขระ (About) ต้องตัดแบบ **grapheme** (`Intl.Segmenter`) กันสระ/วรรณยุกต์ไทยแตก และ accent ไทย **ไม่ใช้ `italic`** (Noto Serif Thai ไม่มี italic จริง)

## ⚠️ "primary" มีสองความหมาย (อย่าสับสน)
- **`--primary: #10B981`** (เขียว) = CSS var ธีม AdReady เดิม → ใช้ผ่าน arbitrary syntax `text-(--primary)`, `bg-(--primary)`
- **Tailwind `--color-primary: #DEDBC8`** (ครีม) ใน `@theme` → ใช้ผ่าน utility `text-primary`, `bg-primary`, `text-primary/70` (โทน landing/ครีม)
สองอันนี้คนละตัวและใช้ syntax ต่างกัน

## Design tokens
| Token | Light | Dark | ใช้กับ |
|---|---|---|---|
| `--primary` | `#10B981` | `#10B981` | CTA/brand |
| `--primary-hover` | `#059669` | `#059669` | hover ปุ่ม primary |
| `--primary-soft` | `rgba(16,185,129,.1)` | `.15` | พื้นหลังอ่อน |
| `--success` | `#10B981` | `#10B981` | ผ่าน / score ready |
| `--warning` | `#F59E0B` | `#FBBF24` | score กลาง |
| `--error` | `#EF4444` | `#F87171` | ไม่ผ่าน / score danger |
| `--background` | `#F9FAF8` | `#0A0A0A` | พื้นหน้า |
| `--foreground` | `#1A1A1A` | `#F5F5F5` | ข้อความหลัก |
| `--muted` / `--subtle` | `#525252` / `#737373` | `#A3A3A3` / `#737373` | ข้อความรอง |
| `--card` / `--surface` | `#FFFFFF` | `#1F2937` | พื้น card |
| `--surface-muted` | `#F5F5F2` | `#374151` | พื้น flat |
| `--border` / `--border-hover` | `#E5E7EB` / `#D1D5DB` | `#374151` / `#4B5563` | เส้นขอบ |
| `--focus-ring` | `rgba(16,185,129,.3)` | `.4` | focus ring |
| `--violet` / `--amber` | `#8B5CF6` / `#FBBF24` | `#A78BFA` / `#FBBF24` | accent |

`@theme inline` map `--color-background`/`--color-foreground` → ใช้ `bg-background` / `text-foreground` ได้
และเพิ่ม `--color-primary: #DEDBC8` (utility `text-primary`) + `--font-serif`

## Landing + AdReady dark palette (cinematic ดำ-ครีม)
หน้า `/`, `/audit`, `/services/fix-adsense` ใช้โทนเดียวกัน (ตั้ง bg ดำ + ครีม + Anuphan ที่ wrapper ของแต่ละ group layout):
| ใช้ | ค่า |
|---|---|
| พื้นหลัง | `#000000` / `#0a0a0a` |
| การ์ด | `#101010`, `#212121` |
| ข้อความหลัก | `#E1E0CC` (inline) |
| ข้อความครีมรอง / accent | `text-primary` = `#DEDBC8` |
| ข้อความ muted | `text-gray-400` / `text-gray-500` |
| เส้นขอบ | `border-white/10`, `border-white/20` |
| nav link | `rgba(225,224,204,0.8)` → hover `#E1E0CC` |

### Noise textures (SVG feTurbulence ใน globals.css)
- `.noise-overlay` (baseFrequency 0.85) — เกรนทับวิดีโอ hero (`opacity-[0.7] mix-blend-overlay`)
- `.bg-noise` (baseFrequency 0.9) — พื้นหลัง Features (`opacity-[0.15]`)

## กติกา
- **ใช้ token เสมอ** ผ่าน syntax Tailwind v4: `bg-(--card)`, `text-(--foreground)`, `border-(--border)`, `text-(--error)` — **ห้าม** hardcode สี hex ใน component
- สีตามความหมาย: pass=`--success`, fail=`--error`, warning=`--warning`, brand/CTA=`--primary`
- ทุก interactive element ต้องมี `focus-visible:ring-2 focus-visible:ring-(--focus-ring)`
- ใช้ shared components จาก `src/components/ui` (Card/Button/Badge) แทนการเขียนสไตล์ซ้ำ
- className joiner: `cx()` จาก `src/lib/cx.ts`

## Score bands (ScoreDisplay)
| คะแนน | สี |
|---|---|
| 0–49 | `text-(--error)` (danger) |
| 50–89 | `text-(--warning)` (warning) |
| 90–100 | `text-(--success)` (ready) |
คะแนนเป็นองค์ประกอบใหญ่ที่สุด (`text-7xl`) และอยู่บนสุดของพื้นที่ผลลัพธ์

## Component variants
- **Card**: default / bordered / elevated / flat / glass / premium
- **Button**: primary / secondary / large / small / pill / link / ghost (+ `loading`, `fullWidth`, `icon`)
- **Badge**: default / primary / success / warning / error / outline / violet / glow — ใช้ success/error แยกสถานะ pass/fail

## a11y & SEO headings
- ผลลัพธ์อยู่ใน `aria-live="polite"`
- ScoreDisplay: valid ปกติ, indeterminate = `role="status"`, invalid = `role="alert"`
- แยก pass/fail ด้วยสี **และ** ข้อความ/ไอคอน (ไม่พึ่งสีอย่างเดียว)
- หัวข้อ landing ที่เป็นแอนิเมชันเป็น decorative → มี `<h1/h2 className="sr-only">` keyword-rich คู่กันเพื่อ SEO/screen reader (`sr-only` ของ Tailwind v4 มีให้ใช้ในตัว)
