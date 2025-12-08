package rules

import "testing"

func TestShouldBlock(t *testing.T) {
	rs := RuleSet{
		Blocklist: []string{"example.com", "ads.test"},
		Allowlist: []string{"allowed.example.com"},
	}

	cases := []struct {
		domain string
		block  bool
	}{
		{"example.com.", true},
		{"www.example.com", true},
		{"allowed.example.com", false},
		{"sub.allowed.example.com", false},
		{"ads.test", true},
		{"safe.com", false},
	}

	for _, c := range cases {
		if got := rs.ShouldBlock(c.domain); got != c.block {
			t.Fatalf("domain %s expected block=%v got %v", c.domain, c.block, got)
		}
	}
}
