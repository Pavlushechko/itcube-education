// internal/repo/file_repo.go

package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileMeta struct {
	ID           uuid.UUID
	Bucket       string
	ObjectKey    string
	OriginalName string
	MimeType     string
	SizeBytes    int64
	CreatedBy    uuid.UUID
	CreatedAt    time.Time
}

type FileRepo struct{ db *pgxpool.Pool }

func NewFileRepo(db *pgxpool.Pool) *FileRepo { return &FileRepo{db: db} }

func (r *FileRepo) Create(ctx context.Context, f FileMeta) error {
	_, err := r.db.Exec(ctx, `
		insert into files(id, bucket, object_key, original_name, mime_type, size_bytes, created_by_user_id)
		values ($1,$2,$3,$4,$5,$6,$7)
	`, f.ID, f.Bucket, f.ObjectKey, f.OriginalName, f.MimeType, f.SizeBytes, f.CreatedBy)
	return err
}

func (r *FileRepo) Get(ctx context.Context, id uuid.UUID) (FileMeta, error) {
	row := r.db.QueryRow(ctx, `
		select id, bucket, object_key, original_name, mime_type, size_bytes, created_by_user_id, created_at
		from files where id=$1
	`, id)

	var f FileMeta
	if err := row.Scan(&f.ID, &f.Bucket, &f.ObjectKey, &f.OriginalName, &f.MimeType, &f.SizeBytes, &f.CreatedBy, &f.CreatedAt); err != nil {
		return FileMeta{}, err
	}
	return f, nil
}

// internal/repo/file_repo.go
type FileBrief struct {
	ID           uuid.UUID
	OriginalName string
	MimeType     string
	SizeBytes    int64
	CreatedAt    time.Time
}

func (r *FileRepo) ListByMaterial(ctx context.Context, materialID uuid.UUID) ([]FileBrief, error) {
	rows, err := r.db.Query(ctx, `
		select f.id, f.original_name, f.mime_type, f.size_bytes, f.created_at
		from material_files mf
		join files f on f.id = mf.file_id
		where mf.material_id=$1
		order by mf.created_at asc
	`, materialID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]FileBrief, 0)
	for rows.Next() {
		var b FileBrief
		if err := rows.Scan(&b.ID, &b.OriginalName, &b.MimeType, &b.SizeBytes, &b.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, rows.Err()
}
