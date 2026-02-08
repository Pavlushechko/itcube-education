// internal/service/assignment_service.go

package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/auth"
	"github.com/Pavlushechko/itcube-education/internal/domain"
	"github.com/Pavlushechko/itcube-education/internal/repo"
)

type AssignmentService struct {
	catalog *repo.CatalogRepo
	appRepo *repo.ApplicationRepo
	asgRepo *repo.AssignmentRepo
}

func NewAssignmentService(catalog *repo.CatalogRepo, appRepo *repo.ApplicationRepo, asgRepo *repo.AssignmentRepo) *AssignmentService {
	return &AssignmentService{catalog: catalog, appRepo: appRepo, asgRepo: asgRepo}
}

// Create: admin OR assigned teacher (not a global role)
func (s *AssignmentService) Create(ctx context.Context, groupID uuid.UUID, title, desc string, dueAt *time.Time) (uuid.UUID, error) {
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}
	role := auth.Role(ctx)

	if role != "admin" {
		assigned, err := s.catalog.IsTeacherInGroup(ctx, groupID, actorID)
		if err != nil {
			return uuid.Nil, err
		}
		if !assigned {
			return uuid.Nil, errors.New("forbidden")
		}
	}

	a := domain.Assignment{
		ID:          uuid.New(),
		GroupID:     groupID,
		Title:       title,
		Description: desc,
		DueAt:       dueAt,
		CreatedBy:   actorID,
	}
	if err := s.asgRepo.Create(ctx, a); err != nil {
		return uuid.Nil, err
	}
	return a.ID, nil
}

// ListForLearner: only if enrolled
func (s *AssignmentService) ListForLearner(ctx context.Context, groupID uuid.UUID) ([]domain.Assignment, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	has, err := s.appRepo.HasEnrollment(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrNoAccessToGroup
	}
	return s.asgRepo.ListByGroup(ctx, groupID)
}
