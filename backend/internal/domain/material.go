// internal/domain/material.go

package domain

import (
	"time"

	"github.com/google/uuid"
)

type MaterialType string

const (
	MaterialFile  MaterialType = "file"
	MaterialLink  MaterialType = "link"
	MaterialText  MaterialType = "text"
	MaterialVideo MaterialType = "video"
)

type Material struct {
	ID        uuid.UUID
	GroupID   uuid.UUID
	Type      MaterialType
	Title     string
	Content   string
	CreatedBy uuid.UUID
	CreatedAt time.Time
}
