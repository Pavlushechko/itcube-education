// internal/service/interview_service.go

package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/auth"
	"github.com/Pavlushechko/itcube-education/internal/domain"
	"github.com/Pavlushechko/itcube-education/internal/outbox"
	"github.com/Pavlushechko/itcube-education/internal/repo"
)

var ErrNotAssignedTeacher = errors.New("teacher is not assigned to this group")

type InterviewService struct {
	appRepo     *repo.ApplicationRepo
	catalogRepo *repo.CatalogRepo
	interviews  *repo.InterviewRepo
	outbox      *outbox.Repo
}

func NewInterviewService(appRepo *repo.ApplicationRepo, catalogRepo *repo.CatalogRepo, interviewRepo *repo.InterviewRepo, outboxRepo *outbox.Repo) *InterviewService {
	return &InterviewService{appRepo: appRepo, catalogRepo: catalogRepo, interviews: interviewRepo, outbox: outboxRepo}
}

func (s *InterviewService) Record(ctx context.Context, appID uuid.UUID, result domain.InterviewResult, comment string) error {
	role := auth.Role(ctx)
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return errors.New("unauthorized")
	}

	app, err := s.appRepo.Get(ctx, appID)
	if err != nil {
		return err
	}

	// admin/moderator always; otherwise must be assigned teacher
	if role != "admin" && role != "moderator" {
		assigned, err := s.catalogRepo.IsTeacherInGroup(ctx, app.GroupID, actorID)
		if err != nil {
			return err
		}
		if !assigned {
			return ErrNotAssignedTeacher
		}
	}

	inv := domain.Interview{
		ApplicationID:     appID,
		GroupID:           app.GroupID,
		CandidateUserID:   app.UserID,
		InterviewerUserID: actorID,
		InterviewerRole:   role, // будет "user" у преподавателя — нормально для MVP
		Result:            result,
		Comment:           comment,
	}

	if err := s.interviews.Upsert(ctx, inv); err != nil {
		return err
	}

	_ = s.outbox.Add(ctx, "interview", appID, "interview.recorded", map[string]any{
		"application_id": appID.String(),
		"group_id":       app.GroupID.String(),
		"candidate_id":   app.UserID.String(),
		"result":         string(result),
		"actor_role":     role,
	})

	return nil
}
