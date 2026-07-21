// Package port declares the outbound ports for the audit domain.
//
// The core depends only on these interfaces, never on net/http, so that the
// audit logic stays free of infrastructure concerns (Req 15.2). Concrete
// adapters (e.g. SafeHTTPClient) live in the infrastructure layer and satisfy
// the Fetcher interface.
package port

import (
	"context"
	"time"
)

// TLSInfo describes the result of a TLS handshake against a target host.
// Valid is true only when the certificate chain, hostname, and validity
// window all check out; Reason explains why Valid is false.
type TLSInfo struct {
	Valid  bool
	Reason string
}

// FetchResult carries the outcome of an outbound HTTP fetch. TLS is non-nil
// when the response was served over a TLS connection. FinalURL reflects the
// URL after any redirects were followed.
type FetchResult struct {
	StatusCode int
	Body       []byte
	TLS        *TLSInfo
	FinalURL   string
}

// Fetcher is the single outbound port the audit domain depends on to reach the
// target website. Implementations MUST enforce the platform's SSRF and body-size
// guards; the core treats this interface as an opaque, safe transport.
type Fetcher interface {
	// Fetch performs an HTTP request with the given method and URL, bounded by
	// the supplied timeout, and returns the response as a FetchResult.
	Fetch(ctx context.Context, method, url string, timeout time.Duration) (*FetchResult, error)
	// FetchTLS performs a TLS handshake against host (bounded by timeout) and
	// reports whether the presented certificate is valid.
	FetchTLS(ctx context.Context, host string, timeout time.Duration) (*TLSInfo, error)
}
