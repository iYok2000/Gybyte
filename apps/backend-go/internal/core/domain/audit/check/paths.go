package check

// Fixed, pre-defined resource paths. Every check composes the resource URL from
// the target's origin (scheme + host) plus one of these constant paths, never
// from user-controlled path/query/fragment segments. This closes off path
// traversal via the submitted URL (Req 2.6).
const (
	adsTxtPath    = "/ads.txt"
	robotsTxtPath = "/robots.txt"
	homepagePath  = "/"
)

// originURL builds a resource URL from the target ORIGIN only:
//
//	scheme://host + fixedPath
//
// Any path, query, or fragment the user supplied on the submitted URL is
// dropped. fixedPath MUST be one of the pre-defined constants above (Req 2.6).
func originURL(target Target, fixedPath string) string {
	return target.URL.Scheme + "://" + target.Host + fixedPath
}
