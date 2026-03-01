package usecase

import (
	"errors"

	"example.com/your-api/internal/modules/hosting/usecase/zipsec"
)

func zipErrCode(err error) string {
	switch {
	case errors.Is(err, zipsec.ErrZipSlip):
		return "hosting.zip_slip"
	case errors.Is(err, zipsec.ErrZipSymlink):
		return "hosting.zip_symlink"
	case errors.Is(err, zipsec.ErrTooManyFiles):
		return "hosting.zip_too_many_files"
	case errors.Is(err, zipsec.ErrTooDeep):
		return "hosting.zip_too_deep"
	case errors.Is(err, zipsec.ErrOverQuota):
		return "hosting.extract_over_quota"
	default:
		return ""
	}
}
