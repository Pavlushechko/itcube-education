// internal/httpapi/handlers_teacher.go

package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/auth"
	"github.com/Pavlushechko/itcube-education/internal/domain"
	"github.com/Pavlushechko/itcube-education/internal/repo"
	"github.com/Pavlushechko/itcube-education/internal/service"
)

type TeacherHandler struct {
	v          *validator.Validate
	catalog    *repo.CatalogRepo
	appRepo    *repo.ApplicationRepo
	interviews *service.InterviewService
}

func NewTeacherHandler(catalog *repo.CatalogRepo, appRepo *repo.ApplicationRepo, invSvc *service.InterviewService) *TeacherHandler {
	return &TeacherHandler{
		v:          validator.New(),
		catalog:    catalog,
		appRepo:    appRepo,
		interviews: invSvc,
	}
}

func (h *TeacherHandler) MyGroups(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	gs, err := h.catalog.ListTeacherGroups(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, gs)
}

// list applications for group (e.g. in_review)
func (h *TeacherHandler) GroupApplications(w http.ResponseWriter, r *http.Request) {
	teacherID, ok := auth.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	gid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	assigned, err := h.catalog.IsTeacherInGroup(r.Context(), gid, teacherID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !assigned && auth.Role(r.Context()) != "admin" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// ✅ дефолт: учитель видит только заявки "в работе"
	statusVal := r.URL.Query().Get("status")
	if statusVal == "" {
		statusVal = "in_review"
	}
	status := &statusVal

	apps, err := h.appRepo.ListByFilter(r.Context(), &gid, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, apps)
}

type recordInterviewReq struct {
	Result  string `json:"result" validate:"required"`
	Comment string `json:"comment"`
}

func (h *TeacherHandler) RecordInterview(w http.ResponseWriter, r *http.Request) {
	// teacher OR moderator/admin can record, this is checked inside service as well
	var req recordInterviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.v.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "appID"))
	if err != nil {
		http.Error(w, "invalid app id", http.StatusBadRequest)
		return
	}

	res := domain.InterviewResult(req.Result)
	if err := h.interviews.Record(r.Context(), appID, res, req.Comment); err != nil {
		msg := err.Error()
		if strings.Contains(msg, "teacher is not assigned") || strings.Contains(msg, "forbidden") {
			http.Error(w, msg, http.StatusForbidden)
			return
		}
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TeacherHandler) GroupStudents(w http.ResponseWriter, r *http.Request) {
	actorID, ok := auth.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	role := auth.Role(r.Context())

	gid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid group id", http.StatusBadRequest)
		return
	}

	// Доступ:
	// - admin всегда
	// - иначе только если пользователь назначен преподавателем этой группы
	if role != "admin" {
		assigned, err := h.catalog.IsTeacherInGroup(r.Context(), gid, actorID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !assigned {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	users, err := h.appRepo.ListEnrolledUsersByGroup(r.Context(), gid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// отдаём просто список uuid (можно и объектами, но так проще)
	writeJSON(w, http.StatusOK, map[string]any{
		"group_id": gid.String(),
		"students": users,
	})
}

func (h *TeacherHandler) ProgramAccess(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	// важно: без 403. просто ok=false если не назначен.
	allowed, err := h.catalog.IsTeacherInProgram(r.Context(), uid, pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": allowed})
}
