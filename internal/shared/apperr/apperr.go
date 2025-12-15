package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code          string
	HTTPStatus    int
	PublicMessage string
	Cause         error
	Fields        map[string]any
}

func (e *AppError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Cause)
	}
	return e.Code
}

func (e *AppError) Unwrap() error { return e.Cause }

func New(code string, status int, publicMsg string) *AppError {
	if status == 0 {
		status = http.StatusInternalServerError
	}
	if publicMsg == "" {
		publicMsg = "Terjadi kesalahan."
	}
	return &AppError{Code: code, HTTPStatus: status, PublicMessage: publicMsg}
}

func Wrap(cause error, code string, status int, publicMsg string) *AppError {
	ae := New(code, status, publicMsg)
	ae.Cause = cause
	return ae
}

func (e *AppError) WithField(k string, v any) *AppError {
	if e.Fields == nil {
		e.Fields = map[string]any{}
	}
	e.Fields[k] = v
	return e
}

func (e *AppError) WithFields(m map[string]any) *AppError {
	for k, v := range m {
		e.WithField(k, v)
	}
	return e
}

func As(err error) (*AppError, bool) {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae, true
	}
	return nil, false
}
