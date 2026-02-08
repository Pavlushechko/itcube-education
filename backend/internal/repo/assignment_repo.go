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
