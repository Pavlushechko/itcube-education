// internal/domain/learning.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

type MaterialRead struct {
	UserID     uuid.UUID
	MaterialID uuid.UUID
	GroupID    uuid.UUID
	ReadAt     time.Time
}

type Assignment struct {
	ID          uuid.UUID
	GroupID     uuid.UUID
	Title       string
	Description string
	DueAt       *time.Time
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type SubmissionStatus string

const (
	SubmissionSubmitted SubmissionStatus = "submitted"
	SubmissionReviewed  SubmissionStatus = "reviewed"
)

type Submission struct {
	ID            uuid.UUID
	AssignmentID  uuid.UUID
	GroupID       uuid.UUID
	StudentUserID uuid.UUID
	ContentType   string // text|link (file later)
	Content       string
	Status        SubmissionStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SubmissionReview struct {
	ID           uuid.UUID
	SubmissionID uuid.UUID
	ReviewerID   uuid.UUID
	Grade        *int
	Comment      string
	CreatedAt    time.Time
}
