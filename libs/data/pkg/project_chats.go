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

type ProjectChat struct {
	StandardTable
	ProjectID        int             `json:"project_id"`
	DealershipUserID *int            `json:"dealership_user_id"`
	InternalUserID   *int            `json:"internal_user_id"`
	MessageType      ChatMessageType `json:"message_type"`
	Message          string          `json:"message"`
	AttachmentURL    *string         `json:"attachment_url"`
}

type ProjectChatModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func projectChatFromGen(genChat model.ProjectChats) *ProjectChat {
	var dealershipUserID *int
	if genChat.DealershipUserID != nil {
		dealershipUserIDVal := int(*genChat.DealershipUserID)
		dealershipUserID = &dealershipUserIDVal
	}

	var internalUserID *int
	if genChat.InternalUserID != nil {
		internalUserIDVal := int(*genChat.InternalUserID)
		internalUserID = &internalUserIDVal
	}

	chat := ProjectChat{
		StandardTable: StandardTable{
			ID:        int(genChat.ID),
			UUID:      genChat.UUID.String(),
			CreatedAt: genChat.CreatedAt,
			UpdatedAt: genChat.UpdatedAt,
			Version:   int(genChat.Version),
		},
		ProjectID:        int(genChat.ProjectID),
		DealershipUserID: dealershipUserID,
		InternalUserID:   internalUserID,
		MessageType:      ChatMessageType(genChat.MessageType),
		Message:          genChat.Message,
		AttachmentURL:    genChat.AttachmentURL,
	}

	return &chat
}

func projectChatToGen(pc *ProjectChat) (*model.ProjectChats, error) {
	var chatUUID uuid.UUID
	var err error

	if pc.UUID != "" {
		chatUUID, err = uuid.Parse(pc.UUID)
		if err != nil {
			return nil, err
		}
	}

	var dealershipUserID *int32
	if pc.DealershipUserID != nil {
		dealershipUserIDVal := int32(*pc.DealershipUserID)
		dealershipUserID = &dealershipUserIDVal
	}

	var internalUserID *int32
	if pc.InternalUserID != nil {
		internalUserIDVal := int32(*pc.InternalUserID)
		internalUserID = &internalUserIDVal
	}

	genChat := model.ProjectChats{
		ID:               int32(pc.ID),
		UUID:             chatUUID,
		ProjectID:        int32(pc.ProjectID),
		DealershipUserID: dealershipUserID,
		InternalUserID:   internalUserID,
		MessageType:      string(pc.MessageType),
		Message:          pc.Message,
		AttachmentURL:    pc.AttachmentURL,
		UpdatedAt:        pc.UpdatedAt,
		CreatedAt:        pc.CreatedAt,
		Version:          int32(pc.Version),
	}

	return &genChat, nil
}

func (m ProjectChatModel) Insert(chat *ProjectChat) error {
	genChat, err := projectChatToGen(chat)
	if err != nil {
		return err
	}

	query := table.ProjectChats.INSERT(
		table.ProjectChats.ProjectID,
		table.ProjectChats.DealershipUserID,
		table.ProjectChats.InternalUserID,
		table.ProjectChats.MessageType,
		table.ProjectChats.Message,
		table.ProjectChats.AttachmentURL,
	).MODEL(
		genChat,
	).RETURNING(
		table.ProjectChats.ID,
		table.ProjectChats.UUID,
		table.ProjectChats.UpdatedAt,
		table.ProjectChats.CreatedAt,
		table.ProjectChats.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.ProjectChats
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	chat.ID = int(dest.ID)
	chat.UUID = dest.UUID.String()
	chat.UpdatedAt = dest.UpdatedAt
	chat.CreatedAt = dest.CreatedAt
	chat.Version = int(dest.Version)

	return nil
}

func (m ProjectChatModel) GetByID(id int) (*ProjectChat, bool, error) {
	query := postgres.SELECT(
		table.ProjectChats.AllColumns,
	).FROM(
		table.ProjectChats,
	).WHERE(
		table.ProjectChats.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.ProjectChats
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return projectChatFromGen(dest), true, nil
}

func (m ProjectChatModel) GetByUUID(uuidStr string) (*ProjectChat, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.ProjectChats.AllColumns,
	).FROM(
		table.ProjectChats,
	).WHERE(
		table.ProjectChats.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.ProjectChats
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return projectChatFromGen(dest), true, nil
}

func (m ProjectChatModel) GetByProjectID(projectID int) ([]*ProjectChat, error) {
	query := postgres.SELECT(
		table.ProjectChats.AllColumns,
	).FROM(
		table.ProjectChats,
	).WHERE(
		table.ProjectChats.ProjectID.EQ(postgres.Int(int64(projectID))),
	).ORDER_BY(
		table.ProjectChats.CreatedAt.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.ProjectChats
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	chats := make([]*ProjectChat, len(dest))
	for i, d := range dest {
		chats[i] = projectChatFromGen(d)
	}

	return chats, nil
}

func (m ProjectChatModel) GetAll() ([]*ProjectChat, error) {
	query := postgres.SELECT(
		table.ProjectChats.AllColumns,
	).FROM(
		table.ProjectChats,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.ProjectChats
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	chats := make([]*ProjectChat, len(dest))
	for i, d := range dest {
		chats[i] = projectChatFromGen(d)
	}

	return chats, nil
}

func (m ProjectChatModel) Update(chat *ProjectChat) error {
	genChat, err := projectChatToGen(chat)
	if err != nil {
		return err
	}

	query := table.ProjectChats.UPDATE(
		table.ProjectChats.Message,
		table.ProjectChats.AttachmentURL,
		table.ProjectChats.Version,
	).MODEL(
		genChat,
	).WHERE(
		postgres.AND(
			table.ProjectChats.ID.EQ(postgres.Int(int64(chat.ID))),
			table.ProjectChats.Version.EQ(postgres.Int(int64(chat.Version))),
		),
	).RETURNING(
		table.ProjectChats.UpdatedAt,
		table.ProjectChats.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.ProjectChats
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	chat.UpdatedAt = dest.UpdatedAt
	chat.Version = int(dest.Version)

	return nil
}

func (m ProjectChatModel) Delete(id int) error {
	query := table.ProjectChats.DELETE().WHERE(
		table.ProjectChats.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
