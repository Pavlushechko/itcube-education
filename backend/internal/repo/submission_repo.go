// internal/repo/submission_repo.go

package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Pavlushechko/itcube-education/internal/domain"
)

type SubmissionRepo struct{ db *pgxpool.Pool }

func NewSubmissionRepo(db *pgxpool.Pool) *SubmissionRepo { return &SubmissionRepo{db: db} }

func (r *SubmissionRepo) Upsert(ctx context.Context, s domain.Submission) error {
	_, err := r.db.Exec(ctx, `
		insert into submissions(id, assignment_id, group_id, student_user_id, content_type, content, status)
		values ($1,$2,$3,$4,$5,$6,$7)
		on conflict (assignment_id, student_user_id) do update set
			content_type=excluded.content_type,
			content=excluded.content,
			status='submitted',
			updated_at=now()
	`, s.ID, s.AssignmentID, s.GroupID, s.StudentUserID, s.ContentType, s.Content, string(s.Status))
	return err
}

func (r *SubmissionRepo) GetByAssignmentAndStudent(ctx context.Context, assignmentID, studentID uuid.UUID) (domain.Submission, bool, error) {
	row := r.db.QueryRow(ctx, `
		select id, assignment_id, group_id, student_user_id, content_type, content, status, created_at, updated_at
		from submissions
		where assignment_id=$1 and student_user_id=$2
	`, assignmentID, studentID)

	var s domain.Submission
	var st string
	err := row.Scan(&s.ID, &s.AssignmentID, &s.GroupID, &s.StudentUserID, &s.ContentType, &s.Content, &st, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return domain.Submission{}, false, nil
	}
	s.Status = domain.SubmissionStatus(st)
	return s, true, nil
}

func (r *SubmissionRepo) ListByGroup(ctx context.Context, groupID uuid.UUID, status *string) ([]domain.Submission, error) {
	q := `
		select id, assignment_id, group_id, student_user_id, content_type, content, status, created_at, updated_at
		from submissions
		where group_id=$1
	`
	args := []any{groupID}
	if status != nil {
		q += ` and status=$2`
		args = append(args, *status)
	}
	q += ` order by created_at desc`

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Submission
	for rows.Next() {
		var s domain.Submission
		var st string
		if err := rows.Scan(&s.ID, &s.AssignmentID, &s.GroupID, &s.StudentUserID, &s.ContentType, &s.Content, &st, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		s.Status = domain.SubmissionStatus(st)
		res = append(res, s)
	}
	return res, rows.Err()
}

func (r *SubmissionRepo) AddReview(ctx context.Context, rv domain.SubmissionReview) error {
	_, err := r.db.Exec(ctx, `
		insert into submission_reviews(id, submission_id, reviewer_user_id, grade, comment)
		values ($1,$2,$3,$4,$5)
	`, rv.ID, rv.SubmissionID, rv.ReviewerID, rv.Grade, rv.Comment)
	return err
}

func (r *SubmissionRepo) LatestReview(ctx context.Context, submissionID uuid.UUID) (domain.SubmissionReview, bool, error) {
	row := r.db.QueryRow(ctx, `
		select id, submission_id, reviewer_user_id, grade, comment, created_at
		from submission_reviews
		where submission_id=$1
		order by created_at desc
		limit 1
	`, submissionID)

	var rv domain.SubmissionReview
	var grade *int
	err := row.Scan(&rv.ID, &rv.SubmissionID, &rv.ReviewerID, &grade, &rv.Comment, &rv.CreatedAt)
	if err != nil {
		return domain.SubmissionReview{}, false, nil
	}
	rv.Grade = grade
	return rv, true, nil
}

func (r *SubmissionRepo) SetStatus(ctx context.Context, submissionID uuid.UUID, st domain.SubmissionStatus) error {
	_, err := r.db.Exec(ctx, `update submissions set status=$2, updated_at=now() where id=$1`, submissionID, string(st))
	return err
}

var _ = time.Now

func (r *SubmissionRepo) Get(ctx context.Context, id uuid.UUID) (domain.Submission, error) {
	row := r.db.QueryRow(ctx, `
		select id, assignment_id, group_id, student_user_id, content_type, content, status, created_at, updated_at
		from submissions
		where id=$1
	`, id)

	var s domain.Submission
	var st string
	if err := row.Scan(&s.ID, &s.AssignmentID, &s.GroupID, &s.StudentUserID, &s.ContentType, &s.Content, &st, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return domain.Submission{}, err
	}
	s.Status = domain.SubmissionStatus(st)
	return s, nil
}
