package proof

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/upload"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

// HandleGetProofDesignDownload returns a short-lived presigned URL for a
// proof's design asset, with a Content-Disposition that forces a download.
//
// The plain asset URL cannot do this from the browser: the <a download>
// attribute is ignored cross-origin, so linking straight at the S3 object just
// renders the SVG in a tab instead of saving it.
func (m ProofModule) HandleGetProofDesignDownload(w http.ResponseWriter, r *http.Request) {
	proof, inlay, project, ok := m.loadProofWithContext(w, r)
	if !ok {
		return
	}

	if proof.DesignAssetURL == "" {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	filename := buildProofDesignFilename(project, inlay, proof)
	key := strings.TrimPrefix(proof.DesignAssetURL, "/")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url, err := upload.GenerateSignedDownloadURL(ctx, m.S3, m.Cfg, key, filename, 15*time.Minute)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to sign design download for proof %d: %w", proof.ID, err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]string{"url": url})
}

// buildProofDesignFilename names the download after the project, the inlay and
// the proof version, e.g. "Smith-Memorial_Dove_v2.svg", so a designer with
// several versions on disk can tell them apart.
func buildProofDesignFilename(project *data.Project, inlay *data.Inlay, proof *data.InlayProof) string {
	ext := filepath.Ext(proof.DesignAssetURL)
	if ext == "" {
		ext = ".svg"
	}

	return fmt.Sprintf("%s_%s_v%d%s",
		upload.SanitizeFilenamePart(project.Name, "project"),
		upload.SanitizeFilenamePart(inlay.Name, "inlay"),
		proof.VersionNumber, ext)
}
