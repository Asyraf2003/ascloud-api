package redact

import (
	"net/http"
	"strings"
)

func MaskToken(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if len(s) <= 8 {
		return "[REDACTED]"
	}
	return s[:4] + "..." + s[len(s)-4:]
}

// RedactHeaders: allowlist beberapa header yang berguna, sisanya diabaikan.
// Token/cookie/api-key selalu dimask.
func RedactHeaders(h http.Header) map[string]string {
	out := map[string]string{}

	// allowlist
	copyIf := func(key string) {
		v := h.Get(key)
		if v != "" {
			out[key] = v
		}
	}

	copyIf("User-Agent")
	copyIf("Content-Type")
	copyIf("X-Request-Id")
	copyIf("X-Forwarded-For")

	// sensitive (mask)
	if v := h.Get("Authorization"); v != "" {
		out["Authorization"] = MaskToken(v)
	}
	if v := h.Get("Cookie"); v != "" {
		out["Cookie"] = "[REDACTED]"
	}
	if v := h.Get("X-Api-Key"); v != "" {
		out["X-Api-Key"] = MaskToken(v)
	}

	return out
}
