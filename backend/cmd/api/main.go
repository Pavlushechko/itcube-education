// cmd/api/main.go

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/Pavlushechko/itcube-education/internal/config"
	"github.com/Pavlushechko/itcube-education/internal/db"
	"github.com/Pavlushechko/itcube-education/internal/httpapi"
	"github.com/Pavlushechko/itcube-education/internal/outbox"
	"github.com/Pavlushechko/itcube-education/internal/repo"
	"github.com/Pavlushechko/itcube-education/internal/service"
)

// go run .\cmd\api
func main() {
	_ = godotenv.Load()

	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		slog.Error("DATABASE_URL is empty")
		os.Exit(1)
	}

	ctx := context.Background()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	catalogRepo := repo.NewCatalogRepo(pool)
	interviewRepo := repo.NewInterviewRepo(pool)

	appRepo := repo.NewApplicationRepo(pool)
	outboxRepo := outbox.New(pool)

	appSvc := service.NewApplicationService(appRepo, catalogRepo, interviewRepo, outboxRepo)
	invSvc := service.NewInterviewService(appRepo, catalogRepo, interviewRepo, outboxRepo)

	appHandler := httpapi.NewApplicationHandler(appSvc, appRepo, catalogRepo)
	catalogHandler := httpapi.NewCatalogHandler(catalogRepo)
	teacherHandler := httpapi.NewTeacherHandler(catalogRepo, appRepo, invSvc)

	matRepo := repo.NewMaterialRepo(pool)
	matSvc := service.NewMaterialService(matRepo, appRepo, catalogRepo)
	matHandler := httpapi.NewMaterialHandler(matSvc)

	progressRepo := repo.NewProgressRepo(pool)
	asgRepo := repo.NewAssignmentRepo(pool)
	subRepo := repo.NewSubmissionRepo(pool)

	progressSvc := service.NewProgressService(progressRepo, matRepo, appRepo)
	asgSvc := service.NewAssignmentService(catalogRepo, appRepo, asgRepo)
	subSvc := service.NewSubmissionService(catalogRepo, appRepo, asgRepo, subRepo)

	progressHandler := httpapi.NewProgressHandler(progressSvc)
	asgHandler := httpapi.NewAssignmentHandler(asgSvc)
	subHandler := httpapi.NewSubmissionHandler(subSvc)

	router := httpapi.NewRouter(httpapi.Deps{
		ApplicationHandler: appHandler,
		CatalogHandler:     catalogHandler,
		TeacherHandler:     teacherHandler,
		MaterialHandler:    matHandler,
		ProgressHandler:    progressHandler,
		AssignmentHandler:  asgHandler,
		SubmissionHandler:  subHandler,
	})

	addr := ":" + cfg.AppPort
	slog.Info("listening", "addr", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
