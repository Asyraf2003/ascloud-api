package http

import (
	"crypto/sha256"
	"encoding/base64"
	"net"
	"strings"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/modules/auth/usecase"
)

func clientInfoFromEcho(c echo.Context) usecase.ClientInfo {
	ua := c.Request().UserAgent()
	uaH := sha256.Sum256([]byte(ua))
	uaHash := base64.RawURLEncoding.EncodeToString(uaH[:])

	dev := strings.TrimSpace(c.Request().Header.Get("X-Device-Id"))
	ip := c.RealIP()
	return usecase.ClientInfo{
		DeviceID:      dev,
		UserAgentHash: uaHash,
		IPPrefix:      ipPrefix(ip),
	}
}

func ipPrefix(ip string) *string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return nil
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil
	}
	if v4 := parsed.To4(); v4 != nil {
		s := net.IPv4(v4[0], v4[1], v4[2], 0).String() + "/24"
		return &s
	}
	s := parsed.String() + "/128"
	return &s
}
