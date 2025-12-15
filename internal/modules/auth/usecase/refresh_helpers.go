package usecase

import (
	"strings"

	"example.com/your-api/internal/modules/auth/domain"
)

func aalFromSession(s domain.Session) string {
	if s.Meta == nil {
		return "aal1"
	}
	v, ok := s.Meta["aal"]
	if !ok {
		return "aal1"
	}
	str, ok := v.(string)
	if !ok {
		return "aal1"
	}
	str = strings.ToLower(strings.TrimSpace(str))
	if str == "" {
		return "aal1"
	}
	return str
}
