// internal/service/material_service.go

package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/auth"
	"github.com/Pavlushechko/itcube-education/internal/domain"
	"github.com/Pavlushechko/itcube-education/internal/repo"
)

var (
	ErrNoAccessToGroup = errors.New("no access to group materials")
)

type MaterialService struct {
	matRepo     *repo.MaterialRepo
	appRepo     *repo.ApplicationRepo
	catalogRepo *repo.CatalogRepo
}

func NewMaterialService(matRepo *repo.MaterialRepo, appRepo *repo.ApplicationRepo, catalogRepo *repo.CatalogRepo) *MaterialService {
	return &MaterialService{matRepo: matRepo, appRepo: appRepo, catalogRepo: catalogRepo}
}

// learner: only if enrolled
func (s *MaterialService) ListForLearner(ctx context.Context, groupID uuid.UUID) ([]domain.Material, error) {
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
	return s.matRepo.ListByGroup(ctx, groupID)
}

// teacher/admin: can create
func (s *MaterialService) CreateForGroup(ctx context.Context, groupID uuid.UUID, typ domain.MaterialType, title, content string) (uuid.UUID, error) {
	role := auth.Role(ctx)
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}

	// admin always
	if role != "admin" {
		assigned, err := s.catalogRepo.IsTeacherInGroup(ctx, groupID, actorID)
		if err != nil {
			return uuid.Nil, err
		}
		if !assigned {
			return uuid.Nil, errors.New("forbidden")
		}
	}

	m := domain.Material{
		ID:        uuid.New(),
		GroupID:   groupID,
		Type:      typ,
		Title:     title,
		Content:   content,
		CreatedBy: actorID,
	}
	if err := s.matRepo.Create(ctx, m); err != nil {
		return uuid.Nil, err
	}
	return m.ID, nil
}
