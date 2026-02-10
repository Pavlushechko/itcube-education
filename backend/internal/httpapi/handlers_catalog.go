// internal/httpapi/handlers_catalog.go

package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/auth"
	"github.com/Pavlushechko/itcube-education/internal/domain"
	"github.com/Pavlushechko/itcube-education/internal/repo"
)

type CatalogHandler struct {
	v       *validator.Validate
	catalog *repo.CatalogRepo
}

type ProgramAdminView struct {
	Program domain.Program  `json:"Program"`
	Cohorts []domain.Cohort `json:"Cohorts"`
	Groups  []domain.Group  `json:"Groups"`
}

func NewCatalogHandler(catalog *repo.CatalogRepo) *CatalogHandler {
	return &CatalogHandler{v: validator.New(), catalog: catalog}
}

// Public: list published programs
func (h *CatalogHandler) ListPrograms(w http.ResponseWriter, r *http.Request) {
	ps, err := h.catalog.ListPublishedPrograms(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, ps)
}

// Public: program page (published) + open groups
func (h *CatalogHandler) GetProgram(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	pid, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	pg, err := h.catalog.GetPublishedProgramWithGroups(r.Context(), pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, pg)
}

// Admin: create draft program
type createProgramReq struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}

func (h *CatalogHandler) CreateProgram(w http.ResponseWriter, r *http.Request) {
	if auth.Role(r.Context()) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var req createProgramReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.v.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := h.catalog.CreateProgramDraft(r.Context(), req.Title, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id.String()})
}

func (h *CatalogHandler) PublishProgram(w http.ResponseWriter, r *http.Request) {
	if auth.Role(r.Context()) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.catalog.PublishProgram(r.Context(), pid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type createCohortReq struct {
	ProgramID string `json:"program_id" validate:"required,uuid"`
	Year      int    `json:"year" validate:"required"`
}

func (h *CatalogHandler) CreateCohort(w http.ResponseWriter, r *http.Request) {
	if auth.Role(r.Context()) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req createCohortReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.v.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pid, _ := uuid.Parse(req.ProgramID)

	if c, ok, err := h.catalog.GetCohortByProgramYear(r.Context(), pid, req.Year); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if ok {
		writeJSON(w, http.StatusOK, map[string]any{"id": c.ID.String()})
		return
	}

	id, err := h.catalog.CreateCohort(r.Context(), pid, req.Year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id.String()})
}

type createGroupReq struct {
	ProgramID         string `json:"program_id" validate:"required,uuid"`
	CohortID          string `json:"cohort_id" validate:"required,uuid"`
	Title             string `json:"title" validate:"required"`
	Capacity          int    `json:"capacity" validate:"required"`
	RequiresInterview bool   `json:"requires_interview"`
	IsOpen            bool   `json:"is_open"`
}

func (h *CatalogHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	if auth.Role(r.Context()) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	var req createGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.v.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pid, _ := uuid.Parse(req.ProgramID)
	cid, _ := uuid.Parse(req.CohortID)
	id, err := h.catalog.CreateGroup(r.Context(), pid, cid, req.Title, req.Capacity, req.RequiresInterview, req.IsOpen)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id.String()})
}

func (h *CatalogHandler) AssignTeacher(w http.ResponseWriter, r *http.Request) {
	if auth.Role(r.Context()) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	gid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	teacherIDStr := r.URL.Query().Get("teacher_user_id")
	if teacherIDStr == "" {
		http.Error(w, "teacher_user_id is required", http.StatusBadRequest)
		return
	}
	tid, err := uuid.Parse(teacherIDStr)
	if err != nil {
		http.Error(w, "invalid teacher_user_id", http.StatusBadRequest)
		return
	}
	if err := h.catalog.AssignTeacherToGroup(r.Context(), gid, tid); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CatalogHandler) CloseGroup(w http.ResponseWriter, r *http.Request) {
	if auth.Role(r.Context()) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	gid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	if err := h.catalog.SetGroupOpen(r.Context(), gid, false); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CatalogHandler) ListProgramsAdmin(w http.ResponseWriter, r *http.Request) {
	role := auth.Role(r.Context())
	if role != "admin" && role != "moderator" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	ps, err := h.catalog.ListAllPrograms(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, ps)
}

func (h *CatalogHandler) GetProgramAdmin(w http.ResponseWriter, r *http.Request) {
	role := auth.Role(r.Context())
	if role != "admin" && role != "moderator" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	pg, err := h.catalog.GetProgramWithGroupsAdmin(r.Context(), pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	cohorts, err := h.catalog.ListCohortsByProgram(r.Context(), pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, ProgramAdminView{
		Program: pg.Program,
		Cohorts: cohorts,
		Groups:  pg.Groups,
	})
}

func (h *CatalogHandler) GetGroupTeachers(w http.ResponseWriter, r *http.Request) {
	role := auth.Role(r.Context())
	if role != "admin" && role != "moderator" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	gid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	teachers, err := h.catalog.ListGroupTeachers(r.Context(), gid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"group_id": gid.String(),
		"teachers": teachers,
	})
}

func (h *CatalogHandler) RemoveTeacher(w http.ResponseWriter, r *http.Request) {
	if auth.Role(r.Context()) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	gid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	teacherIDStr := r.URL.Query().Get("teacher_user_id")
	if teacherIDStr == "" {
		http.Error(w, "teacher_user_id is required", http.StatusBadRequest)
		return
	}
	tid, err := uuid.Parse(teacherIDStr)
	if err != nil {
		http.Error(w, "invalid teacher_user_id", http.StatusBadRequest)
		return
	}

	if err := h.catalog.RemoveTeacherFromGroup(r.Context(), gid, tid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type updateGroupReq struct {
	Title             *string `json:"title"`
	Capacity          *int    `json:"capacity"`
	IsOpen            *bool   `json:"is_open"`
	RequiresInterview *bool   `json:"requires_interview"`
}

func (h *CatalogHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	if auth.Role(r.Context()) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	gid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	var req updateGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := h.catalog.UpdateGroup(r.Context(), gid, req.Title, req.Capacity, req.IsOpen, req.RequiresInterview); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type updateProgramReq struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

func (h *CatalogHandler) UpdateProgram(w http.ResponseWriter, r *http.Request) {
	if auth.Role(r.Context()) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req updateProgramReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if req.Title == nil && req.Description == nil {
		http.Error(w, "nothing to update", http.StatusBadRequest)
		return
	}

	if err := h.catalog.UpdateProgram(r.Context(), pid, req.Title, req.Description); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
