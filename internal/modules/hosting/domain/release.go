package domain

import "time"

type ReleaseStatus string

const (
	ReleaseStatusPending ReleaseStatus = "pending"
	ReleaseStatusSuccess ReleaseStatus = "success"
	ReleaseStatusFailed  ReleaseStatus = "failed"
)

type Release struct {
	ID         ReleaseID
	SiteID     SiteID
	Status     ReleaseStatus
	SizeBytes  int64
	ErrorCode  string
	Violations []string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
