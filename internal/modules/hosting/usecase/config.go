package usecase

import "time"

type Config struct {
	PresignTTL     time.Duration
	MaxUploadBytes int64
}

func DefaultConfig() Config {
	return Config{
		PresignTTL:     15 * time.Minute,
		MaxUploadBytes: 20 * 1024 * 1024, // 20MB (best-effort; extract quota enforced in Step 5)
	}
}
