package zipsec

import "errors"

var (
	ErrZipSlip        = errors.New("zip slip")
	ErrZipSymlink     = errors.New("zip symlink")
	ErrTooManyFiles   = errors.New("zip too many files")
	ErrTooDeep        = errors.New("zip too deep")
	ErrOverQuota      = errors.New("zip over quota")
	ErrDisallowedFile = errors.New("zip disallowed file")
)

// ViolationsError allows returning multiple violations while keeping errors.Is() working.
// errors.Is(err, ErrZipSlip) will be true if ErrZipSlip is inside Errs.
type ViolationsError struct {
	Errs []error
}

func (e *ViolationsError) Error() string {
	return "zip violations"
}

func (e *ViolationsError) Unwrap() []error {
	return e.Errs
}
