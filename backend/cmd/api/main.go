// cmd/api/main.go

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

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
	programHandler := httpapi.NewProgramHandler(catalogRepo)
	teacherHandler := httpapi.NewTeacherHandler(catalogRepo, appRepo, invSvc)

	matRepo := repo.NewMaterialRepo(pool)
	fileRepo := repo.NewFileRepo(pool)
	matSvc := service.NewMaterialService(matRepo, appRepo, catalogRepo, fileRepo)
	matHandler := httpapi.NewMaterialHandler(matSvc)

	asgRepo := repo.NewAssignmentRepo(pool)
	subRepo := repo.NewSubmissionRepo(pool)

	asgSvc := service.NewAssignmentService(catalogRepo, appRepo, asgRepo)
	subSvc := service.NewSubmissionService(catalogRepo, appRepo, asgRepo, subRepo)

	asgHandler := httpapi.NewAssignmentHandler(asgSvc)
	subHandler := httpapi.NewSubmissionHandler(subSvc)

	var filesHandler *httpapi.FilesHandler

	// ---- files (MinIO) ----
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccess := os.Getenv("MINIO_ACCESS_KEY")
	minioSecret := os.Getenv("MINIO_SECRET_KEY")
	minioBucket := os.Getenv("MINIO_BUCKET")
	minioUseSSLStr := os.Getenv("MINIO_USE_SSL")
	minioUseSSL := false
	if minioUseSSLStr != "" {
		v, _ := strconv.ParseBool(minioUseSSLStr)
		minioUseSSL = v
	}

	if minioEndpoint != "" && minioAccess != "" && minioSecret != "" && minioBucket != "" {
		mc, err := minio.New(minioEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(minioAccess, minioSecret, ""),
			Secure: minioUseSSL,
		})
		if err != nil {
			slog.Error("minio init", "err", err)
			os.Exit(1)
		}

		fileSvc := service.NewFileService(mc, minioBucket, fileRepo)
		filesHandler = httpapi.NewFilesHandler(fileSvc)
	} else {
		slog.Warn("minio disabled: env vars not set")
	}

	router := httpapi.NewRouter(httpapi.Deps{
		ApplicationHandler: appHandler,
		CatalogHandler:     catalogHandler,
		ProgramHandler:     programHandler,
		TeacherHandler:     teacherHandler,
		MaterialHandler:    matHandler,
		AssignmentHandler:  asgHandler,
		SubmissionHandler:  subHandler,
		FilesHandler:       filesHandler,
	})

	addr := ":" + cfg.AppPort
	slog.Info("listening", "addr", addr)

	if err := http.ListenAndServe(addr, router); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
