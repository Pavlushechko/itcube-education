// internal/repo/progress_repo.go

package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProgressRepo struct{ db *pgxpool.Pool }

func NewProgressRepo(db *pgxpool.Pool) *ProgressRepo { return &ProgressRepo{db: db} }

// MarkRead is idempotent.
func (r *ProgressRepo) MarkRead(ctx context.Context, userID, materialID, groupID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		insert into material_reads(user_id, material_id, group_id)
		values ($1,$2,$3)
		on conflict (user_id, material_id) do update set read_at=now()
	`, userID, materialID, groupID)
	return err
}

func (r *ProgressRepo) IsRead(ctx context.Context, userID, materialID uuid.UUID) (bool, error) {
	row := r.db.QueryRow(ctx, `select exists(select 1 from material_reads where user_id=$1 and material_id=$2)`, userID, materialID)
	var ok bool
	return ok, row.Scan(&ok)
}
