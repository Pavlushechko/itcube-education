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

type createMaterialReq struct {
	Type    string `json:"type" validate:"required"`
	Title   string `json:"title" validate:"required"`
	Content string `json:"content"`
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

	id, err := h.svc.CreateForGroup(r.Context(), gid, domain.MaterialType(req.Type), req.Title, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id.String()})
}
