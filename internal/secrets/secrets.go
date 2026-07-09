// Package secrets holds the shared secret-detection patterns used both by
// the policy engine's secret-in-args rule (internal/policy, policies/*.yaml)
// and by the audit logger's redaction pass (internal/logger). Keeping the
// patterns in one place avoids two regex sets drifting apart.
package secrets

import "regexp"

// KeyNamePattern matches common secret-bearing key/parameter *names*
// (api_key, secret, token, password, ...). It mirrors the "secret-in-args"
// rule shipped in policies/default.yaml. It flags that a value is likely
// sensitive based on its key, not its shape.
const KeyNamePattern = `(?i)(api[_-]?key|secret|token|password|passwd)`

// valuePatterns match the *shape* of well-known secret values, independent
// of what key they're stored under (e.g. a bare AWS access key pasted into
// a shell command). These are what actually let us redact values in audit
// logs — matching on key name alone can't find a secret value that has no
// "key=" prefix at all.
var valuePatterns = []*regexp.Regexp{
	// AWS access key ID, e.g. AKIAIOSFODNN7EXAMPLE
	regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`),
	// AWS secret access key: 40-char base64-ish, only reliably detectable
	// near an aws_secret_access_key-style label; kept conservative via the
	// key-name pattern above instead of a blanket 40-char regex (too many
	// false positives on hashes/tokens otherwise).

	// OpenAI-style secret keys: sk-..., sk-proj-...
	regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{16,}\b`),
	// Generic Bearer tokens
	regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9._-]{16,}\b`),
	// GitHub tokens: ghp_, gho_, ghu_, ghs_, ghr_, github_pat_
	regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{20,}\b`),
	regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{20,}\b`),
	// Slack tokens: xox[baprs]-...
	regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{10,}\b`),
	// Credentials embedded in a URL: scheme://user:pass@host
	regexp.MustCompile(`\b[a-zA-Z][a-zA-Z0-9+.-]*://[^\s/:@]+:[^\s/:@]+@`),
	// Private key headers (PEM)
	regexp.MustCompile(`-----BEGIN (RSA |EC |OPENSSH |DSA |)PRIVATE KEY-----`),
}

// Redacted is the placeholder written in place of a detected secret.
const Redacted = "[REDACTED]"

// RedactString scans s for known secret value shapes and replaces each
// match with Redacted. It does not use KeyNamePattern (that pattern matches
// key *names*, not values, and blanket-replacing on it would nuke
// non-secret text like the word "token" in a sentence) — value-shape
// patterns only, so redaction is precise.
func RedactString(s string) string {
	for _, re := range valuePatterns {
		s = re.ReplaceAllString(s, Redacted)
	}
	return s
}

// ContainsSecret reports whether s matches any known secret value shape.
func ContainsSecret(s string) bool {
	for _, re := range valuePatterns {
		if re.MatchString(s) {
			return true
		}
	}
	return false
}
