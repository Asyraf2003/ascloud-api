package domain

import "errors"

var (
	ErrUploadNotFound      = errors.New("upload not found")
	ErrUploadTooLarge      = errors.New("upload too large")
	ErrUploadSiteMismatch  = errors.New("upload site mismatch")
	ErrUploadAlreadyQueued = errors.New("upload already queued")

	ErrSiteNotFound      = errors.New("site not found")
	ErrReleaseNotFound   = errors.New("release not found")
	ErrSiteAlreadyExists = errors.New("hosting site already exists")
	ErrSiteInvalid       = errors.New("hosting site invalid")
)
