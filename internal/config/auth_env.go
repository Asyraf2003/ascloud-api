package config

import (
	"os"
	"strings"
	"time"
)

func getenvListCSV(k, def string) []string {
	raw := strings.TrimSpace(os.Getenv(k))
	if raw == "" {
		raw = def
	}
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getenvDuration(k, def string) time.Duration {
	raw := strings.TrimSpace(os.Getenv(k))
	if raw == "" {
		raw = def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0
	}
	return d
}

func getenvBool(k string, def bool) bool {
	raw := strings.TrimSpace(os.Getenv(k))
	if raw == "" {
		return def
	}
	switch strings.ToLower(raw) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}
