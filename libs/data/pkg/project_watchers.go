package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/model"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg/gen/glassact/public/table"
	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProjectWatcher subscribes a single user to a project's notifications. A
// missing row means the user has never been subscribed; IsWatching = false
// means they explicitly unwatched and auto-subscription must leave them alone.
type ProjectWatcher struct {
	StandardTable
	ProjectID        int  `json:"project_id"`
	DealershipUserID *int `json:"dealership_user_id"`
	InternalUserID   *int `json:"internal_user_id"`
	IsWatching       bool `json:"is_watching"`
}

type ProjectWatcherModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func projectWatcherFromGen(gen model.ProjectWatchers) *ProjectWatcher {
	watcher := &ProjectWatcher{
		StandardTable: StandardTable{
			ID:        int(gen.ID),
			UUID:      gen.UUID.String(),
			CreatedAt: gen.CreatedAt,
			UpdatedAt: gen.UpdatedAt,
			Version:   int(gen.Version),
		},
		ProjectID:  int(gen.ProjectID),
		IsWatching: gen.IsWatching,
	}

	if gen.DealershipUserID != nil {
		dealershipUserID := int(*gen.DealershipUserID)
		watcher.DealershipUserID = &dealershipUserID
	}
	if gen.InternalUserID != nil {
		internalUserID := int(*gen.InternalUserID)
		watcher.InternalUserID = &internalUserID
	}

	return watcher
}

// userColumn returns the watcher column that holds the given user's id. The two
// nullable columns are mutually exclusive (project_watchers_user_check), and
// each has its own partial unique index, so every statement below has to target
// the one matching the user's type.
func userColumn(user AuthUser) string {
	if user.IsDealership() {
		return "dealership_user_id"
	}
	return "internal_user_id"
}

// autoSubscribeQuery subscribes a user without ever overwriting an existing
// row, so an explicit unwatch survives later activity on the project.
func autoSubscribeQuery(user AuthUser) string {
	column := userColumn(user)
	return `INSERT INTO project_watchers (project_id, ` + column + `, is_watching)
	        VALUES ($1, $2, true)
	        ON CONFLICT (project_id, ` + column + `) WHERE ` + column + ` IS NOT NULL DO NOTHING`
}

func (m ProjectWatcherModel) AutoSubscribe(projectID int, user AuthUser) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.STDB.ExecContext(ctx, autoSubscribeQuery(user), projectID, user.GetID())
	return err
}

func (m ProjectWatcherModel) TxAutoSubscribe(tx *sql.Tx, projectID int, user AuthUser) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, autoSubscribeQuery(user), projectID, user.GetID())
	return err
}

// SetWatching records an explicit choice, which does override whatever
// auto-subscription previously decided.
func (m ProjectWatcherModel) SetWatching(projectID int, user AuthUser, isWatching bool) error {
	column := userColumn(user)
	query := `INSERT INTO project_watchers (project_id, ` + column + `, is_watching)
	          VALUES ($1, $2, $3)
	          ON CONFLICT (project_id, ` + column + `) WHERE ` + column + ` IS NOT NULL
	          DO UPDATE SET is_watching = $3`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.STDB.ExecContext(ctx, query, projectID, user.GetID(), isWatching)
	return err
}

func (m ProjectWatcherModel) userIDMatches(user AuthUser) postgres.BoolExpression {
	if user.IsDealership() {
		return table.ProjectWatchers.DealershipUserID.EQ(postgres.Int(int64(user.GetID())))
	}
	return table.ProjectWatchers.InternalUserID.EQ(postgres.Int(int64(user.GetID())))
}

func (m ProjectWatcherModel) GetForUser(projectID int, user AuthUser) (*ProjectWatcher, bool, error) {
	query := postgres.SELECT(
		table.ProjectWatchers.AllColumns,
	).FROM(
		table.ProjectWatchers,
	).WHERE(
		postgres.AND(
			table.ProjectWatchers.ProjectID.EQ(postgres.Int(int64(projectID))),
			m.userIDMatches(user),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.ProjectWatchers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return projectWatcherFromGen(dest), true, nil
}

// IsWatching reports whether the user currently receives this project's
// notifications. A user with no row is not watching.
func (m ProjectWatcherModel) IsWatching(projectID int, user AuthUser) (bool, error) {
	watcher, found, err := m.GetForUser(projectID, user)
	if err != nil || !found {
		return false, err
	}
	return watcher.IsWatching, nil
}

// GetDealershipWatchers returns the active dealership users watching a project.
func (m ProjectWatcherModel) GetDealershipWatchers(projectID int) ([]*DealershipUser, error) {
	query := postgres.SELECT(
		table.DealershipUsers.AllColumns,
	).FROM(
		table.ProjectWatchers.INNER_JOIN(
			table.DealershipUsers,
			table.DealershipUsers.ID.EQ(table.ProjectWatchers.DealershipUserID),
		),
	).WHERE(
		postgres.AND(
			table.ProjectWatchers.ProjectID.EQ(postgres.Int(int64(projectID))),
			table.ProjectWatchers.IsWatching.IS_TRUE(),
			table.DealershipUsers.IsActive.IS_TRUE(),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.DealershipUsers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	users := make([]*DealershipUser, len(dest))
	for i, d := range dest {
		users[i] = dealershipUserFromGen(d)
	}

	return users, nil
}

// GetInternalWatchers returns the active internal users watching a project.
func (m ProjectWatcherModel) GetInternalWatchers(projectID int) ([]*InternalUser, error) {
	query := postgres.SELECT(
		table.InternalUsers.AllColumns,
	).FROM(
		table.ProjectWatchers.INNER_JOIN(
			table.InternalUsers,
			table.InternalUsers.ID.EQ(table.ProjectWatchers.InternalUserID),
		),
	).WHERE(
		postgres.AND(
			table.ProjectWatchers.ProjectID.EQ(postgres.Int(int64(projectID))),
			table.ProjectWatchers.IsWatching.IS_TRUE(),
			table.InternalUsers.IsActive.IS_TRUE(),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InternalUsers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	users := make([]*InternalUser, len(dest))
	for i, d := range dest {
		users[i] = internalUserFromGen(d)
	}

	return users, nil
}

// CountForProject counts the active watchers on both sides, for the watch
// button's badge.
func (m ProjectWatcherModel) CountForProject(projectID int) (int, error) {
	dealershipWatchers, err := m.GetDealershipWatchers(projectID)
	if err != nil {
		return 0, err
	}

	internalWatchers, err := m.GetInternalWatchers(projectID)
	if err != nil {
		return 0, err
	}

	return len(dealershipWatchers) + len(internalWatchers), nil
}
