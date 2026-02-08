// internal/httpapi/handlers_assignments.go

package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/service"
)

type AssignmentHandler struct {
	v   *validator.Validate
	svc *service.AssignmentService
}

func NewAssignmentHandler(svc *service.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{v: validator.New(), svc: svc}
}

type createAssignmentReq struct {
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description"`
	DueAt       *string `json:"due_at"` // ISO8601, optional
}

func (h *AssignmentHandler) CreateForGroup(w http.ResponseWriter, r *http.Request) {
	gid, err := uuid.Parse(chi.URLParam(r, "groupID"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	var req createAssignmentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.v.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var due *time.Time
	if req.DueAt != nil && *req.DueAt != "" {
		t, err := time.Parse(time.RFC3339, *req.DueAt)
		if err != nil {
			http.Error(w, "invalid due_at", http.StatusBadRequest)
			return
		}
		due = &t
	}

	id, err := h.svc.Create(r.Context(), gid, req.Title, req.Description, due)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id.String()})
}

func (h *AssignmentHandler) ListForLearner(w http.ResponseWriter, r *http.Request) {
	gid, err := uuid.Parse(chi.URLParam(r, "groupID"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	as, err := h.svc.ListForLearner(r.Context(), gid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeJSON(w, http.StatusOK, as)
}
