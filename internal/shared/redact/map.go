package redact

import "strings"

func looksSensitiveKey(k string) bool {
	k = strings.ToLower(strings.TrimSpace(k))
	if k == "" {
		return false
	}
	subs := []string{
		"password", "passwd", "secret", "token", "authorization", "cookie",
		"refresh", "access", "id_token", "apikey", "api_key",
	}
	for _, s := range subs {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}

// RedactMap: sanitize meta map (untuk audit/log fields). Recursive.
func RedactMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		if looksSensitiveKey(k) {
			out[k] = "[REDACTED]"
			continue
		}
		out[k] = redactAny(v)
	}
	return out
}

func redactAny(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return RedactMap(t)
	case []any:
		cp := make([]any, 0, len(t))
		for _, it := range t {
			cp = append(cp, redactAny(it))
		}
		return cp
	default:
		return v
	}
}
