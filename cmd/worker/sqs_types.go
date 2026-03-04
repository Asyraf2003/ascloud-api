package main

type sqsBatchResponse struct {
	BatchItemFailures []sqsBatchItemFailure `json:"batchItemFailures"`
}

type sqsBatchItemFailure struct {
	ItemIdentifier string `json:"itemIdentifier"`
}

func isPermanent(code string) bool {
	switch code {
	case "hosting.bad_message",
		"hosting.site_mismatch",
		"hosting.upload_not_queued",
		"hosting.zip_too_large",
		"hosting.zip_slip",
		"hosting.zip_symlink",
		"hosting.file_disallowed",
		"hosting.zip_too_many_files",
		"hosting.zip_too_deep",
		"hosting.extract_over_quota":
		return true
	default:
		return false
	}
}
