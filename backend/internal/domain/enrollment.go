// internal/domain/enrollment.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

type Enrollment struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	GroupID   uuid.UUID
	CreatedAt time.Time
}
