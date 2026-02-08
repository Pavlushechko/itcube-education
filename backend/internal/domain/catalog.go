// internal/domain/catalog.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProgramStatus string

const (
	ProgramDraft     ProgramStatus = "draft"
	ProgramPublished ProgramStatus = "published"
)

type Program struct {
	ID          uuid.UUID
	Title       string
	Description string
	Status      ProgramStatus
	CreatedAt   time.Time
}

type Group struct {
	ID                uuid.UUID
	ProgramID         uuid.UUID
	CohortID          uuid.UUID
	Title             string
	Capacity          int
	IsOpen            bool
	RequiresInterview bool
	CreatedAt         time.Time
}

type Cohort struct {
	ID        uuid.UUID
	ProgramID uuid.UUID
	Year      int
	CreatedAt time.Time
}
