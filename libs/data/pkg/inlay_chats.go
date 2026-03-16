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

type ChatMessageType string

type chatMessageTypes struct {
	Text          ChatMessageType
	Image         ChatMessageType
	ProofSent     ChatMessageType
	ProofApproved ChatMessageType
	ProofDeclined ChatMessageType
	System        ChatMessageType
}

var ChatMessageTypes = chatMessageTypes{
	Text:          ChatMessageType("text"),
	Image:         ChatMessageType("image"),
	ProofSent:     ChatMessageType("proof_sent"),
	ProofApproved: ChatMessageType("proof_approved"),
	ProofDeclined: ChatMessageType("proof_declined"),
	System:        ChatMessageType("system"),
}

type InlayChat struct {
	StandardTable
	InlayID          int             `json:"inlay_id"`
	DealershipUserID *int            `json:"dealership_user_id"`
	InternalUserID   *int            `json:"internal_user_id"`
	MessageType      ChatMessageType `json:"message_type"`
	Message          string          `json:"message"`
	AttachmentURL    *string         `json:"attachment_url"`
}

type InlayChatModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func inlayChatFromGen(genChat model.InlayChats) *InlayChat {
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

	chat := InlayChat{
		StandardTable: StandardTable{
			ID:        int(genChat.ID),
			UUID:      genChat.UUID.String(),
			CreatedAt: genChat.CreatedAt,
			UpdatedAt: genChat.UpdatedAt,
			Version:   int(genChat.Version),
		},
		InlayID:          int(genChat.InlayID),
		DealershipUserID: dealershipUserID,
		InternalUserID:   internalUserID,
		MessageType:      ChatMessageType(genChat.MessageType),
		Message:          genChat.Message,
		AttachmentURL:    genChat.AttachmentURL,
	}

	return &chat
}

func inlayChatToGen(ic *InlayChat) (*model.InlayChats, error) {
	var chatUUID uuid.UUID
	var err error

	if ic.UUID != "" {
		chatUUID, err = uuid.Parse(ic.UUID)
		if err != nil {
			return nil, err
		}
	}

	var dealershipUserID *int32
	if ic.DealershipUserID != nil {
		dealershipUserIDVal := int32(*ic.DealershipUserID)
		dealershipUserID = &dealershipUserIDVal
	}

	var internalUserID *int32
	if ic.InternalUserID != nil {
		internalUserIDVal := int32(*ic.InternalUserID)
		internalUserID = &internalUserIDVal
	}

	genChat := model.InlayChats{
		ID:               int32(ic.ID),
		UUID:             chatUUID,
		InlayID:          int32(ic.InlayID),
		DealershipUserID: dealershipUserID,
		InternalUserID:   internalUserID,
		MessageType:      string(ic.MessageType),
		Message:          ic.Message,
		AttachmentURL:    ic.AttachmentURL,
		UpdatedAt:        ic.UpdatedAt,
		CreatedAt:        ic.CreatedAt,
		Version:          int32(ic.Version),
	}

	return &genChat, nil
}

func (m InlayChatModel) insertChat(ctx context.Context, executor qrm.Queryable, chat *InlayChat) error {
	genChat, err := inlayChatToGen(chat)
	if err != nil {
		return err
	}

	query := table.InlayChats.INSERT(
		table.InlayChats.InlayID,
		table.InlayChats.DealershipUserID,
		table.InlayChats.InternalUserID,
		table.InlayChats.MessageType,
		table.InlayChats.Message,
		table.InlayChats.AttachmentURL,
	).MODEL(
		genChat,
	).RETURNING(
		table.InlayChats.ID,
		table.InlayChats.UUID,
		table.InlayChats.UpdatedAt,
		table.InlayChats.CreatedAt,
		table.InlayChats.Version,
	)

	var dest model.InlayChats
	err = query.QueryContext(ctx, executor, &dest)
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

func (m InlayChatModel) Insert(chat *InlayChat) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.insertChat(ctx, m.STDB, chat)
}

func (m InlayChatModel) TxInsert(tx *sql.Tx, chat *InlayChat) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.insertChat(ctx, tx, chat)
}

func (m InlayChatModel) GetByID(id int) (*InlayChat, bool, error) {
	query := postgres.SELECT(
		table.InlayChats.AllColumns,
	).FROM(
		table.InlayChats,
	).WHERE(
		table.InlayChats.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayChats
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayChatFromGen(dest), true, nil
}

func (m InlayChatModel) GetByUUID(uuidStr string) (*InlayChat, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.InlayChats.AllColumns,
	).FROM(
		table.InlayChats,
	).WHERE(
		table.InlayChats.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayChats
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return inlayChatFromGen(dest), true, nil
}

func (m InlayChatModel) GetByInlayID(inlayID int) ([]*InlayChat, error) {
	query := postgres.SELECT(
		table.InlayChats.AllColumns,
	).FROM(
		table.InlayChats,
	).WHERE(
		table.InlayChats.InlayID.EQ(postgres.Int(int64(inlayID))),
	).ORDER_BY(
		table.InlayChats.CreatedAt.ASC(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayChats
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	chats := make([]*InlayChat, len(dest))
	for i, d := range dest {
		chats[i] = inlayChatFromGen(d)
	}

	return chats, nil
}

func (m InlayChatModel) GetAll() ([]*InlayChat, error) {
	query := postgres.SELECT(
		table.InlayChats.AllColumns,
	).FROM(
		table.InlayChats,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayChats
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	chats := make([]*InlayChat, len(dest))
	for i, d := range dest {
		chats[i] = inlayChatFromGen(d)
	}

	return chats, nil
}

func (m InlayChatModel) Update(chat *InlayChat) error {
	genChat, err := inlayChatToGen(chat)
	if err != nil {
		return err
	}

	query := table.InlayChats.UPDATE(
		table.InlayChats.Message,
		table.InlayChats.AttachmentURL,
		table.InlayChats.Version,
	).MODEL(
		genChat,
	).WHERE(
		postgres.AND(
			table.InlayChats.ID.EQ(postgres.Int(int64(chat.ID))),
			table.InlayChats.Version.EQ(postgres.Int(int64(chat.Version))),
		),
	).RETURNING(
		table.InlayChats.UpdatedAt,
		table.InlayChats.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayChats
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	chat.UpdatedAt = dest.UpdatedAt
	chat.Version = int(dest.Version)

	return nil
}

func (m InlayChatModel) Delete(id int) error {
	query := table.InlayChats.DELETE().WHERE(
		table.InlayChats.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
