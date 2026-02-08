// internal/service/submission_service.go

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

type SubmissionService struct {
	catalog *repo.CatalogRepo
	appRepo *repo.ApplicationRepo
	asgRepo *repo.AssignmentRepo
	subRepo *repo.SubmissionRepo
}

func NewSubmissionService(catalog *repo.CatalogRepo, appRepo *repo.ApplicationRepo, asgRepo *repo.AssignmentRepo, subRepo *repo.SubmissionRepo) *SubmissionService {
	return &SubmissionService{catalog: catalog, appRepo: appRepo, asgRepo: asgRepo, subRepo: subRepo}
}

// Student submits result (MVP: upsert single submission)
func (s *SubmissionService) Submit(ctx context.Context, assignmentID uuid.UUID, contentType, content string) (uuid.UUID, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}

	asg, err := s.asgRepo.Get(ctx, assignmentID)
	if err != nil {
		return uuid.Nil, err
	}

	// access: must be enrolled in group
	has, err := s.appRepo.HasEnrollment(ctx, userID, asg.GroupID)
	if err != nil {
		return uuid.Nil, err
	}
	if !has {
		return uuid.Nil, ErrNoAccessToGroup
	}

	// optional: due date check (MVP: allow after due, but you can forbid)
	_ = time.Now()

	id := uuid.New()
	sub := domain.Submission{
		ID:            id,
		AssignmentID:  assignmentID,
		GroupID:       asg.GroupID,
		StudentUserID: userID,
		ContentType:   contentType,
		Content:       content,
		Status:        domain.SubmissionSubmitted,
	}
	if err := s.subRepo.Upsert(ctx, sub); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

// Student views own submission + latest review
func (s *SubmissionService) MySubmission(ctx context.Context, assignmentID uuid.UUID) (domain.Submission, *domain.SubmissionReview, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return domain.Submission{}, nil, errors.New("unauthorized")
	}

	sub, ok2, err := s.subRepo.GetByAssignmentAndStudent(ctx, assignmentID, userID)
	if err != nil {
		return domain.Submission{}, nil, err
	}
	if !ok2 {
		return domain.Submission{}, nil, errors.New("not found")
	}

	rv, ok3, err := s.subRepo.LatestReview(ctx, sub.ID)
	if err != nil {
		return domain.Submission{}, nil, err
	}
	if !ok3 {
		return sub, nil, nil
	}
	return sub, &rv, nil
}

// Teacher/Admin lists submissions for group
func (s *SubmissionService) ListForTeacher(ctx context.Context, groupID uuid.UUID, status *string) ([]domain.Submission, error) {
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	role := auth.Role(ctx)

	if role != "admin" {
		assigned, err := s.catalog.IsTeacherInGroup(ctx, groupID, actorID)
		if err != nil {
			return nil, err
		}
		if !assigned {
			return nil, errors.New("forbidden")
		}
	}

	return s.subRepo.ListByGroup(ctx, groupID, status)
}

// Teacher/Admin reviews a submission (grade/comment)
func (s *SubmissionService) Review(ctx context.Context, submissionID uuid.UUID, grade *int, comment string) error {
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return errors.New("unauthorized")
	}
	role := auth.Role(ctx)

	// We need submission's group to check teacher assignment.
	// MVP shortcut: query submission row here via SQL in repo:
	sub, err := s.subRepoGet(ctx, submissionID)
	if err != nil {
		return err
	}

	if role != "admin" {
		assigned, err := s.catalog.IsTeacherInGroup(ctx, sub.GroupID, actorID)
		if err != nil {
			return err
		}
		if !assigned {
			return errors.New("forbidden")
		}
	}

	rv := domain.SubmissionReview{
		ID:           uuid.New(),
		SubmissionID: submissionID,
		ReviewerID:   actorID,
		Grade:        grade,
		Comment:      comment,
	}
	if err := s.subRepo.AddReview(ctx, rv); err != nil {
		return err
	}
	return s.subRepo.SetStatus(ctx, submissionID, domain.SubmissionReviewed)
}

// small helper (keep MVP simple)
func (s *SubmissionService) subRepoGet(ctx context.Context, submissionID uuid.UUID) (domain.Submission, error) {
	// add a Get(id) to SubmissionRepo; inline minimal version here would be messy.
	// We'll require repo method:
	// r.Get(ctx, id) (domain.Submission, error)
	return s.subRepo.Get(ctx, submissionID)
}
