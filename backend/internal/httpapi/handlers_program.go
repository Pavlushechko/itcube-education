// internal/httpapi/handlers_program.go

package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/auth"
	"github.com/Pavlushechko/itcube-education/internal/domain"
	"github.com/Pavlushechko/itcube-education/internal/repo"
)

type ProgramHandler struct {
	catalog *repo.CatalogRepo
}

func NewProgramHandler(catalog *repo.CatalogRepo) *ProgramHandler {
	return &ProgramHandler{catalog: catalog}
}

type ProgramPrivateView struct {
	Program domain.Program  `json:"program"`
	Cohorts []domain.Cohort `json:"cohorts"`
	Groups  []domain.Group  `json:"groups"`
}

// GET /programs/{id}
// 200: staff всегда; teacher(user) только если назначен хотя бы в одну группу этого курса
// 403: обычный пользователь (фронт покажет public /catalog/programs/{id})
func (h *ProgramHandler) GetProgramPrivate(w http.ResponseWriter, r *http.Request) {
	actorID, ok := auth.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	pid, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	role := auth.Role(r.Context())
	isStaff := role == "admin" || role == "moderator"

	p, err := h.catalog.GetProgram(r.Context(), pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	cohorts, err := h.catalog.ListCohortsByProgram(r.Context(), pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var groups []domain.Group
	if isStaff {
		groups, err = h.catalog.ListGroupsByProgram(r.Context(), pid)
	} else {
		groups, err = h.catalog.ListTeacherGroupsByProgram(r.Context(), actorID, pid)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isStaff && len(groups) == 0 {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	writeJSON(w, http.StatusOK, ProgramPrivateView{
		Program: p,
		Cohorts: cohorts,
		Groups:  groups,
	})
}
