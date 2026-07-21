// Package http ให้ SafeHTTPClient ซึ่งเป็น outbound adapter ที่ implement port.Fetcher
// ของโดเมน audit โดยบังคับใช้การป้องกัน SSRF ระดับ dial-time (DNS rebinding),
// จำกัดขนาด body ที่อ่าน และประเมิน TLS จริง — Requirements 2.7, 3.1, 4.1, 4.5, 14.2
//
// ชื่อแพ็กเกจตรงกับไดเรกทอรี (`http`) แต่ผู้ใช้งานควร alias เป็น `outboundhttp`
// เมื่อ import เพื่อเลี่ยงชนกับ stdlib `net/http` ณ จุดเรียกใช้
package http

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"monorepo/backend-go/internal/application/audit/validation"
	"monorepo/backend-go/internal/core/domain/audit/port"
)

// ค่าคงที่เชิงนโยบายของ SafeHTTPClient (ถอดจาก design.md)
const (
	// maxResponseBodyBytes เพดานขนาด body ที่ยอมอ่านต่อหนึ่งคำขอ (5 MiB)
	// กันการดาวน์โหลด response ขนาดใหญ่จนหน่วยความจำหมด
	maxResponseBodyBytes = 5 << 20
	// dialTimeout งบเวลาสูงสุดสำหรับสร้างการเชื่อมต่อ TCP หนึ่งครั้ง
	dialTimeout = 10 * time.Second
	// tlsHandshakeTimeout งบเวลาสูงสุดสำหรับ TLS handshake ของ transport
	tlsHandshakeTimeout = 10 * time.Second
)

// compile-time assert ว่า *SafeHTTPClient สอดคล้องกับ port.Fetcher
var _ port.Fetcher = (*SafeHTTPClient)(nil)

// SafeHTTPClient เป็น adapter ขาออกที่ห่อ net/http.Client พร้อมการ์ด SSRF ระดับ dial
// และ redirect guard ที่แชร์ URLValidator เดียวกันกับเลเยอร์ application
type SafeHTTPClient struct {
	validator *validation.URLValidator
	client    *http.Client
	dialer    *net.Dialer
	// guardIPs ตรวจสอบชุด IP ที่ resolve ได้ก่อน dial (ค่าเริ่มต้น = validator.GuardResolvedIPs)
	// เป็น seam ที่ผ่อนปรนได้เฉพาะในเทสต์ (เพื่อเชื่อมต่อ httptest บน loopback) โปรดักชันคงพฤติกรรมเดิม
	guardIPs func(ips []net.IP) error
}

// NewSafeHTTPClient สร้าง SafeHTTPClient ที่ใช้ guardedDialContext (กัน DNS rebinding)
// และ redirect guard ของ validator โดย "ไม่" ตั้ง Client.Timeout — ใช้ context ต่อคำขอแทน
func NewSafeHTTPClient(validator *validation.URLValidator) *SafeHTTPClient {
	c := &SafeHTTPClient{
		validator: validator,
		dialer:    &net.Dialer{Timeout: dialTimeout},
		guardIPs:  validator.GuardResolvedIPs,
	}

	transport := &http.Transport{
		DialContext:         c.guardedDialContext,
		TLSHandshakeTimeout: tlsHandshakeTimeout,
		ForceAttemptHTTP2:   true,
	}

	c.client = &http.Client{
		Transport:     transport,
		CheckRedirect: validator.NewRedirectGuard(),
		// ไม่ตั้ง Timeout — บังคับกำหนดเวลาผ่าน context ต่อคำขอ (Fetch/FetchTLS)
	}

	return c
}

// Fetch ทำคำขอ HTTP ตาม method/url ภายใน timeout ที่กำหนด แล้วอ่าน body
// สูงสุด maxResponseBodyBytes; ถ้า response มาผ่าน TLS ตั้ง TLS.Valid = true
// (handshake ที่สำเร็จผ่าน transport = Go ตรวจ chain/hostname/validity ให้แล้ว)
// — Requirements 2.7, 3.1, 4.1, 14.2
func (c *SafeHTTPClient) Fetch(ctx context.Context, method, url string, timeout time.Duration) (*port.FetchResult, error) {
	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, method, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	// จำกัดขนาด body ที่อ่านด้วย io.LimitReader กันการดาวน์โหลดขนาดใหญ่เกินเพดาน
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes))
	if err != nil {
		return nil, err
	}

	result := &port.FetchResult{
		StatusCode: resp.StatusCode,
		Body:       body,
		FinalURL:   resp.Request.URL.String(),
	}
	if resp.TLS != nil {
		result.TLS = &port.TLSInfo{Valid: true}
	}

	return result, nil
}

// FetchTLS ทำ TLS handshake ตรงไปยัง host (เติมพอร์ต 443 ถ้าไม่ระบุ) ภายใน timeout
// ผ่าน guardedDialContext (กัน SSRF/rebinding); handshake ล้มเหลว → TLSInfo{Valid:false}
// พร้อมเหตุผลจาก describeTLSError; dial ล้มเหลว → คืน error (→ Req 4.5)
// หมายเหตุ: การตรวจ revocation (Req 4.1 ข้อ ง) เป็น best-effort — Go stdlib ไม่ทำ OCSP/CRL ตอน handshake
// — Requirements 4.1, 4.5
func (c *SafeHTTPClient) FetchTLS(ctx context.Context, host string, timeout time.Duration) (*port.TLSInfo, error) {
	handshakeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	hostname, portStr := splitHostPortDefault(host, "443")
	addr := net.JoinHostPort(hostname, portStr)

	// dial raw TCP ผ่าน guard (re-resolve + ตรวจ IP + dial verified IP)
	rawConn, err := c.guardedDialContext(handshakeCtx, "tcp", addr)
	if err != nil {
		// เชื่อมต่อไม่ได้/ถูกบล็อก/หมดเวลา → error (mapped เป็น fail ที่เลเยอร์ check — Req 4.5)
		return nil, err
	}
	defer func() { _ = rawConn.Close() }()

	// ServerName ต้องเป็น hostname เพื่อให้ Go ตรวจ SNI + hostname matching
	tlsConn := tls.Client(rawConn, &tls.Config{
		ServerName: hostname,
		MinVersion: tls.VersionTLS12,
	})

	if err := tlsConn.HandshakeContext(handshakeCtx); err != nil {
		// handshake ล้มเหลว = ใบรับรองไม่ผ่านเงื่อนไขข้อใดข้อหนึ่ง → Valid:false พร้อมเหตุผล
		return &port.TLSInfo{Valid: false, Reason: describeTLSError(err)}, nil
	}
	_ = tlsConn.Close()

	// handshake สำเร็จ = chain เชื่อถือได้ + hostname ตรง + อยู่ในช่วงวันมีผล
	return &port.TLSInfo{Valid: true}, nil
}

// guardedDialContext เป็น dial hook ที่ป้องกัน DNS rebinding:
// re-resolve host ตอน dial → GuardResolvedIPs ทุก IP → dial ไปที่ "verified IP" ตรง ๆ
// (net.JoinHostPort(ip, port)) ไม่ใช่ hostname เพื่อกัน host ที่ผ่าน guard แล้ว rebind เป็น internal
func (c *SafeHTTPClient) guardedDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	// ถ้า host เป็น IP literal อยู่แล้ว ตรวจตรงแล้ว dial (ไม่ต้อง resolve)
	if ip := net.ParseIP(host); ip != nil {
		if err := c.guardIPs([]net.IP{ip}); err != nil {
			return nil, err
		}
		return c.dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), portStr))
	}

	// re-resolve ชื่อโฮสต์ ณ เวลา dial แล้วตรวจ "ทุก" IP ก่อนต่อ
	resolver := &net.Resolver{}
	addrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, errors.New("no IP addresses resolved for host")
	}

	ips := make([]net.IP, 0, len(addrs))
	for _, a := range addrs {
		ips = append(ips, a.IP)
	}
	if err := c.guardIPs(ips); err != nil {
		return nil, err
	}

	// dial ไปที่ verified IP ตัวแรกโดยตรง (กัน rebinding ระหว่าง resolve กับ dial)
	return c.dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].String(), portStr))
}

// splitHostPortDefault แยก host:port; ถ้าไม่มีพอร์ตติดมา ใช้ defaultPort
// รองรับทั้งชื่อโฮสต์ล้วนและ IP literal (ที่มาจาก url.Hostname() ซึ่งตัดวงเล็บ/พอร์ตออกแล้ว)
func splitHostPortDefault(host, defaultPort string) (string, string) {
	if h, p, err := net.SplitHostPort(host); err == nil {
		return h, p
	}
	// ไม่มีพอร์ต → ใช้ค่าเริ่มต้น
	return host, defaultPort
}

// describeTLSError แปลง error จาก TLS handshake เป็นข้อความเหตุผลที่มนุษย์อ่านได้
// โดยแยกกรณี hostname ไม่ตรง / ใบรับรองหมดอายุ / สายใบรับรองไม่น่าเชื่อถือ — Req 4.3
func describeTLSError(err error) string {
	var hostErr x509.HostnameError
	if errors.As(err, &hostErr) {
		return "ชื่อโดเมนในใบรับรองไม่ตรงกับเว็บไซต์เป้าหมาย"
	}

	var certInvalid x509.CertificateInvalidError
	if errors.As(err, &certInvalid) {
		if certInvalid.Reason == x509.Expired {
			return "ใบรับรอง TLS หมดอายุหรือยังไม่เริ่มมีผลใช้งาน"
		}
		return "ใบรับรอง TLS ไม่ถูกต้อง"
	}

	var unknownAuth x509.UnknownAuthorityError
	if errors.As(err, &unknownAuth) {
		return "ไม่สามารถตรวจสอบสายใบรับรอง (chain of trust) ได้"
	}

	return "การเชื่อมต่อ TLS ล้มเหลว"
}
