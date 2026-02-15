package domain

import "errors"

var (
	ErrUploadNotFound      = errors.New("upload not found")
	ErrUploadTooLarge      = errors.New("upload too large")
	ErrUploadSiteMismatch  = errors.New("upload site mismatch")
	ErrUploadAlreadyQueued = errors.New("upload already queued")
)
