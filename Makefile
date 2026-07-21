# AdReady — root Makefile
# NOTE: การตรวจสอบ (audit) เป็น stateless ไม่มี DB/migration
# แต่คงรูปแบบ (shape) ของ Makefile ต้นทางไว้เพื่อ parity

.PHONY: setup rundev web go build lint test clean

# ติดตั้ง dependency ทั้ง workspace
setup:
	pnpm install

# รันทุกแอปพร้อมกัน (turbo run dev) — เทียบเท่า `pnpm dev`
rundev:
	pnpm dev

# รันเฉพาะ frontend (apps/web)
web:
	pnpm web

# รันเฉพาะ backend (apps/backend-go)
go:
	pnpm go

# build ทั้งหมด (next build + go build ผ่าน turbo)
build:
	pnpm build

# lint ทั้งหมด
lint:
	pnpm lint

# รันเทสต์ทั้งหมด
test:
	pnpm turbo run test

# ล้าง build artifacts
clean:
	pnpm clean
