// Package check defines the Audit_Check strategy interface and the checks that
// evaluate a target website against AdSense readiness criteria.
package check

import "net/url"

// Target is the validated, normalized subject of an audit run.
//
// It is placed in package check (rather than service) to avoid an import cycle:
// the engine in package service depends on package check for the Check
// interface, while each Check needs the Target type. Keeping Target here lets
// checks reference it without service importing check and check importing
// service.
type Target struct {
	// URL is the fully validated and normalized target URL.
	URL *url.URL
	// Host is the target hostname (URL.Hostname()), used for TLS and origin
	// composition.
	Host string
}
