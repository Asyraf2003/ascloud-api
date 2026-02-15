package dynamodb

import "strings"

func uplPK(id string) string { return "upl#" + strings.TrimSpace(id) }
