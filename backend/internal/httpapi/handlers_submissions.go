// internal/httpapi/handlers_submissions.go

package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/service"
)

type SubmissionHandler struct {
	v   *validator.Validate
	svc *service.SubmissionService
}

func NewSubmissionHandler(svc *service.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{v: validator.New(), svc: svc}
}

type submitReq struct {
	ContentType string `json:"content_type" validate:"required"`
	Content     string `json:"content" validate:"required"`
}

func (h *SubmissionHandler) Submit(w http.ResponseWriter, r *http.Request) {
	aid, err := uuid.Parse(chi.URLParam(r, "assignmentID"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}
	var req submitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.v.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.svc.Submit(r.Context(), aid, req.ContentType, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id.String()})
}

func (h *SubmissionHandler) MySubmission(w http.ResponseWriter, r *http.Request) {
	aid, err := uuid.Parse(chi.URLParam(r, "assignmentID"))
	if err != nil {
		http.Error(w, "invalid assignment id", http.StatusBadRequest)
		return
	}
	sub, rv, err := h.svc.MySubmission(r.Context(), aid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"submission": sub,
		"review":     rv, // null if none
	})
}

func (h *SubmissionHandler) ListForTeacher(w http.ResponseWriter, r *http.Request) {
	gid, err := uuid.Parse(chi.URLParam(r, "groupID"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}
	var status *string
	if v := r.URL.Query().Get("status"); v != "" {
		status = &v
	}
	subs, err := h.svc.ListForTeacher(r.Context(), gid, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	writeJSON(w, http.StatusOK, subs)
}

type reviewReq struct {
	Grade   *int   `json:"grade"`
	Comment string `json:"comment"`
}

func (h *SubmissionHandler) Review(w http.ResponseWriter, r *http.Request) {
	sid, err := uuid.Parse(chi.URLParam(r, "submissionID"))
	if err != nil {
		http.Error(w, "invalid submission id", http.StatusBadRequest)
		return
	}
	var req reviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.svc.Review(r.Context(), sid, req.Grade, req.Comment); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
