package zipsec

import "errors"

var (
	ErrZipSlip      = errors.New("zip slip")
	ErrZipSymlink   = errors.New("zip symlink")
	ErrTooManyFiles = errors.New("zip too many files")
	ErrTooDeep      = errors.New("zip too deep")
	ErrOverQuota    = errors.New("zip over quota")
)
