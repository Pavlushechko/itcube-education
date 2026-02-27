// internal/repo/assignment_repo.go

package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Pavlushechko/itcube-education/internal/domain"
)

type AssignmentRepo struct{ db *pgxpool.Pool }

func NewAssignmentRepo(db *pgxpool.Pool) *AssignmentRepo { return &AssignmentRepo{db: db} }

type AssignmentForLearnerView struct {
	// поля как у domain.Assignment (чтобы фронту было привычно)
	ID          uuid.UUID  `json:"ID"`
	GroupID     uuid.UUID  `json:"GroupID"`
	Title       string     `json:"Title"`
	Description string     `json:"Description"`
	DueAt       *time.Time `json:"DueAt"`
	CreatedBy   uuid.UUID  `json:"CreatedBy"`
	CreatedAt   time.Time  `json:"CreatedAt"`
	UpdatedAt   time.Time  `json:"UpdatedAt"`

	// новое
	MyStatus string `json:"MyStatus"` // not_done | in_review | reviewed
}

func (r *AssignmentRepo) Create(ctx context.Context, a domain.Assignment) error {
	_, err := r.db.Exec(ctx, `
		insert into assignments(id, group_id, title, description, due_at, created_by_user_id)
		values ($1,$2,$3,$4,$5,$6)
	`, a.ID, a.GroupID, a.Title, a.Description, a.DueAt, a.CreatedBy)
	return err
}

func (r *AssignmentRepo) ListByGroup(ctx context.Context, groupID uuid.UUID) ([]domain.Assignment, error) {
	rows, err := r.db.Query(ctx, `
		select id, group_id, title, description, due_at, created_by_user_id, created_at, updated_at
		from assignments
		where group_id=$1
		order by created_at desc
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Assignment
	for rows.Next() {
		var a domain.Assignment
		var due *time.Time
		if err := rows.Scan(&a.ID, &a.GroupID, &a.Title, &a.Description, &due, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		a.DueAt = due
		res = append(res, a)
	}
	return res, rows.Err()
}

func (r *AssignmentRepo) Get(ctx context.Context, id uuid.UUID) (domain.Assignment, error) {
	row := r.db.QueryRow(ctx, `
		select id, group_id, title, description, due_at, created_by_user_id, created_at, updated_at
		from assignments where id=$1
	`, id)
	var a domain.Assignment
	var due *time.Time
	if err := row.Scan(&a.ID, &a.GroupID, &a.Title, &a.Description, &due, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
		return domain.Assignment{}, err
	}
	a.DueAt = due
	return a, nil
}

func (r *AssignmentRepo) ListByGroupForLearnerView(ctx context.Context, groupID, userID uuid.UUID) ([]AssignmentForLearnerView, error) {
	rows, err := r.db.Query(ctx, `
		select 
			a.id, a.group_id, a.title, a.description, a.due_at, a.created_by_user_id, a.created_at, a.updated_at,
			case
				when s.id is null then 'not_done'
				when rv.id is null then 'in_review'
				else 'reviewed'
			end as my_status
		from assignments a
		left join submissions s
			on s.assignment_id = a.id and s.student_user_id = $2
		left join lateral (
			select id
			from submission_reviews
			where submission_id = s.id
			order by created_at desc
			limit 1
		) rv on true
		where a.group_id = $1
		order by a.created_at desc
	`, groupID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]AssignmentForLearnerView, 0)
	for rows.Next() {
		var a AssignmentForLearnerView
		if err := rows.Scan(
			&a.ID, &a.GroupID, &a.Title, &a.Description, &a.DueAt, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt,
			&a.MyStatus,
		); err != nil {
			return nil, err
		}
		res = append(res, a)
	}
	return res, rows.Err()
}
