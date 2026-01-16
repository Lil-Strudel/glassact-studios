package upload

import (
	"context"
	"fmt"
	"net/http"
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
	)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, result)
}
