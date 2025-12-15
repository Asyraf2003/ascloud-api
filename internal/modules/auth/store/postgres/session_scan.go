package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"example.com/your-api/internal/modules/auth/domain"
)

func scanSession(ctx context.Context, row *sql.Row) (domain.Session, error) {
	_ = ctx

	var (
		id, userID, projectIDStr string
		refreshHash              string
		deviceID                 string
		uaHash                   string
		ipPrefixStr              string
		createdAt                time.Time
		expiresAt                time.Time
		revokedAt                sql.NullTime
		metaRaw                  []byte
	)

	err := row.Scan(
		&id, &userID, &projectIDStr,
		&refreshHash,
		&deviceID,
		&uaHash,
		&ipPrefixStr,
		&createdAt,
		&expiresAt,
		&revokedAt,
		&metaRaw,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Session{}, domain.ErrSessionNotFound
		}
		return domain.Session{}, err
	}

	var projectID *string
	if projectIDStr != "" {
		projectID = &projectIDStr
	}

	var ipPrefix *string
	if ipPrefixStr != "" {
		ipPrefix = &ipPrefixStr
	}

	var revokedPtr *time.Time
	if revokedAt.Valid {
		revokedPtr = &revokedAt.Time
	}

	meta := map[string]any{}
	if len(metaRaw) > 0 {
		_ = json.Unmarshal(metaRaw, &meta)
	}

	return domain.Session{
		ID:               id,
		UserID:           userID,
		ProjectID:        projectID,
		RefreshTokenHash: refreshHash,
		DeviceID:         deviceID,
		UserAgentHash:    uaHash,
		IPPrefix:         ipPrefix,
		CreatedAt:        createdAt,
		ExpiresAt:        expiresAt,
		RevokedAt:        revokedPtr,
		Meta:             meta,
	}, nil
}
