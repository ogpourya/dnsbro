package rules

import "strings"

// RuleSet holds simple allow/block lists.
type RuleSet struct {
	Blocklist []string
	Allowlist []string
}

// ShouldBlock returns true if the domain should be blocked.
func (r RuleSet) ShouldBlock(domain string) bool {
	d := strings.TrimSuffix(strings.ToLower(domain), ".")
	for _, allow := range r.Allowlist {
		if matchDomain(d, allow) {
			return false
		}
	}
	for _, block := range r.Blocklist {
		if matchDomain(d, block) {
			return true
		}
	}
	return false
}

func matchDomain(domain, rule string) bool {
	r := strings.TrimSuffix(strings.ToLower(rule), ".")
	if r == "" {
		return false
	}
	return domain == r || strings.HasSuffix(domain, "."+r)
}
