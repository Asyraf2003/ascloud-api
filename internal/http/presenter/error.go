package presenter

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/apperr"
	"example.com/your-api/internal/shared/redact"
)

func requestID(c echo.Context) string {
	if v := c.Response().Header().Get(echo.HeaderXRequestID); v != "" {
		return v
	}
	if v := c.Request().Header.Get(echo.HeaderXRequestID); v != "" {
		return v
	}
	return ""
}

func defaultPublic(status int) (string, string) {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST", "Permintaan tidak valid."
	case http.StatusUnauthorized:
		return "UNAUTHORIZED", "Tidak terautentikasi."
	case http.StatusForbidden:
		return "FORBIDDEN", "Tidak punya akses."
	case http.StatusNotFound:
		return "NOT_FOUND", "Tidak ditemukan."
	case http.StatusTooManyRequests:
		return "RATE_LIMITED", "Terlalu banyak permintaan."
	default:
		return "INTERNAL", "Terjadi kesalahan."
	}
}

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	rid := requestID(c)
	status := http.StatusInternalServerError
	code, msg := defaultPublic(status)

	var fields map[string]any

	if ae, ok := apperr.As(err); ok {
		status = ae.HTTPStatus
		code = ae.Code
		msg = ae.PublicMessage
		fields = ae.Fields
	} else {
		var he *echo.HTTPError
		if errors.As(err, &he) {
			status = he.Code
			code, msg = defaultPublic(status)
		}
	}

	h := redact.RedactHeaders(c.Request().Header)
	if fields != nil {
		fields = redact.RedactMap(fields) // inilah “JSONB/meta” yang disaring dulu
	}

	if status >= 500 {
		c.Logger().Errorf("request_id=%s status=%d code=%s err=%v headers=%v fields=%v",
			rid, status, code, err, h, fields)
	} else {
		c.Logger().Warnf("request_id=%s status=%d code=%s err=%v headers=%v fields=%v",
			rid, status, code, err, h, fields)
	}

	_ = c.JSON(status, ErrorEnvelope{
		Error: ErrorBody{Code: code, Message: msg, RequestID: rid},
	})
}
