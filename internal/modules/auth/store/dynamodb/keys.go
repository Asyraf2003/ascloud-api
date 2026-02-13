package dynamodb

import "strings"

func sidPK(id string) string                { return "sid#" + strings.TrimSpace(id) }
func rhPK(hash string) string               { return "rh#" + strings.TrimSpace(hash) }
func usrPK(uid string) string               { return "usr#" + strings.TrimSpace(uid) }
func idpPK(provider, subject string) string { return "idp#" + provider + "#sub#" + subject }
func accPK(id string) string                { return "acc#" + strings.TrimSpace(id) }
