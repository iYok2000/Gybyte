# AdReady — Documentation Index

**AdReady** = AdSense Compliance Auditor. ผู้ใช้กรอก URL → ตรวจเกณฑ์พื้นฐาน Google AdSense
(`ads.txt`, HTTPS/TLS, `robots.txt`, ลิงก์ Privacy/Contact/About) → คืน `auditScore` (0–100 หรือ `null`)
พร้อมรายการผลลัพธ์ + ปุ่ม "แก้ปัญหานี้" สำหรับรายการที่ไม่ผ่าน

เว็บประกอบด้วย **หน้า Landing การตลาด** (`/`, ธีม cinematic ดำ-ครีม) + **เครื่องมือตรวจสอบ** (`/audit`)
+ **หน้าบริการ** (`/services/fix-adsense`) รองรับ **สองภาษา (TH/EN)** สลับได้จากปุ่มบน nav

## เอกสารในโฟลเดอร์นี้
| ไฟล์ | เนื้อหา |
|---|---|
| [PROJECT.md](./PROJECT.md) | ภาพรวมฟีเจอร์, โครงสร้าง monorepo, **path ที่ต้องทำงานของแต่ละงาน**, i18n, SEO, คำสั่งรัน |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | สถาปัตยกรรม BE (Hexagonal+CQRS), flow FE, API contract, routes, i18n, SEO |
| [CSS_STANDARD.md](./CSS_STANDARD.md) | ฟอนต์, design tokens, landing palette, แนวทาง UI/Tailwind v4 |

## Spec ต้นทาง (source of truth)
`.../mono-repo/.kiro/specs/adready-checker/` → `requirements.md`, `design.md`, `tasks.md`
เอกสารในโฟลเดอร์นี้เป็นบทสรุปใช้งานจริง (รวมงานหลัง spec เช่น landing/i18n/SEO/branding) ถ้าเรื่องที่ตรงกับ spec ขัดแย้งให้ยึด spec

## เริ่มใช้งานเร็ว
```bash
make setup     # ติดตั้ง dependency ครั้งแรก
make rundev    # รัน FE (:3000) + BE (:8080) พร้อมกัน (turbo คุมทั้งคู่)
```
- `http://localhost:3000` → หน้า Landing (AdReady)
- `http://localhost:3000/audit` → เครื่องมือตรวจสอบ
- `http://localhost:3000/services/fix-adsense` → หน้าบริการ

> ต้องมี `apps/web/.env.local` (ตั้ง `NEXT_PUBLIC_GO_API_URL=http://localhost:8080`) และ root `.env` (มี `JWT_SECRET`) — ดู gotchas ใน PROJECT.md
