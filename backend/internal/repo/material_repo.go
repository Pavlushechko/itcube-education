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

func (r *MaterialRepo) AddFiles(ctx context.Context, materialID uuid.UUID, fileIDs []uuid.UUID) error {
	if len(fileIDs) == 0 {
		return nil
	}

	for _, fid := range fileIDs {
		_, err := r.db.Exec(ctx, `
			insert into material_files(material_id, file_id)
			values ($1,$2)
			on conflict do nothing
		`, materialID, fid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MaterialRepo) ListFileIDs(ctx context.Context, materialID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `
		select file_id
		from material_files
		where material_id=$1
		order by created_at asc
	`, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		res = append(res, id)
	}
	return res, rows.Err()
}

func (r *MaterialRepo) Update(ctx context.Context, id uuid.UUID, typ *domain.MaterialType, title *string, content *string, externalURL *string) error {
	_, err := r.db.Exec(ctx, `
		update materials
		set
			type = coalesce($2, type),
			title = coalesce($3, title),
			content = coalesce($4, content),
			external_url = coalesce($5, external_url)
		where id = $1
	`, id,
		func() any {
			if typ == nil {
				return nil
			}
			return string(*typ)
		}(),
		title,
		content,
		externalURL,
	)
	return err
}

func (r *MaterialRepo) AddFile(ctx context.Context, materialID, fileID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		insert into material_files(material_id, file_id)
		values ($1,$2)
		on conflict do nothing
	`, materialID, fileID)
	return err
}

func (r *MaterialRepo) RemoveFile(ctx context.Context, materialID, fileID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		delete from material_files
		where material_id=$1 and file_id=$2
	`, materialID, fileID)
	return err
}
