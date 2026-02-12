//go:build legacy_postgres
// +build legacy_postgres

package postgres

import (
	"fmt"

	"github.com/google/uuid"
)

func ParseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid: %w", err)
	}
	return id, nil
}
