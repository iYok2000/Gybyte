package dto

import (
	"encoding/json"
	"strings"
	"testing"

	"monorepo/backend-go/internal/core/domain/audit/valueobject"
)

// TestNewAuditResponse_EmptyIsIndeterminate verifies that an empty result set
// yields a nil AuditScore that serializes to JSON null — not 0 and not 100
// (Req 8.6). This distinguishes "unable to evaluate" from a real score.
func TestNewAuditResponse_EmptyIsIndeterminate(t *testing.T) {
	resp := NewAuditResponse(nil)

	if resp.AuditScore != nil {
		t.Fatalf("expected nil AuditScore for empty results, got %d", *resp.AuditScore)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	if !strings.Contains(string(data), `"auditScore":null`) {
		t.Fatalf("expected auditScore to serialize to null, got: %s", data)
	}
	if strings.Contains(string(data), `"auditScore":0`) || strings.Contains(string(data), `"auditScore":100`) {
		t.Fatalf("auditScore must not be 0 or 100 for an empty audit, got: %s", data)
	}
}

// TestNewAuditResponse_ScoreRounding checks representative scoring cases,
// including the round-half-up boundary (Req 8.5).
func TestNewAuditResponse_ScoreRounding(t *testing.T) {
	pass := valueobject.Result{Title: "p", Status: valueobject.StatusPass, Message: ""}
	fail := valueobject.Result{Title: "f", Status: valueobject.StatusFail, Message: "x"}

	cases := []struct {
		name    string
		results []valueobject.Result
		want    int
	}{
		{"all pass", []valueobject.Result{pass, pass}, 100},
		{"all fail", []valueobject.Result{fail, fail}, 0},
		{"1 of 2 → 50", []valueobject.Result{pass, fail}, 50},
		{"1 of 3 → round-half-up 33", []valueobject.Result{pass, fail, fail}, 33},
		{"2 of 3 → round-half-up 67", []valueobject.Result{pass, pass, fail}, 67},
		{"3 of 8 → 37.5 rounds up to 38", []valueobject.Result{pass, pass, pass, fail, fail, fail, fail, fail}, 38},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := NewAuditResponse(tc.results)
			if resp.AuditScore == nil {
				t.Fatalf("expected non-nil AuditScore for %d results", len(tc.results))
			}
			if *resp.AuditScore != tc.want {
				t.Fatalf("want %d, got %d", tc.want, *resp.AuditScore)
			}
		})
	}
}
