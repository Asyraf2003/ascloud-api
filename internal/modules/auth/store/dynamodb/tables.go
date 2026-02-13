package dynamodb

import (
	"os"
	"strings"
)

type TableNames struct {
	Accounts    string
	Identities  string
	Sessions    string
	AuditEvents string
}

func TableNamesFromEnv(prefix string) TableNames {
	g := func(k, def string) string {
		v := strings.TrimSpace(os.Getenv(prefix + k))
		if v == "" {
			return def
		}
		return v
	}
	return TableNames{
		Accounts:    g("DDB_ACCOUNTS_TABLE", "accounts"),
		Identities:  g("DDB_IDENTITIES_TABLE", "auth_identities"),
		Sessions:    g("DDB_SESSIONS_TABLE", "sessions"),
		AuditEvents: g("DDB_AUDIT_EVENTS_TABLE", "audit_events"),
	}
}
