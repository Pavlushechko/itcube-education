// internal/domain/interview.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

type InterviewResult string

const (
	InterviewPending        InterviewResult = "pending"
	InterviewRecommended    InterviewResult = "recommended"
	InterviewNotRecommended InterviewResult = "not_recommended"
	InterviewNeedsMore      InterviewResult = "needs_more"
)

type Interview struct {
	ID                uuid.UUID
	ApplicationID     uuid.UUID
	GroupID           uuid.UUID
	CandidateUserID   uuid.UUID
	InterviewerUserID uuid.UUID
	InterviewerRole   string // teacher|moderator
	Result            InterviewResult
	Comment           string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
