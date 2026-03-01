package dynamodb

import "strings"

func uplPK(id string) string  { return "upl#" + strings.TrimSpace(id) }
func sitePK(id string) string { return "site#" + strings.TrimSpace(id) }
func relPK(id string) string  { return "rel#" + strings.TrimSpace(id) }
