// internal/httpapi/router.go

package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Pavlushechko/itcube-education/internal/auth"
)

type Deps struct {
	ApplicationHandler *ApplicationHandler
	CatalogHandler     *CatalogHandler
	ProgramHandler     *ProgramHandler
	TeacherHandler     *TeacherHandler
	MaterialHandler    *MaterialHandler
	ProgressHandler    *ProgressHandler
	AssignmentHandler  *AssignmentHandler
	SubmissionHandler  *SubmissionHandler
	FilesHandler       *FilesHandler
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(auth.Middleware)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	r.Route("/files", func(r chi.Router) {
		r.Post("/upload", d.FilesHandler.Upload)
		r.Get("/{id}/download-url", d.FilesHandler.DownloadURL)
		r.Get("/{id}/download", d.FilesHandler.Download)
	})
	// Public catalog
	r.Route("/catalog", func(r chi.Router) {
		r.Get("/programs", d.CatalogHandler.ListPrograms)
		r.Get("/programs/{id}", d.CatalogHandler.GetProgram)
	})
	r.Get("/applications", d.ApplicationHandler.List)
	// Private program view (staff/teacher)
	r.Get("/programs/{id}", d.ProgramHandler.GetProgramPrivate)

	// Admin catalog management
	r.Route("/admin", func(r chi.Router) {
		// existing:
		r.Get("/applications", d.ApplicationHandler.List)         //
		r.Get("/programs", d.CatalogHandler.ListProgramsAdmin)    //
		r.Get("/programs/{id}", d.CatalogHandler.GetProgramAdmin) //
		r.Get("/groups/{id}/teachers", d.CatalogHandler.GetGroupTeachers)
		r.Post("/applications/{id}/status", d.ApplicationHandler.ChangeStatus)
		r.Post("/programs", d.CatalogHandler.CreateProgram)
		r.Post("/programs/{id}/publish", d.CatalogHandler.PublishProgram)
		r.Post("/cohorts", d.CatalogHandler.CreateCohort)
		r.Post("/groups", d.CatalogHandler.CreateGroup)
		r.Post("/groups/{id}/teachers", d.CatalogHandler.AssignTeacher) // teacher_user_id in query
		r.Post("/groups/{groupID}/materials", d.MaterialHandler.CreateForGroup)
		r.Post("/groups/{id}/close", d.CatalogHandler.CloseGroup)
		r.Patch("/groups/{id}", d.CatalogHandler.UpdateGroup)
		r.Patch("/programs/{id}", d.CatalogHandler.UpdateProgram)
		r.Delete("/groups/{id}/teachers", d.CatalogHandler.RemoveTeacher)
	})

	// Teacher
	r.Route("/teacher", func(r chi.Router) {
		r.Get("/groups", d.TeacherHandler.MyGroups)
		r.Get("/groups/{id}/applications", d.TeacherHandler.GroupApplications) //
		r.Post("/applications/{appID}/interview", d.TeacherHandler.RecordInterview)
		// router.go inside /teacher
		r.Get("/groups/{groupID}/materials", d.MaterialHandler.ListForTeacher)
		r.Get("/materials/{materialID}", d.MaterialHandler.GetPageForTeacher)
		r.Post("/groups/{groupID}/materials", d.MaterialHandler.CreateForGroup)
		r.Post("/materials/{materialID}/update", d.MaterialHandler.UpdateForTeacher)
		r.Post("/materials/{materialID}/files/add", d.MaterialHandler.AddFileForTeacher)
		r.Post("/materials/{materialID}/files/remove", d.MaterialHandler.RemoveFileForTeacher)
		r.Get("/groups/{groupID}/assignments", d.AssignmentHandler.ListForTeacher)
		r.Post("/groups/{groupID}/assignments", d.AssignmentHandler.CreateForGroup)
		r.Get("/groups/{groupID}/submissions", d.SubmissionHandler.ListForTeacher)
		r.Post("/submissions/{submissionID}/review", d.SubmissionHandler.Review)
		r.Get("/groups/{id}/students", d.TeacherHandler.GroupStudents)
		r.Get("/programs/{id}/access", d.TeacherHandler.ProgramAccess)

	})

	// Learner area (after enrollment)
	r.Route("/learn", func(r chi.Router) {
		r.Get("/groups/{groupID}", d.MaterialHandler.GroupInfo)
		r.Get("/groups/{groupID}/materials", d.MaterialHandler.ListForLearner)
		r.Get("/materials/{materialID}", d.MaterialHandler.GetPageForLearner)
		// mark material as read
		r.Post("/materials/{materialID}/read", d.ProgressHandler.MarkRead)
		// assignments
		r.Get("/groups/{groupID}/assignments", d.AssignmentHandler.ListForLearner)
		r.Post("/assignments/{assignmentID}/submissions", d.SubmissionHandler.Submit)
		r.Get("/assignments/{assignmentID}/submissions/me", d.SubmissionHandler.MySubmission)
	})

	// r.Get("/programs", ...)

	// Applications
	r.Route("/enrollments", func(r chi.Router) {
		r.Post("/applications", d.ApplicationHandler.Create)
		r.Get("/me/applications", d.ApplicationHandler.ListMine)
		r.Post("/applications/{id}/cancel", d.ApplicationHandler.CancelMyApplication)
	})

	return r
}
