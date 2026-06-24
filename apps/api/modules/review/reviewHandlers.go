package review

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type ReviewModule struct {
	*app.Application
}

func NewReviewModule(app *app.Application) *ReviewModule {
	return &ReviewModule{app}
}

// reviewItem is one entry in the internal review queue: an inlay that needs
// internal action, with the project it belongs to and (for approval items) the
// pending proof to act on.
type reviewItem struct {
	ProjectUUID  string           `json:"project_uuid"`
	ProjectName  string           `json:"project_name"`
	Inlay        *data.Inlay      `json:"inlay"`
	PendingProof *data.InlayProof `json:"pending_proof,omitempty"`
}

type reviewQueue struct {
	NeedsApproval []reviewItem `json:"needs_approval"`
	NeedsProof    []reviewItem `json:"needs_proof"`
}

// HandleGetReviewQueue returns, for internal users only, every inlay across all
// projects awaiting internal action: customized catalog proofs to approve and
// custom inlays still needing a proof.
func (m *ReviewModule) HandleGetReviewQueue(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	if !user.IsInternal() {
		m.WriteError(w, r, m.Err.Forbidden, nil)
		return
	}

	projectCache := make(map[int]*data.Project)
	resolveProject := func(id int) (*data.Project, error) {
		if p, ok := projectCache[id]; ok {
			return p, nil
		}
		p, found, err := m.Db.Projects.GetByID(id)
		if err != nil {
			return nil, err
		}
		if !found {
			projectCache[id] = nil
			return nil, nil
		}
		projectCache[id] = p
		return p, nil
	}

	approvalInlays, err := m.Db.Inlays.GetNeedingInternalApproval()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	needsApproval := make([]reviewItem, 0, len(approvalInlays))
	for _, inlay := range approvalInlays {
		project, err := resolveProject(inlay.ProjectID)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
		if project == nil {
			continue
		}

		item := reviewItem{
			ProjectUUID: project.UUID,
			ProjectName: project.Name,
			Inlay:       inlay,
		}

		proof, found, err := m.Db.InlayProofs.GetLatestByInlayID(inlay.ID)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
		if found {
			item.PendingProof = proof
		}

		needsApproval = append(needsApproval, item)
	}

	proofInlays, err := m.Db.Inlays.GetCustomNeedingProof()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	needsProof := make([]reviewItem, 0, len(proofInlays))
	for _, inlay := range proofInlays {
		project, err := resolveProject(inlay.ProjectID)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
		if project == nil {
			continue
		}

		needsProof = append(needsProof, reviewItem{
			ProjectUUID: project.UUID,
			ProjectName: project.Name,
			Inlay:       inlay,
		})
	}

	m.WriteJSON(w, r, http.StatusOK, reviewQueue{
		NeedsApproval: needsApproval,
		NeedsProof:    needsProof,
	})
}
