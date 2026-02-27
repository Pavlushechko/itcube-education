// internal/service/material_service.go

package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/auth"
	"github.com/Pavlushechko/itcube-education/internal/domain"
	"github.com/Pavlushechko/itcube-education/internal/repo"
)

var (
	ErrNoAccessToGroup = errors.New("no access to group materials")
)

type MaterialService struct {
	matRepo     *repo.MaterialRepo
	appRepo     *repo.ApplicationRepo
	catalogRepo *repo.CatalogRepo
	fileRepo    *repo.FileRepo
}

func NewMaterialService(
	matRepo *repo.MaterialRepo,
	appRepo *repo.ApplicationRepo,
	catalogRepo *repo.CatalogRepo,
	fileRepo *repo.FileRepo,
) *MaterialService {
	return &MaterialService{matRepo: matRepo, appRepo: appRepo, catalogRepo: catalogRepo, fileRepo: fileRepo}
}

// learner: only if enrolled
func (s *MaterialService) ListForLearner(ctx context.Context, groupID uuid.UUID) ([]domain.Material, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	has, err := s.appRepo.HasEnrollment(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrNoAccessToGroup
	}
	return s.matRepo.ListByGroup(ctx, groupID)
}

// teacher/admin: can create
func (s *MaterialService) CreateForGroup(
	ctx context.Context,
	groupID uuid.UUID,
	typ domain.MaterialType,
	title string,
	content string,
	attachments []string,
) (uuid.UUID, error) {

	role := auth.Role(ctx)
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return uuid.Nil, errors.New("unauthorized")
	}

	// admin always
	if role != "admin" {
		assigned, err := s.catalogRepo.IsTeacherInGroup(ctx, groupID, actorID)
		if err != nil {
			return uuid.Nil, err
		}
		if !assigned {
			return uuid.Nil, errors.New("forbidden")
		}
	}

	// parse file UUIDs
	fileIDs := make([]uuid.UUID, 0, len(attachments))
	for _, a := range attachments {
		if a == "" {
			continue
		}
		fid, err := uuid.Parse(a)
		if err != nil {
			return uuid.Nil, errors.New("invalid attachment id")
		}
		fileIDs = append(fileIDs, fid)
	}

	m := domain.Material{
		ID:        uuid.New(),
		GroupID:   groupID,
		Type:      typ,
		Title:     title,
		Content:   content,
		CreatedBy: actorID,
	}

	if err := s.matRepo.Create(ctx, m); err != nil {
		return uuid.Nil, err
	}

	if err := s.matRepo.AddFiles(ctx, m.ID, fileIDs); err != nil {
		return uuid.Nil, err
	}

	return m.ID, nil
}

type GroupInfo struct {
	ProgramID    uuid.UUID
	ProgramTitle string
	GroupID      uuid.UUID
	GroupTitle   string
}

func (s *MaterialService) GroupInfo(ctx context.Context, groupID uuid.UUID) (GroupInfo, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return GroupInfo{}, errors.New("unauthorized")
	}

	has, err := s.appRepo.HasEnrollment(ctx, userID, groupID)
	if err != nil {
		return GroupInfo{}, err
	}
	if !has {
		return GroupInfo{}, ErrNoAccessToGroup
	}

	pid, ptitle, gtitle, err := s.catalogRepo.GetGroupProgramInfo(ctx, groupID)
	if err != nil {
		return GroupInfo{}, err
	}

	return GroupInfo{
		ProgramID:    pid,
		ProgramTitle: ptitle,
		GroupID:      groupID,
		GroupTitle:   gtitle,
	}, nil
}

func (s *MaterialService) ListForTeacher(ctx context.Context, groupID uuid.UUID) ([]domain.Material, error) {
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	role := auth.Role(ctx)
	if role != "admin" {
		assigned, err := s.catalogRepo.IsTeacherInGroup(ctx, groupID, actorID)
		if err != nil {
			return nil, err
		}
		if !assigned {
			return nil, errors.New("forbidden")
		}
	}
	return s.matRepo.ListByGroup(ctx, groupID)
}

func (s *MaterialService) GetForTeacher(ctx context.Context, materialID uuid.UUID) (map[string]any, error) {
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}

	m, err := s.matRepo.Get(ctx, materialID)
	if err != nil {
		return nil, err
	}

	role := auth.Role(ctx)
	if role != "admin" {
		assigned, err := s.catalogRepo.IsTeacherInGroup(ctx, m.GroupID, actorID)
		if err != nil {
			return nil, err
		}
		if !assigned {
			return nil, errors.New("forbidden")
		}
	}

	fids, err := s.matRepo.ListFileIDs(ctx, m.ID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(fids))
	for _, id := range fids {
		ids = append(ids, id.String())
	}

	return map[string]any{
		"id":         m.ID.String(),
		"group_id":   m.GroupID.String(),
		"type":       string(m.Type),
		"title":      m.Title,
		"content":    m.Content,
		"file_ids":   ids,
		"created_at": m.CreatedAt,
	}, nil
}

type MaterialPage struct {
	ID          string    `json:"id"`
	GroupID     string    `json:"group_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	ExternalURL *string   `json:"external_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`

	Files []struct {
		ID           string `json:"id"`
		OriginalName string `json:"original_name"`
		MimeType     string `json:"mime_type"`
		SizeBytes    int64  `json:"size_bytes"`
	} `json:"files"`
}

func (s *MaterialService) GetMaterialPageForTeacher(ctx context.Context, materialID uuid.UUID) (MaterialPage, error) {
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return MaterialPage{}, errors.New("unauthorized")
	}

	m, err := s.matRepo.Get(ctx, materialID)
	if err != nil {
		return MaterialPage{}, err
	}

	role := auth.Role(ctx)
	if role != "admin" {
		assigned, err := s.catalogRepo.IsTeacherInGroup(ctx, m.GroupID, actorID)
		if err != nil {
			return MaterialPage{}, err
		}
		if !assigned {
			return MaterialPage{}, errors.New("forbidden")
		}
	}

	files, err := s.fileRepo.ListByMaterial(ctx, materialID)
	if err != nil {
		return MaterialPage{}, err
	}

	out := MaterialPage{
		ID:        m.ID.String(),
		GroupID:   m.GroupID.String(),
		Type:      string(m.Type),
		Title:     m.Title,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
	// если добавишь external_url в domain/material.go — заполни тут
	out.Files = make([]struct {
		ID           string `json:"id"`
		OriginalName string `json:"original_name"`
		MimeType     string `json:"mime_type"`
		SizeBytes    int64  `json:"size_bytes"`
	}, 0, len(files))

	for _, f := range files {
		out.Files = append(out.Files, struct {
			ID           string `json:"id"`
			OriginalName string `json:"original_name"`
			MimeType     string `json:"mime_type"`
			SizeBytes    int64  `json:"size_bytes"`
		}{
			ID: f.ID.String(), OriginalName: f.OriginalName, MimeType: f.MimeType, SizeBytes: f.SizeBytes,
		})
	}
	return out, nil
}

func (s *MaterialService) GetMaterialPageForLearner(ctx context.Context, materialID uuid.UUID) (MaterialPage, error) {
	userID, ok := auth.UserID(ctx)
	if !ok {
		return MaterialPage{}, errors.New("unauthorized")
	}

	m, err := s.matRepo.Get(ctx, materialID)
	if err != nil {
		return MaterialPage{}, err
	}

	has, err := s.appRepo.HasEnrollment(ctx, userID, m.GroupID)
	if err != nil {
		return MaterialPage{}, err
	}
	if !has {
		return MaterialPage{}, ErrNoAccessToGroup
	}

	files, err := s.fileRepo.ListByMaterial(ctx, materialID)
	if err != nil {
		return MaterialPage{}, err
	}

	out := MaterialPage{
		ID:        m.ID.String(),
		GroupID:   m.GroupID.String(),
		Type:      string(m.Type),
		Title:     m.Title,
		Content:   m.Content,
		CreatedAt: m.CreatedAt,
	}
	// если добавишь external_url в domain/material.go — заполни тут
	out.Files = make([]struct {
		ID           string `json:"id"`
		OriginalName string `json:"original_name"`
		MimeType     string `json:"mime_type"`
		SizeBytes    int64  `json:"size_bytes"`
	}, 0, len(files))

	for _, f := range files {
		out.Files = append(out.Files, struct {
			ID           string `json:"id"`
			OriginalName string `json:"original_name"`
			MimeType     string `json:"mime_type"`
			SizeBytes    int64  `json:"size_bytes"`
		}{
			ID: f.ID.String(), OriginalName: f.OriginalName, MimeType: f.MimeType, SizeBytes: f.SizeBytes,
		})
	}
	return out, nil
}

func (s *MaterialService) UpdateForTeacher(
	ctx context.Context,
	materialID uuid.UUID,
	typ *string,
	title *string,
	content *string,
	externalURL *string,
) error {
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return errors.New("unauthorized")
	}

	// material -> group
	m, err := s.matRepo.Get(ctx, materialID)
	if err != nil {
		return err
	}

	role := auth.Role(ctx)
	if role != "admin" {
		assigned, err := s.catalogRepo.IsTeacherInGroup(ctx, m.GroupID, actorID)
		if err != nil {
			return err
		}
		if !assigned {
			return errors.New("forbidden")
		}
	}

	// parse typ if provided
	var tParsed *domain.MaterialType
	if typ != nil {
		tt := domain.MaterialType(*typ)
		// optional: validate allowed values
		switch tt {
		case domain.MaterialFile, domain.MaterialLink, domain.MaterialText, domain.MaterialVideo:
			tParsed = &tt
		default:
			return errors.New("invalid type")
		}
	}

	// normalize external_url: allow explicit null/empty -> set null
	// If externalURL != nil and trimmed == "" -> set to nil to clear.
	if externalURL != nil {
		v := strings.TrimSpace(*externalURL)
		if v == "" {
			externalURL = nil
		} else {
			externalURL = &v
		}
	}

	return s.matRepo.Update(ctx, materialID, tParsed, title, content, externalURL)
}

func (s *MaterialService) AddFileForTeacher(ctx context.Context, materialID, fileID uuid.UUID) error {
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return errors.New("unauthorized")
	}

	m, err := s.matRepo.Get(ctx, materialID)
	if err != nil {
		return err
	}

	role := auth.Role(ctx)
	if role != "admin" {
		assigned, err := s.catalogRepo.IsTeacherInGroup(ctx, m.GroupID, actorID)
		if err != nil {
			return err
		}
		if !assigned {
			return errors.New("forbidden")
		}
	}

	// (опционально) можно проверить что file существует: s.fileRepo.Get(ctx,fileID)
	return s.matRepo.AddFile(ctx, materialID, fileID)
}

func (s *MaterialService) RemoveFileForTeacher(ctx context.Context, materialID, fileID uuid.UUID) error {
	actorID, ok := auth.UserID(ctx)
	if !ok {
		return errors.New("unauthorized")
	}

	m, err := s.matRepo.Get(ctx, materialID)
	if err != nil {
		return err
	}

	role := auth.Role(ctx)
	if role != "admin" {
		assigned, err := s.catalogRepo.IsTeacherInGroup(ctx, m.GroupID, actorID)
		if err != nil {
			return err
		}
		if !assigned {
			return errors.New("forbidden")
		}
	}

	return s.matRepo.RemoveFile(ctx, materialID, fileID)
}
