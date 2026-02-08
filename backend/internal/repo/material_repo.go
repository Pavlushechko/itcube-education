// internal/repo/material_repo.go

package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Pavlushechko/itcube-education/internal/domain"
)

type MaterialRepo struct{ db *pgxpool.Pool }

func NewMaterialRepo(db *pgxpool.Pool) *MaterialRepo { return &MaterialRepo{db: db} }

func (r *MaterialRepo) Create(ctx context.Context, m domain.Material) error {
	_, err := r.db.Exec(ctx, `
		insert into materials(id, group_id, type, title, content, created_by_user_id)
		values ($1,$2,$3,$4,$5,$6)
	`, m.ID, m.GroupID, string(m.Type), m.Title, m.Content, m.CreatedBy)
	return err
}

func (r *MaterialRepo) ListByGroup(ctx context.Context, groupID uuid.UUID) ([]domain.Material, error) {
	rows, err := r.db.Query(ctx, `
		select id, group_id, type, title, content, created_by_user_id, created_at
		from materials
		where group_id=$1
		order by created_at desc
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]domain.Material, 0)
	for rows.Next() {
		var m domain.Material
		var t string
		if err := rows.Scan(&m.ID, &m.GroupID, &t, &m.Title, &m.Content, &m.CreatedBy, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.Type = domain.MaterialType(t)
		res = append(res, m)
	}
	return res, rows.Err()
}

func (r *MaterialRepo) Get(ctx context.Context, id uuid.UUID) (domain.Material, error) {
	row := r.db.QueryRow(ctx, `
		select id, group_id, type, title, content, created_by_user_id, created_at
		from materials
		where id=$1
	`, id)

	var m domain.Material
	var t string
	if err := row.Scan(&m.ID, &m.GroupID, &t, &m.Title, &m.Content, &m.CreatedBy, &m.CreatedAt); err != nil {
		return domain.Material{}, err
	}
	m.Type = domain.MaterialType(t)
	return m, nil
}
