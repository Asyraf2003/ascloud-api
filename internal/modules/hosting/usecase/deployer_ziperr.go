package usecase

import (
	"errors"

	"example.com/your-api/internal/modules/hosting/usecase/zipsec"
)

func zipErrCodes(err error) []string {
	// severity order (primary = first match)
	type m struct {
		e    error
		code string
	}
	maps := []m{
		{zipsec.ErrZipSlip, "hosting.zip_slip"},
		{zipsec.ErrZipSymlink, "hosting.zip_symlink"},
		{zipsec.ErrDisallowedFile, "hosting.file_disallowed"},
		{zipsec.ErrTooManyFiles, "hosting.zip_too_many_files"},
		{zipsec.ErrTooDeep, "hosting.zip_too_deep"},
		{zipsec.ErrOverQuota, "hosting.extract_over_quota"},
	}

	out := make([]string, 0, 2)
	for _, it := range maps {
		if errors.Is(err, it.e) {
			out = append(out, it.code)
		}
	}
	return out
}

func zipErrCode(err error) string {
	codes := zipErrCodes(err)
	if len(codes) == 0 {
		return ""
	}
	return codes[0]
}
