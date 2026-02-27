package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Pavlushechko/itcube-education/internal/auth"
	"github.com/Pavlushechko/itcube-education/internal/repo"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type FileService struct {
	minio  *minio.Client
	bucket string
	files  *repo.FileRepo
}

type UploadURL struct {
	FileID    uuid.UUID `json:"file_id"`
	PutURL    string    `json:"put_url"`
	ObjectKey string    `json:"object_key"`
	Bucket    string    `json:"bucket"`
}

func NewFileService(minioClient *minio.Client, bucket string, files *repo.FileRepo) *FileService {
	if minioClient == nil {
		panic("minioClient is nil")
	}
	if files == nil {
		panic("files repo is nil")
	}
	if bucket == "" {
		bucket = "education"
	}
	return &FileService{minio: minioClient, bucket: bucket, files: files}
}

func NewMinioClient(endpoint, accessKey, secretKey string, useSSL bool) (*minio.Client, error) {
	return minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
}

// Создаём запись + выдаём presigned PUT.
// sizeBytes — метаданные, MinIO сам размер на presign не валидирует.
func (s *FileService) CreateUploadURL(ctx context.Context, originalName, mime string, sizeBytes int64) (UploadURL, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return UploadURL{}, errors.New("unauthorized")
	}
	if originalName == "" {
		return UploadURL{}, errors.New("original_name is required")
	}
	if strings.TrimSpace(mime) == "" {
		mime = "application/octet-stream"
	}

	id := uuid.New()
	objectKey := fmt.Sprintf("%s/%s", userID.String(), id.String())

	// ensure bucket exists
	exists, err := s.minio.BucketExists(ctx, s.bucket)
	if err != nil {
		return UploadURL{}, err
	}
	if !exists {
		if err := s.minio.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return UploadURL{}, err
		}
	}

	meta := repo.FileMeta{
		ID:           id,
		Bucket:       s.bucket,
		ObjectKey:    objectKey,
		OriginalName: originalName,
		MimeType:     mime,
		SizeBytes:    sizeBytes,
		CreatedBy:    userID,
	}
	if err := s.files.Create(ctx, meta); err != nil {
		return UploadURL{}, err
	}

	u, err := s.minio.PresignedPutObject(ctx, s.bucket, objectKey, 15*time.Minute)
	if err != nil {
		return UploadURL{}, err
	}

	return UploadURL{
		FileID:    id,
		PutURL:    u.String(),
		ObjectKey: objectKey,
		Bucket:    s.bucket,
	}, nil
}

func (s *FileService) GetDownloadURL(ctx context.Context, fileID uuid.UUID) (string, error) {
	_, ok := auth.UserID(ctx)
	if !ok {
		return "", errors.New("unauthorized")
	}

	f, err := s.files.Get(ctx, fileID)
	if err != nil {
		return "", err
	}

	u, err := s.minio.PresignedGetObject(ctx, f.Bucket, f.ObjectKey, 15*time.Minute, url.Values{})
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Решение B: стримим файл через backend, чтобы браузер скачал без переходов на MinIO.
func (s *FileService) StreamDownload(ctx context.Context, w http.ResponseWriter, fileID uuid.UUID) error {
	_, ok := auth.UserID(ctx)
	if !ok {
		return errors.New("unauthorized")
	}

	meta, err := s.files.Get(ctx, fileID)
	if err != nil {
		return err
	}

	obj, err := s.minio.GetObject(ctx, meta.Bucket, meta.ObjectKey, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer obj.Close()

	st, statErr := obj.Stat()

	ct := meta.MimeType
	if strings.TrimSpace(ct) == "" {
		ct = "application/octet-stream"
	}

	fname := meta.OriginalName
	if strings.TrimSpace(fname) == "" {
		fname = "file"
	}
	encoded := url.PathEscape(fname)

	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", encoded))
	if statErr == nil && st.Size > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", st.Size))
	}

	_, err = io.Copy(w, obj)
	return err
}

func (s *FileService) UploadFile(ctx context.Context, originalName, mime string, sizeBytes int64, body io.Reader) (uuid.UUID, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}
	if originalName == "" {
		return uuid.Nil, errors.New("original_name is required")
	}
	if strings.TrimSpace(mime) == "" {
		mime = "application/octet-stream"
	}

	id := uuid.New()
	objectKey := fmt.Sprintf("%s/%s", userID.String(), id.String())

	// ensure bucket exists
	exists, err := s.minio.BucketExists(ctx, s.bucket)
	if err != nil {
		return uuid.Nil, err
	}
	if !exists {
		if err := s.minio.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return uuid.Nil, err
		}
	}

	// 1) put object
	_, err = s.minio.PutObject(ctx, s.bucket, objectKey, body, sizeBytes, minio.PutObjectOptions{
		ContentType: mime,
	})
	if err != nil {
		return uuid.Nil, err
	}

	// 2) save metadata
	meta := repo.FileMeta{
		ID: id, Bucket: s.bucket, ObjectKey: objectKey,
		OriginalName: originalName, MimeType: mime, SizeBytes: sizeBytes,
		CreatedBy: userID,
	}
	if err := s.files.Create(ctx, meta); err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
