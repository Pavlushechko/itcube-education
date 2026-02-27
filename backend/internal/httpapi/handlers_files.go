package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/Pavlushechko/itcube-education/internal/service"
)

type FilesHandler struct {
	svc *service.FileService
}

func NewFilesHandler(svc *service.FileService) *FilesHandler {
	return &FilesHandler{svc: svc}
}

type uploadURLReq struct {
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	SizeBytes    int64  `json:"size_bytes"`
}

// internal/httpapi/handlers_files.go
func (h *FilesHandler) CreateUploadURL(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.svc == nil {
		http.Error(w, "files service not initialized", http.StatusInternalServerError)
		return
	}

	var req uploadURLReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	res, err := h.svc.CreateUploadURL(r.Context(), req.OriginalName, req.MimeType, req.SizeBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

func (h *FilesHandler) DownloadURL(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}
	u, err := h.svc.GetDownloadURL(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"url": u})
}

func (h *FilesHandler) Download(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid file id", http.StatusBadRequest)
		return
	}
	if h == nil || h.svc == nil {
		http.Error(w, "files service not initialized", http.StatusInternalServerError)
		return
	}

	// svc сам выставит content-type + disposition и стримит тело
	if err := h.svc.StreamDownload(r.Context(), w, id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *FilesHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.svc == nil {
		http.Error(w, "files service not initialized", 500)
		return
	}

	// max 50MB (подгони)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "bad multipart", 400)
		return
	}

	f, hdr, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", 400)
		return
	}
	defer f.Close()

	mime := hdr.Header.Get("Content-Type")
	id, err := h.svc.UploadFile(r.Context(), hdr.Filename, mime, hdr.Size, f)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	writeJSON(w, 201, map[string]any{"file_id": id.String()})
}
