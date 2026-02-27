package upload

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
)

type UploadModule struct {
	*app.Application
}

func NewUploadModule(app *app.Application) *UploadModule {
	return &UploadModule{
		app,
	}
}

func (m UploadModule) HandlePostUpload(w http.ResponseWriter, r *http.Request) {
	const maxFileSize = 50 << 20 // 50MB
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	err := r.ParseMultipartForm(maxFileSize)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("failed to parse multipart form: %w", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("file not found in request: %w", err))
		return
	}
	defer file.Close()

	if header.Size > maxFileSize {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("file size exceeds maximum allowed size of 50MB"))
		return
	}

	if header.Size == 0 {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("file is empty"))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	uploadPath := r.FormValue("uploadPath")
	if uploadPath == "" {
		uploadPath = "uploads"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := UploadFileToS3(
		ctx,
		m.S3,
		m.Cfg,
		file,
		header.Filename,
		header.Size,
		contentType,
		uploadPath,
	)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, result)
}

func (m UploadModule) HandleGetFile(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")
	if path == "" {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("file path is required"))
		return
	}

	path = strings.TrimPrefix(path, "/")
	if path == "" {
		m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("file path cannot be empty"))
		return
	}

	key := fmt.Sprintf("file/%s", path)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	signedURL, err := GenerateSignedURL(ctx, m.S3, m.Cfg, key, 15*time.Minute)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	http.Redirect(w, r, signedURL, http.StatusFound)
}
