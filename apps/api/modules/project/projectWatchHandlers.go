package project

import (
	"fmt"
	"net/http"
)

// projectWatcherSummary identifies a watcher to other people on the project.
// It deliberately omits the email address, since dealership users would
// otherwise see internal staff contact details.
type projectWatcherSummary struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
	UserType string `json:"user_type"`
}

type watchStateResponse struct {
	IsWatching   bool `json:"is_watching"`
	WatcherCount int  `json:"watcher_count"`
}

// HandlePutProjectWatch records an explicit subscribe/unsubscribe. Watching is
// not permission-gated: anyone who can read the project can choose to follow it.
func (m ProjectModule) HandlePutProjectWatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IsWatching *bool `json:"is_watching" validate:"required"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	project, ok := m.getProjectWithAccessCheck(w, r)
	if !ok {
		return
	}

	user := m.ContextGetUser(r)

	err = m.Db.ProjectWatchers.SetWatching(project.ID, user, *body.IsWatching)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to set watch state for project %d: %w", project.ID, err))
		return
	}

	watcherCount, err := m.Db.ProjectWatchers.CountForProject(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to count watchers for project %d: %w", project.ID, err))
		return
	}

	m.WriteJSON(w, r, http.StatusOK, watchStateResponse{
		IsWatching:   *body.IsWatching,
		WatcherCount: watcherCount,
	})
}

func (m ProjectModule) HandleGetProjectWatchers(w http.ResponseWriter, r *http.Request) {
	project, ok := m.getProjectWithAccessCheck(w, r)
	if !ok {
		return
	}

	dealershipWatchers, err := m.Db.ProjectWatchers.GetDealershipWatchers(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to get dealership watchers for project %d: %w", project.ID, err))
		return
	}

	internalWatchers, err := m.Db.ProjectWatchers.GetInternalWatchers(project.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, fmt.Errorf("failed to get internal watchers for project %d: %w", project.ID, err))
		return
	}

	summaries := make([]projectWatcherSummary, 0, len(dealershipWatchers)+len(internalWatchers))
	for _, user := range dealershipWatchers {
		summaries = append(summaries, projectWatcherSummary{
			UUID:     user.UUID,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Role:     string(user.Role),
			UserType: "dealership",
		})
	}
	for _, user := range internalWatchers {
		summaries = append(summaries, projectWatcherSummary{
			UUID:     user.UUID,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Role:     string(user.Role),
			UserType: "internal",
		})
	}

	m.WriteJSON(w, r, http.StatusOK, summaries)
}
