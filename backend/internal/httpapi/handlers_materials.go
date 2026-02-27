// internal/httpapi/handlers_materials.go

package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/domain"
	"github.com/Pavlushechko/itcube-education/internal/service"
)

type MaterialHandler struct {
	v   *validator.Validate
	svc *service.MaterialService
}

func NewMaterialHandler(svc *service.MaterialService) *MaterialHandler {
	return &MaterialHandler{v: validator.New(), svc: svc}
}

// learner endpoint: only after enrollment
func (h *MaterialHandler) ListForLearner(w http.ResponseWriter, r *http.Request) {
	gid, err := uuid.Parse(chi.URLParam(r, "groupID"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	ms, err := h.svc.ListForLearner(r.Context(), gid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeJSON(w, http.StatusOK, ms)
}

func (h *MaterialHandler) ListForTeacher(w http.ResponseWriter, r *http.Request) {
	gid, err := uuid.Parse(chi.URLParam(r, "groupID"))
	if err != nil {
		http.Error(w, "invalid group id", 400)
		return
	}

	ms, err := h.svc.ListForTeacher(r.Context(), gid)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}

	writeJSON(w, http.StatusOK, ms)
}

func (h *MaterialHandler) GetForTeacher(w http.ResponseWriter, r *http.Request) {
	mid, err := uuid.Parse(chi.URLParam(r, "materialID"))
	if err != nil {
		http.Error(w, "invalid material id", 400)
		return
	}

	dto, err := h.svc.GetForTeacher(r.Context(), mid)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}

	writeJSON(w, 200, dto)
}

type createMaterialReq struct {
	Type        string   `json:"type" validate:"required"`
	Title       string   `json:"title" validate:"required"`
	Content     string   `json:"content"`
	Attachments []string `json:"attachments"` // file_id[]
}

// teacher/admin endpoint
func (h *MaterialHandler) CreateForGroup(w http.ResponseWriter, r *http.Request) {
	gid, err := uuid.Parse(chi.URLParam(r, "groupID"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	var req createMaterialReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.v.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.svc.CreateForGroup(
		r.Context(),
		gid,
		domain.MaterialType(req.Type),
		req.Title,
		req.Content,
		req.Attachments,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id.String()})
}

// GET /learn/groups/{groupID} — only if enrolled
func (h *MaterialHandler) GroupInfo(w http.ResponseWriter, r *http.Request) {
	gid, err := uuid.Parse(chi.URLParam(r, "groupID"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	info, err := h.svc.GroupInfo(r.Context(), gid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	writeJSON(w, http.StatusOK, info)
}

// Teacher material page: content + files
func (h *MaterialHandler) GetPageForTeacher(w http.ResponseWriter, r *http.Request) {
	mid, err := uuid.Parse(chi.URLParam(r, "materialID"))
	if err != nil {
		http.Error(w, "invalid material id", http.StatusBadRequest)
		return
	}

	page, err := h.svc.GetMaterialPageForTeacher(r.Context(), mid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	writeJSON(w, http.StatusOK, page)
}

// Learner material page: only if enrolled
func (h *MaterialHandler) GetPageForLearner(w http.ResponseWriter, r *http.Request) {
	mid, err := uuid.Parse(chi.URLParam(r, "materialID"))
	if err != nil {
		http.Error(w, "invalid material id", http.StatusBadRequest)
		return
	}

	page, err := h.svc.GetMaterialPageForLearner(r.Context(), mid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	writeJSON(w, http.StatusOK, page)
}

type updateMaterialReq struct {
	Type        *string `json:"type"`
	Title       *string `json:"title"`
	Content     *string `json:"content"`
	ExternalURL *string `json:"external_url"` // null/"" -> очистить
}

func (h *MaterialHandler) UpdateForTeacher(w http.ResponseWriter, r *http.Request) {
	mid, err := uuid.Parse(chi.URLParam(r, "materialID"))
	if err != nil {
		http.Error(w, "invalid material id", http.StatusBadRequest)
		return
	}

	var req updateMaterialReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	if err := h.svc.UpdateForTeacher(r.Context(), mid, req.Type, req.Title, req.Content, req.ExternalURL); err != nil {
		// unauthorized/forbidden/etc
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type materialFileReq struct {
	FileID string `json:"file_id"`
}

func (h *MaterialHandler) AddFileForTeacher(w http.ResponseWriter, r *http.Request) {
	mid, err := uuid.Parse(chi.URLParam(r, "materialID"))
	if err != nil {
		http.Error(w, "invalid material id", 400)
		return
	}

	var req materialFileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	fid, err := uuid.Parse(req.FileID)
	if err != nil {
		http.Error(w, "invalid file id", 400)
		return
	}

	if err := h.svc.AddFileForTeacher(r.Context(), mid, fid); err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *MaterialHandler) RemoveFileForTeacher(w http.ResponseWriter, r *http.Request) {
	mid, err := uuid.Parse(chi.URLParam(r, "materialID"))
	if err != nil {
		http.Error(w, "invalid material id", 400)
		return
	}

	var req materialFileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", 400)
		return
	}
	fid, err := uuid.Parse(req.FileID)
	if err != nil {
		http.Error(w, "invalid file id", 400)
		return
	}

	if err := h.svc.RemoveFileForTeacher(r.Context(), mid, fid); err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
