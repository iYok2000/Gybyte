# Architecture

## Data flow (end-to-end)
```
Browser (/audit)
  → FE useAudit → auditService.runAudit()
  → POST /api/audit         (Next.js proxy BFF, same-origin)
  → ${NEXT_PUBLIC_GO_API_URL}/api/audit   (Go/Gin :8080)
  → RunAuditHandler.Handle → URLValidator → Engine.Run → checks (concurrent)
  → dto.NewAuditResponse → JSON { auditScore, results }
```
Proxy หน้าที่: forward client IP (X-Forwarded-For) เพื่อ rate limit, relay status/body/Retry-After,
แปลง connection fail/timeout → 502. ไม่ทำ audit logic เอง

## Backend — Hexagonal + CQRS (Query side, stateless)
```
infrastructure/adapter/http
  handler.RunAudit → application/audit/query.RunAuditHandler.Handle
                       ├─ validation.URLValidator (ValidateAndNormalize → GuardSSRF)
                       ├─ core/domain/audit/service.Engine.Run
                       │     └─ check.Check[] (ads.txt, HTTPS, robots.txt, compliance links)
                       │           └─ port.Fetcher  ◄── outbound/http.SafeHTTPClient (adapter)
                       └─ dto.NewAuditResponse (score + schema)
```
- **core ไม่ import `net/http`** — ขึ้นกับ `port.Fetcher` เท่านั้น; `SafeHTTPClient` เป็น adapter ที่ implement
- checks **ห้ามคืน error** → encode failure เป็น fail result (per-check isolation)
- Engine.Run: (1) reachability gate HEAD 10s ก่อน launch goroutine (fail → `ErrSiteUnreachable`, ไม่มี partial); (2) run ทุก check concurrent เขียนลง `slots[idx]` แล้ว flatten ตามลำดับ (deterministic); (3) `runCheck` ห่อ timeout 35s + `recover()` (panic → 1 fail result ไม่ crash)
- `urlValidator` ตัวเดียวถูกแชร์ระหว่าง SafeHTTPClient (redirect/dial guard) และ query handler (input guard)

## SSRF / safety (URLValidator + SafeHTTPClient)
- scheme ∈ {http,https}, ความยาว ≤2048 rune, ตัด control chars + trim
- block: localhost, cloud metadata, loopback/private/link-local/unspecified/multicast, CGNAT, TEST-NET, class-E (v4+v6)
- redirect guard: re-validate ทุก hop (cap 10)
- DNS-rebinding: `guardedDialContext` re-resolve + ตรวจทุก IP แล้ว dial ไป verified IP ตรง ๆ
- body cap 5 MiB, TLS `MinVersion TLS12`, revocation = best-effort

## API contract
`POST /api/audit`  body `{ "url": "https://example.com" }`

**200** — `AuditResponse`
```json
{ "auditScore": 67, "results": [
  { "title": "ads.txt", "status": "fail", "message": "..." },
  { "title": "HTTPS",   "status": "pass", "message": "..." }
] }
```
- `auditScore`: int 0–100 หรือ `null` (ประเมินไม่ได้/results ว่าง)
- `status` ∈ {`pass`,`fail`} เท่านั้น; fail → `message` ไม่ว่าง
- ลำดับ results: ads.txt → HTTPS → robots.txt → Privacy → Contact → About Us

**Errors**
| กรณี | status | body |
|---|---|---|
| URL ผิด/ scheme ไม่ถูก | 400 | `{success:false,error:{code:"VALIDATION_ERROR",message}}` |
| ปลายทางภายใน/redirect ต้องห้าม | 400 | `{...error.code:"BAD_REQUEST"}` |
| เว็บเข้าไม่ถึงทั้งหมด | 400 | BAD_REQUEST (ไม่มี partial) |
| rate limit เกิน | 429 | `{...error:{code:"RATE_LIMIT_EXCEEDED", retryAfter}}` + header `Retry-After` |
| error ภายใน | 500 | `{...error.code:"INTERNAL_ERROR"}` (generic, ไม่ leak) |
| proxy ต่อ backend ไม่ได้/timeout | 502 | `UPSTREAM_UNAVAILABLE` / `UPSTREAM_TIMEOUT` |

## Frontend structure
- feature-scoped: `app/(audit)/audit/{_components,_hooks,_utils}`
- `useAudit`: state loading/result/error + `lastUrlRef`; `retry()` ยิงซ้ำ URL เดิม
- `useCountdown`: นับถอยหลัง retryAfter สด (429) ปิดปุ่ม retry จน 0
- `ScoreDisplay`: loading→spinner; `null`→"ประเมินไม่ได้"; นอกช่วง/ไม่ใช่ int→"คะแนนไม่ถูกต้อง"; valid→เลขใหญ่ + score band color
- ทุก error message เป็นภาษาไทย ไม่ leak รายละเอียดภายใน
