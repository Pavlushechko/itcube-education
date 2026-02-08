// internal/service/progress_service.go

package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/auth"
	"github.com/Pavlushechko/itcube-education/internal/repo"
)

type ProgressService struct {
	progress *repo.ProgressRepo
	matRepo  *repo.MaterialRepo
	appRepo  *repo.ApplicationRepo
}

func NewProgressService(progress *repo.ProgressRepo, matRepo *repo.MaterialRepo, appRepo *repo.ApplicationRepo) *ProgressService {
	return &ProgressService{progress: progress, matRepo: matRepo, appRepo: appRepo}
}

func (s *ProgressService) MarkMaterialRead(ctx context.Context, materialID uuid.UUID) error {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return errors.New("unauthorized")
	}

	// Need group_id for access check: easiest is to load material by listing group materials.
	// MVP: query materials table directly via matRepo helper? We'll add a tiny helper to MaterialRepo:
	// Get(materialID) -> (groupID, ...)
	m, err := s.matRepo.Get(ctx, materialID)
	if err != nil {
		return err
	}

	has, err := s.appRepo.HasEnrollment(ctx, userID, m.GroupID)
	if err != nil {
		return err
	}
	if !has {
		return ErrNoAccessToGroup
	}

	return s.progress.MarkRead(ctx, userID, materialID, m.GroupID)
}
