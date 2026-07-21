package check

import (
	"context"
	"errors"
	"net/url"
	"time"

	"monorepo/backend-go/internal/core/domain/audit/port"
)

// fakeFetcher is a configurable in-memory port.Fetcher for tests. It records
// the URLs it was asked to fetch and returns canned responses/errors so checks
// can be exercised without any real network access.
type fakeFetcher struct {
	fetchResult *port.FetchResult
	fetchErr    error
	tlsInfo     *port.TLSInfo
	tlsErr      error

	fetchedURLs []string
	fetchedHost []string
}

func (f *fakeFetcher) Fetch(_ context.Context, _, u string, _ time.Duration) (*port.FetchResult, error) {
	f.fetchedURLs = append(f.fetchedURLs, u)
	if f.fetchErr != nil {
		return nil, f.fetchErr
	}
	return f.fetchResult, nil
}

func (f *fakeFetcher) FetchTLS(_ context.Context, host string, _ time.Duration) (*port.TLSInfo, error) {
	f.fetchedHost = append(f.fetchedHost, host)
	if f.tlsErr != nil {
		return nil, f.tlsErr
	}
	return f.tlsInfo, nil
}

// errUnreachable is a sentinel used to simulate transport failures.
var errUnreachable = errors.New("connection failed")

// testTarget builds a Target for the given absolute URL string.
func testTarget(raw string) Target {
	u, _ := url.Parse(raw)
	return Target{URL: u, Host: u.Hostname()}
}
