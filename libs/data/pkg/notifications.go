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
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationEventType string

type notificationEventTypes struct {
	ProofReady       NotificationEventType
	ProofApproved    NotificationEventType
	ProofDeclined    NotificationEventType
	OrderPlaced      NotificationEventType
	InlayStepChanged NotificationEventType
	InlayBlocked     NotificationEventType
	InlayUnblocked   NotificationEventType
	ProjectShipped   NotificationEventType
	ProjectDelivered NotificationEventType
	InvoiceSent      NotificationEventType
	PaymentReceived  NotificationEventType
	ChatMessage      NotificationEventType
}

var NotificationEventTypes = notificationEventTypes{
	ProofReady:       NotificationEventType("proof_ready"),
	ProofApproved:    NotificationEventType("proof_approved"),
	ProofDeclined:    NotificationEventType("proof_declined"),
	OrderPlaced:      NotificationEventType("order_placed"),
	InlayStepChanged: NotificationEventType("inlay_step_changed"),
	InlayBlocked:     NotificationEventType("inlay_blocked"),
	InlayUnblocked:   NotificationEventType("inlay_unblocked"),
	ProjectShipped:   NotificationEventType("project_shipped"),
	ProjectDelivered: NotificationEventType("project_delivered"),
	InvoiceSent:      NotificationEventType("invoice_sent"),
	PaymentReceived:  NotificationEventType("payment_received"),
	ChatMessage:      NotificationEventType("chat_message"),
}

type Notification struct {
	ID               int                   `json:"id"`
	UUID             string                `json:"uuid"`
	DealershipUserID *int                  `json:"dealership_user_id"`
	InternalUserID   *int                  `json:"internal_user_id"`
	EventType        NotificationEventType `json:"event_type"`
	Title            string                `json:"title"`
	Body             string                `json:"body"`
	ProjectID        *int                  `json:"project_id"`
	InlayID          *int                  `json:"inlay_id"`
	ReadAt           *time.Time            `json:"read_at"`
	EmailSentAt      *time.Time            `json:"email_sent_at"`
	CreatedAt        time.Time             `json:"created_at"`
}

type NotificationModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func notificationFromGen(genNotif model.Notifications) *Notification {
	var dealershipUserID *int
	if genNotif.DealershipUserID != nil {
		dealershipUserIDVal := int(*genNotif.DealershipUserID)
		dealershipUserID = &dealershipUserIDVal
	}

	var internalUserID *int
	if genNotif.InternalUserID != nil {
		internalUserIDVal := int(*genNotif.InternalUserID)
		internalUserID = &internalUserIDVal
	}

	var projectID *int
	if genNotif.ProjectID != nil {
		projectIDVal := int(*genNotif.ProjectID)
		projectID = &projectIDVal
	}

	var inlayID *int
	if genNotif.InlayID != nil {
		inlayIDVal := int(*genNotif.InlayID)
		inlayID = &inlayIDVal
	}

	notif := Notification{
		ID:               int(genNotif.ID),
		UUID:             genNotif.UUID.String(),
		DealershipUserID: dealershipUserID,
		InternalUserID:   internalUserID,
		EventType:        NotificationEventType(genNotif.EventType),
		Title:            genNotif.Title,
		Body:             genNotif.Body,
		ProjectID:        projectID,
		InlayID:          inlayID,
		ReadAt:           genNotif.ReadAt,
		EmailSentAt:      genNotif.EmailSentAt,
		CreatedAt:        genNotif.CreatedAt,
	}

	return &notif
}

func notificationToGen(n *Notification) (*model.Notifications, error) {
	var notifUUID uuid.UUID
	var err error

	if n.UUID != "" {
		notifUUID, err = uuid.Parse(n.UUID)
		if err != nil {
			return nil, err
		}
	}

	var dealershipUserID *int32
	if n.DealershipUserID != nil {
		dealershipUserIDVal := int32(*n.DealershipUserID)
		dealershipUserID = &dealershipUserIDVal
	}

	var internalUserID *int32
	if n.InternalUserID != nil {
		internalUserIDVal := int32(*n.InternalUserID)
		internalUserID = &internalUserIDVal
	}

	var projectID *int32
	if n.ProjectID != nil {
		projectIDVal := int32(*n.ProjectID)
		projectID = &projectIDVal
	}

	var inlayID *int32
	if n.InlayID != nil {
		inlayIDVal := int32(*n.InlayID)
		inlayID = &inlayIDVal
	}

	genNotif := model.Notifications{
		ID:               int32(n.ID),
		UUID:             notifUUID,
		DealershipUserID: dealershipUserID,
		InternalUserID:   internalUserID,
		EventType:        model.NotificationEventType(n.EventType),
		Title:            n.Title,
		Body:             n.Body,
		ProjectID:        projectID,
		InlayID:          inlayID,
		ReadAt:           n.ReadAt,
		EmailSentAt:      n.EmailSentAt,
		CreatedAt:        n.CreatedAt,
	}

	return &genNotif, nil
}

func (m NotificationModel) Insert(notif *Notification) error {
	genNotif, err := notificationToGen(notif)
	if err != nil {
		return err
	}

	query := table.Notifications.INSERT(
		table.Notifications.DealershipUserID,
		table.Notifications.InternalUserID,
		table.Notifications.EventType,
		table.Notifications.Title,
		table.Notifications.Body,
		table.Notifications.ProjectID,
		table.Notifications.InlayID,
	).MODEL(
		genNotif,
	).RETURNING(
		table.Notifications.ID,
		table.Notifications.UUID,
		table.Notifications.CreatedAt,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Notifications
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	notif.ID = int(dest.ID)
	notif.UUID = dest.UUID.String()
	notif.CreatedAt = dest.CreatedAt

	return nil
}

func (m NotificationModel) GetByID(id int) (*Notification, bool, error) {
	query := postgres.SELECT(
		table.Notifications.AllColumns,
	).FROM(
		table.Notifications,
	).WHERE(
		table.Notifications.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Notifications
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return notificationFromGen(dest), true, nil
}

func (m NotificationModel) GetByUUID(uuidStr string) (*Notification, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.Notifications.AllColumns,
	).FROM(
		table.Notifications,
	).WHERE(
		table.Notifications.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Notifications
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return notificationFromGen(dest), true, nil
}

func (m NotificationModel) GetForDealershipUser(dealershipUserID int) ([]*Notification, error) {
	query := postgres.SELECT(
		table.Notifications.AllColumns,
	).FROM(
		table.Notifications,
	).WHERE(
		table.Notifications.DealershipUserID.EQ(postgres.Int(int64(dealershipUserID))),
	).ORDER_BY(
		table.Notifications.CreatedAt.DESC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Notifications
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	notifs := make([]*Notification, len(dest))
	for i, d := range dest {
		notifs[i] = notificationFromGen(d)
	}

	return notifs, nil
}

func (m NotificationModel) GetForInternalUser(internalUserID int) ([]*Notification, error) {
	query := postgres.SELECT(
		table.Notifications.AllColumns,
	).FROM(
		table.Notifications,
	).WHERE(
		table.Notifications.InternalUserID.EQ(postgres.Int(int64(internalUserID))),
	).ORDER_BY(
		table.Notifications.CreatedAt.DESC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Notifications
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	notifs := make([]*Notification, len(dest))
	for i, d := range dest {
		notifs[i] = notificationFromGen(d)
	}

	return notifs, nil
}

func (m NotificationModel) GetUnreadForDealershipUser(dealershipUserID int) ([]*Notification, error) {
	query := postgres.SELECT(
		table.Notifications.AllColumns,
	).FROM(
		table.Notifications,
	).WHERE(
		postgres.AND(
			table.Notifications.DealershipUserID.EQ(postgres.Int(int64(dealershipUserID))),
			table.Notifications.ReadAt.IS_NULL(),
		),
	).ORDER_BY(
		table.Notifications.CreatedAt.DESC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Notifications
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	notifs := make([]*Notification, len(dest))
	for i, d := range dest {
		notifs[i] = notificationFromGen(d)
	}

	return notifs, nil
}

func (m NotificationModel) GetUnreadForInternalUser(internalUserID int) ([]*Notification, error) {
	query := postgres.SELECT(
		table.Notifications.AllColumns,
	).FROM(
		table.Notifications,
	).WHERE(
		postgres.AND(
			table.Notifications.InternalUserID.EQ(postgres.Int(int64(internalUserID))),
			table.Notifications.ReadAt.IS_NULL(),
		),
	).ORDER_BY(
		table.Notifications.CreatedAt.DESC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Notifications
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	notifs := make([]*Notification, len(dest))
	for i, d := range dest {
		notifs[i] = notificationFromGen(d)
	}

	return notifs, nil
}

func (m NotificationModel) GetAll() ([]*Notification, error) {
	query := postgres.SELECT(
		table.Notifications.AllColumns,
	).FROM(
		table.Notifications,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Notifications
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	notifs := make([]*Notification, len(dest))
	for i, d := range dest {
		notifs[i] = notificationFromGen(d)
	}

	return notifs, nil
}

func (m NotificationModel) MarkRead(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.STDB.ExecContext(ctx,
		"UPDATE notifications SET read_at = now() WHERE id = $1",
		id,
	)
	return err
}

func (m NotificationModel) MarkAllReadForDealershipUser(dealershipUserID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.STDB.ExecContext(ctx,
		"UPDATE notifications SET read_at = now() WHERE dealership_user_id = $1 AND read_at IS NULL",
		dealershipUserID,
	)
	return err
}

func (m NotificationModel) MarkAllReadForInternalUser(internalUserID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.STDB.ExecContext(ctx,
		"UPDATE notifications SET read_at = now() WHERE internal_user_id = $1 AND read_at IS NULL",
		internalUserID,
	)
	return err
}

func (m NotificationModel) MarkEmailSent(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.STDB.ExecContext(ctx,
		"UPDATE notifications SET email_sent_at = now() WHERE id = $1",
		id,
	)
	return err
}

func (m NotificationModel) Delete(id int) error {
	query := table.Notifications.DELETE().WHERE(
		table.Notifications.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
