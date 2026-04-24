package dashboard

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
)

type DashboardModule struct {
	*app.Application
}

func NewDashboardModule(app *app.Application) *DashboardModule {
	return &DashboardModule{app}
}

func (m *DashboardModule) HandleGetDealershipDashboard(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	if !user.IsDealership() {
		m.WriteError(w, r, m.Err.Forbidden, nil)
		return
	}

	dealershipID := user.GetDealershipID()
	if dealershipID == nil {
		m.WriteError(w, r, m.Err.Forbidden, nil)
		return
	}

	dashboard, err := m.Db.Dashboard.GetDealershipDashboard(*dealershipID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, dashboard)
}

func (m *DashboardModule) HandleGetInternalDashboard(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	if !user.IsInternal() {
		m.WriteError(w, r, m.Err.Forbidden, nil)
		return
	}

	dashboard, err := m.Db.Dashboard.GetInternalDashboard()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, dashboard)
}
