// internal/domain/application.go

package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type ApplicationStatus string

const (
	AppSubmitted ApplicationStatus = "submitted" // Отправлена
	AppInReview  ApplicationStatus = "in_review" // На рассмотрении
	AppApproved  ApplicationStatus = "approved"  // Одобрена (финал)
	AppRejected  ApplicationStatus = "rejected"  // Отклонена (финал)
	AppCancelled ApplicationStatus = "cancelled" // Отменена (финал)
)

func (s ApplicationStatus) IsFinal() bool {
	return s == AppApproved || s == AppRejected || s == AppCancelled
}

type EnrollmentApplication struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	GroupID   uuid.UUID
	Status    ApplicationStatus
	Comment   string // комментарий пользователя (опционально)
	CreatedAt time.Time
	UpdatedAt time.Time
}

var (
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrFinalStatus       = errors.New("cannot change final status")
)

func CanTransition(from, to ApplicationStatus, actorRole string) error {
	if from.IsFinal() {
		return ErrFinalStatus
	}

	// MVP правила:
	// Пользователь: submitted -> cancelled (и только свою)
	// Модератор/Админ: submitted -> in_review -> approved/rejected
	switch actorRole {
	case "user":
		if from == AppSubmitted && to == AppCancelled {
			return nil
		}
		return ErrInvalidTransition
	case "moderator", "admin":
		if from == AppSubmitted && to == AppInReview {
			return nil
		}
		if from == AppInReview && (to == AppApproved || to == AppRejected) {
			return nil
		}
		return ErrInvalidTransition
	default:
		return errors.New("unknown role")
	}
}
