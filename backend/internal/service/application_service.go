// internal/service/application_service.go

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

var (
	ErrNoSeats           = errors.New("group is full")
	ErrProgramNotVisible = errors.New("program is not published")
	ErrGroupClosed       = errors.New("group is closed for applications")
	ErrInterviewRequired = errors.New("interview result is required before approval")
	ErrInterviewFailed   = errors.New("interview is not recommended")
)

type ApplicationService struct {
	appRepo     *repo.ApplicationRepo
	catalogRepo *repo.CatalogRepo
	interviews  *repo.InterviewRepo
	outbox      *outbox.Repo
}

func NewApplicationService(appRepo *repo.ApplicationRepo, catalogRepo *repo.CatalogRepo, interviewRepo *repo.InterviewRepo, outboxRepo *outbox.Repo) *ApplicationService {
	return &ApplicationService{appRepo: appRepo, catalogRepo: catalogRepo, interviews: interviewRepo, outbox: outboxRepo}
}

func (s *ApplicationService) Create(ctx context.Context, groupID uuid.UUID, comment string) (uuid.UUID, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}

	// ✅ запрет: учитель не может подавать заявку на СВОЙ курс
	pid, err := s.catalogRepo.GetGroupProgramID(ctx, groupID)
	if err != nil {
		return uuid.Nil, err
	}
	isTeacher, err := s.catalogRepo.IsTeacherInProgram(ctx, userID, pid)
	if err != nil {
		return uuid.Nil, err
	}
	if isTeacher {
		return uuid.Nil, errors.New("teacher cannot apply to own program")
	}

	rej, err := s.appRepo.HasRejected(ctx, userID, groupID)
	if err != nil {
		return uuid.Nil, err
	}
	if rej {
		return uuid.Nil, errors.New("cannot reapply after rejected") // или своё Err...
	}

	// ✅ правило повтора по ЭТОЙ ЖЕ группе:
	// cancelled -> можно снова (новая запись)
	// rejected  -> нельзя снова в эту же группу
	// submitted/in_review/approved -> нельзя (уже есть активная/успешная)
	if last, ok, err := s.appRepo.GetLatestByUserGroup(ctx, userID, groupID); err != nil {
		return uuid.Nil, err
	} else if ok {
		switch last.Status {
		case domain.AppCancelled:
			// ok, разрешаем создать новую
		case domain.AppRejected:
			return uuid.Nil, errors.New("cannot reapply to this group after rejection")
		case domain.AppSubmitted, domain.AppInReview, domain.AppApproved:
			return uuid.Nil, errors.New("application already exists")
		default:
			// на всякий случай: не плодим дублей
			return uuid.Nil, errors.New("application already exists")
		}
	}

	published, open, err := s.catalogRepo.IsGroupAvailableForApply(ctx, groupID)
	if err != nil {
		return uuid.Nil, err
	}
	if !published {
		return uuid.Nil, ErrProgramNotVisible
	}
	if !open {
		return uuid.Nil, ErrGroupClosed
	}

	app := domain.EnrollmentApplication{
		ID:      uuid.New(),
		UserID:  userID,
		GroupID: groupID,
		Status:  domain.AppSubmitted,
		Comment: comment,
	}
	if err := s.appRepo.Create(ctx, app); err != nil {
		return uuid.Nil, err
	}

	_ = s.outbox.Add(ctx, "enrollment_application", app.ID, "application.created", map[string]any{
		"application_id": app.ID.String(),
		"user_id":        userID.String(),
		"group_id":       groupID.String(),
	})

	return app.ID, nil
}

func (s *ApplicationService) ChangeStatus(ctx context.Context, appID uuid.UUID, to domain.ApplicationStatus, reason string) error {
	actorRole := auth.Role(ctx)

	actorID, ok := auth.UserID(ctx)
	if !ok {
		return errors.New("unauthorized")
	}

	app, err := s.appRepo.Get(ctx, appID)
	if err != nil {
		return err
	}

	// RBAC на уровне сервиса:
	if actorRole == "user" {
		if app.UserID != actorID {
			return errors.New("forbidden")
		}
	}

	if err := domain.CanTransition(app.Status, to, actorRole); err != nil {
		return err
	}

	// Если одобряем — проверяем места
	if to == domain.AppApproved {
		cap, err := s.appRepo.GroupCapacity(ctx, app.GroupID)
		if err != nil {
			return err
		}

		req, err := s.catalogRepo.GroupRequiresInterview(ctx, app.GroupID)
		if err != nil {
			return err
		}
		if req {
			inv, ok, err := s.interviews.GetByApplication(ctx, appID)
			if err != nil {
				return err
			}
			if !ok {
				return ErrInterviewRequired
			}
			if inv.Result != domain.InterviewRecommended {
				return ErrInterviewFailed
			}
		}

		cnt, err := s.appRepo.CountEnrollmentsByGroup(ctx, app.GroupID)
		if err != nil {
			return err
		}
		if cnt >= cap {
			return ErrNoSeats
		}
	}

	from := app.Status

	// меняем статус
	if err := s.appRepo.UpdateStatus(ctx, appID, to); err != nil {
		return err
	}

	// аудит
	if err := s.appRepo.InsertAudit(ctx, appID, actorID, actorRole, from, to, reason); err != nil {
		return err
	}

	// side-effect: enrollment
	if to == domain.AppApproved {
		if err := s.appRepo.CreateEnrollment(ctx, app.UserID, app.GroupID); err != nil {
			return err
		}
	}

	_ = s.outbox.Add(ctx, "enrollment_application", appID, "application.status_changed", map[string]any{
		"application_id": appID.String(),
		"from":           string(from),
		"to":             string(to),
		"actor_role":     actorRole,
	})

	return nil
}

func (s *ApplicationService) Cancel(ctx context.Context, appID uuid.UUID) error {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return errors.New("unauthorized")
	}

	ok2, err := s.appRepo.CancelByUser(ctx, appID, userID)
	if err != nil {
		return err
	}
	if !ok2 {
		// либо не твоя заявка, либо уже не тот статус
		return errors.New("cannot cancel application")
	}

	_ = s.outbox.Add(ctx, "enrollment_application", appID, "application.cancelled", map[string]any{
		"application_id": appID.String(),
		"user_id":        userID.String(),
	})

	return nil
}
