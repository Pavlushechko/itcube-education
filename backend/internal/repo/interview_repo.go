// internal/repo/interview_repo.go

package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Pavlushechko/itcube-education/internal/domain"
)

type InterviewRepo struct{ db *pgxpool.Pool }

func NewInterviewRepo(db *pgxpool.Pool) *InterviewRepo { return &InterviewRepo{db: db} }

func (r *InterviewRepo) GetByApplication(ctx context.Context, appID uuid.UUID) (domain.Interview, bool, error) {
	row := r.db.QueryRow(ctx, `
		select id, application_id, group_id, candidate_user_id, interviewer_user_id, interviewer_role,
		       result, comment, created_at, updated_at
		from interviews
		where application_id=$1
	`, appID)

	var i domain.Interview
	var res string
	err := row.Scan(&i.ID, &i.ApplicationID, &i.GroupID, &i.CandidateUserID, &i.InterviewerUserID, &i.InterviewerRole,
		&res, &i.Comment, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		// no rows -> false
		return domain.Interview{}, false, nil
	}
	i.Result = domain.InterviewResult(res)
	return i, true, nil
}

func (r *InterviewRepo) Upsert(ctx context.Context, in domain.Interview) error {
	now := time.Now()
	_, err := r.db.Exec(ctx, `
		insert into interviews(id, application_id, group_id, candidate_user_id, interviewer_user_id, interviewer_role, result, comment, created_at, updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8, now(), now())
		on conflict (application_id) do update set
			interviewer_user_id=excluded.interviewer_user_id,
			interviewer_role=excluded.interviewer_role,
			result=excluded.result,
			comment=excluded.comment,
			updated_at=now()
	`, uuid.New(), in.ApplicationID, in.GroupID, in.CandidateUserID, in.InterviewerUserID, in.InterviewerRole, string(in.Result), in.Comment)
	_ = now
	return err
}
