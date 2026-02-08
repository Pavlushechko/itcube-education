// internal/httpapi/handlers_application.go

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
	"github.com/Pavlushechko/itcube-education/internal/service"
)

type ApplicationHandler struct {
	validate *validator.Validate
	svc      *service.ApplicationService
	appRepo  *repo.ApplicationRepo
	catalog  *repo.CatalogRepo
}

func NewApplicationHandler(svc *service.ApplicationService, appRepo *repo.ApplicationRepo, catalog *repo.CatalogRepo) *ApplicationHandler {
	return &ApplicationHandler{
		validate: validator.New(),
		svc:      svc,
		appRepo:  appRepo,
		catalog:  catalog,
	}
}

type createAppReq struct {
	GroupID string `json:"group_id" validate:"required,uuid"`
	Comment string `json:"comment"`
}

func (h *ApplicationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createAppReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	gid, _ := uuid.Parse(req.GroupID)

	id, err := h.svc.Create(r.Context(), gid, req.Comment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id.String()})
}

func (h *ApplicationHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	apps, err := h.appRepo.ListByUser(r.Context(), uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, apps)
}

func (h *ApplicationHandler) ListByFilter(w http.ResponseWriter, r *http.Request) {
	role := auth.Role(r.Context())
	if role != "admin" && role != "moderator" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var groupID *uuid.UUID
	if v := r.URL.Query().Get("group_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			http.Error(w, "invalid group_id", http.StatusBadRequest)
			return
		}
		groupID = &id
	}

	var status *string
	if v := r.URL.Query().Get("status"); v != "" {
		status = &v
	}

	apps, err := h.appRepo.ListByFilter(r.Context(), groupID, status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, apps)
}

type changeStatusReq struct {
	Status string `json:"status" validate:"required"`
	Reason string `json:"reason"`
}

func (h *ApplicationHandler) ChangeStatus(w http.ResponseWriter, r *http.Request) {
	role := auth.Role(r.Context())
	if role != "admin" && role != "moderator" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	idStr := chi.URLParam(r, "id")
	appID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req changeStatusReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	to := domain.ApplicationStatus(req.Status)
	if err := h.svc.ChangeStatus(r.Context(), appID, to, req.Reason); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *ApplicationHandler) List(w http.ResponseWriter, r *http.Request) {
	actorID, ok := auth.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	role := auth.Role(r.Context())
	isStaff := role == "admin" || role == "moderator"

	// filters
	var groupID *uuid.UUID
	if v := r.URL.Query().Get("group_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			http.Error(w, "invalid group_id", http.StatusBadRequest)
			return
		}
		groupID = &id
	}

	var programID *uuid.UUID
	if v := r.URL.Query().Get("program_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			http.Error(w, "invalid program_id", http.StatusBadRequest)
			return
		}
		programID = &id
	}

	var status *string
	if v := r.URL.Query().Get("status"); v != "" {
		status = &v
	}

	// Для не-staff обязательно нужен фильтр, иначе можно случайно "попросить всё"
	if !isStaff && groupID == nil && programID == nil {
		http.Error(w, "group_id or program_id is required", http.StatusBadRequest)
		return
	}

	// Case 1: group_id filter
	if groupID != nil {
		if !isStaff {
			assigned, err := h.catalog.IsTeacherInGroup(r.Context(), *groupID, actorID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !assigned {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
		}

		apps, err := h.appRepo.ListByFilter(r.Context(), groupID, status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, apps)
		return
	}

	// Case 2: program_id filter
	if programID != nil {
		var apps []domain.EnrollmentApplication
		var err error

		if isStaff {
			apps, err = h.appRepo.ListByProgram(r.Context(), *programID, status)
		} else {
			// преподаватель = назначение, а не роль
			apps, err = h.appRepo.ListForTeacherByProgram(r.Context(), actorID, *programID, status)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, apps)
		return
	}

	// staff без фильтров — можно, но лучше не надо без пагинации
	http.Error(w, "program_id or group_id is required", http.StatusBadRequest)
}
