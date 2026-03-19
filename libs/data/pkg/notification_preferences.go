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

type NotificationPreference struct {
	ID               int                   `json:"id"`
	DealershipUserID *int                  `json:"dealership_user_id"`
	InternalUserID   *int                  `json:"internal_user_id"`
	EventType        NotificationEventType `json:"event_type"`
	EmailEnabled     bool                  `json:"email_enabled"`
}

type NotificationPreferencesModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func notificationPrefFromDealershipGen(gen model.DealershipUserNotificationPrefs) *NotificationPreference {
	id := int(gen.DealershipUserID)
	return &NotificationPreference{
		ID:               int(gen.ID),
		DealershipUserID: &id,
		EventType:        NotificationEventType(gen.EventType),
		EmailEnabled:     gen.EmailEnabled,
	}
}

func notificationPrefFromInternalGen(gen model.InternalUserNotificationPrefs) *NotificationPreference {
	id := int(gen.InternalUserID)
	return &NotificationPreference{
		ID:             int(gen.ID),
		InternalUserID: &id,
		EventType:      NotificationEventType(gen.EventType),
		EmailEnabled:   gen.EmailEnabled,
	}
}

func (m NotificationPreferencesModel) GetForDealershipUser(userID int) ([]*NotificationPreference, error) {
	query := postgres.SELECT(
		table.DealershipUserNotificationPrefs.AllColumns,
	).FROM(
		table.DealershipUserNotificationPrefs,
	).WHERE(
		table.DealershipUserNotificationPrefs.DealershipUserID.EQ(postgres.Int(int64(userID))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.DealershipUserNotificationPrefs
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	prefs := make([]*NotificationPreference, len(dest))
	for i, d := range dest {
		prefs[i] = notificationPrefFromDealershipGen(d)
	}

	return prefs, nil
}

func (m NotificationPreferencesModel) GetForInternalUser(userID int) ([]*NotificationPreference, error) {
	query := postgres.SELECT(
		table.InternalUserNotificationPrefs.AllColumns,
	).FROM(
		table.InternalUserNotificationPrefs,
	).WHERE(
		table.InternalUserNotificationPrefs.InternalUserID.EQ(postgres.Int(int64(userID))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InternalUserNotificationPrefs
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	prefs := make([]*NotificationPreference, len(dest))
	for i, d := range dest {
		prefs[i] = notificationPrefFromInternalGen(d)
	}

	return prefs, nil
}

func (m NotificationPreferencesModel) UpsertForDealershipUser(userID int, eventType NotificationEventType, emailEnabled bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.STDB.ExecContext(ctx,
		`INSERT INTO dealership_user_notification_prefs (dealership_user_id, event_type, email_enabled)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (dealership_user_id, event_type) DO UPDATE SET email_enabled = $3`,
		userID, string(eventType), emailEnabled,
	)
	return err
}

func (m NotificationPreferencesModel) UpsertForInternalUser(userID int, eventType NotificationEventType, emailEnabled bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.STDB.ExecContext(ctx,
		`INSERT INTO internal_user_notification_prefs (internal_user_id, event_type, email_enabled)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (internal_user_id, event_type) DO UPDATE SET email_enabled = $3`,
		userID, string(eventType), emailEnabled,
	)
	return err
}

func (m NotificationPreferencesModel) IsEmailEnabledForDealershipUser(userID int, eventType NotificationEventType) (bool, error) {
	query := postgres.SELECT(
		table.DealershipUserNotificationPrefs.AllColumns,
	).FROM(
		table.DealershipUserNotificationPrefs,
	).WHERE(
		postgres.AND(
			table.DealershipUserNotificationPrefs.DealershipUserID.EQ(postgres.Int(int64(userID))),
			table.DealershipUserNotificationPrefs.EventType.EQ(postgres.String(string(eventType))),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipUserNotificationPrefs
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return false, err
	}

	return dest.EmailEnabled, nil
}

func (m NotificationPreferencesModel) IsEmailEnabledForInternalUser(userID int, eventType NotificationEventType) (bool, error) {
	query := postgres.SELECT(
		table.InternalUserNotificationPrefs.AllColumns,
	).FROM(
		table.InternalUserNotificationPrefs,
	).WHERE(
		postgres.AND(
			table.InternalUserNotificationPrefs.InternalUserID.EQ(postgres.Int(int64(userID))),
			table.InternalUserNotificationPrefs.EventType.EQ(postgres.String(string(eventType))),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalUserNotificationPrefs
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) || errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return false, err
	}

	return dest.EmailEnabled, nil
}
