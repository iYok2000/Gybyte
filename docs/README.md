# AdReady — Documentation Index

**AdReady** = AdSense Compliance Auditor. ผู้ใช้กรอก URL → ตรวจเกณฑ์พื้นฐาน Google AdSense
(`ads.txt`, HTTPS/TLS, `robots.txt`, ลิงก์ Privacy/Contact/About) → คืน `auditScore` (0–100 หรือ `null`)
พร้อมรายการผลลัพธ์ + ปุ่ม "แก้ปัญหานี้" สำหรับรายการที่ไม่ผ่าน

## เอกสารในโฟลเดอร์นี้
| ไฟล์ | เนื้อหา |
|---|---|
| [PROJECT.md](./PROJECT.md) | ภาพรวมฟีเจอร์, โครงสร้าง monorepo, **path ที่ต้องทำงานของแต่ละงาน**, คำสั่งรัน |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | สถาปัตยกรรม BE (Hexagonal+CQRS), flow FE, API contract |
| [CSS_STANDARD.md](./CSS_STANDARD.md) | Design tokens + แนวทาง UI/Tailwind v4 |

## Spec ต้นทาง (source of truth)
`.../mono-repo/.kiro/specs/adready-checker/` → `requirements.md`, `design.md`, `tasks.md`
เอกสารในโฟลเดอร์นี้เป็นบทสรุปใช้งานจริง ถ้าขัดแย้งให้ยึด spec

## เริ่มใช้งานเร็ว
```bash
make setup     # ติดตั้ง dependency ครั้งแรก
make rundev    # รัน FE (:3000) + BE (:8080) พร้อมกัน
```
เข้า http://localhost:3000 (redirect → /audit)
