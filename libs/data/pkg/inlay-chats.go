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

type SenderType string

type senderTypes struct {
	GlassAct SenderType
	Customer SenderType
}

var SenderTypes = senderTypes{
	GlassAct: SenderType("glassact"),
	Customer: SenderType("customer"),
}

type InlayChat struct {
	StandardTable
	InlayID    int        `json:"inlay_id"`
	UserID     int        `json:"user_id"`
	SenderType SenderType `json:"sender_type"`
	Message    string     `json:"message"`
}

type InlayChatModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func inlayChatFromGen(genChat model.InlayChats) *InlayChat {
	chat := InlayChat{
		StandardTable: StandardTable{
			ID:        int(genChat.ID),
			UUID:      genChat.UUID.String(),
			CreatedAt: genChat.CreatedAt,
			UpdatedAt: genChat.UpdatedAt,
			Version:   int(genChat.Version),
		},
		InlayID:    int(genChat.InlayID),
		UserID:     int(genChat.UserID),
		SenderType: SenderType(genChat.SenderType),
		Message:    genChat.Message,
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

	genChat := model.InlayChats{
		ID:         int32(ic.ID),
		UUID:       chatUUID,
		InlayID:    int32(ic.InlayID),
		UserID:     int32(ic.UserID),
		SenderType: string(ic.SenderType),
		Message:    ic.Message,
		UpdatedAt:  ic.UpdatedAt,
		CreatedAt:  ic.CreatedAt,
		Version:    int32(ic.Version),
	}

	return &genChat, nil
}

func (m InlayChatModel) Insert(inlayChat *InlayChat) error {
	genChat, err := inlayChatToGen(inlayChat)
	if err != nil {
		return err
	}

	query := table.InlayChats.INSERT(
		table.InlayChats.InlayID,
		table.InlayChats.UserID,
		table.InlayChats.SenderType,
		table.InlayChats.Message,
	).MODEL(
		genChat,
	).RETURNING(
		table.InlayChats.ID,
		table.InlayChats.UUID,
		table.InlayChats.UpdatedAt,
		table.InlayChats.CreatedAt,
		table.InlayChats.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InlayChats
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	inlayChat.ID = int(dest.ID)
	inlayChat.UUID = dest.UUID.String()
	inlayChat.UpdatedAt = dest.UpdatedAt
	inlayChat.CreatedAt = dest.CreatedAt
	inlayChat.Version = int(dest.Version)

	return nil
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

	inlayChats := make([]*InlayChat, len(dest))
	for i, d := range dest {
		inlayChats[i] = inlayChatFromGen(d)
	}

	return inlayChats, nil
}

func (m InlayChatModel) GetAllByInlayID(inlayID int) ([]*InlayChat, error) {
	query := postgres.SELECT(
		table.InlayChats.AllColumns,
	).FROM(
		table.InlayChats,
	).WHERE(
		table.InlayChats.InlayID.EQ(postgres.Int(int64(inlayID))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InlayChats
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	inlayChats := make([]*InlayChat, len(dest))
	for i, d := range dest {
		inlayChats[i] = inlayChatFromGen(d)
	}

	return inlayChats, nil
}

func (m InlayChatModel) Update(inlayChat *InlayChat) error {
	genChat, err := inlayChatToGen(inlayChat)
	if err != nil {
		return err
	}

	query := table.InlayChats.UPDATE(
		table.InlayChats.InlayID,
		table.InlayChats.UserID,
		table.InlayChats.SenderType,
		table.InlayChats.Message,
		table.InlayChats.Version,
	).MODEL(
		genChat,
	).WHERE(
		postgres.AND(
			table.InlayChats.ID.EQ(postgres.Int(int64(inlayChat.ID))),
			table.InlayChats.Version.EQ(postgres.Int(int64(inlayChat.Version))),
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

	inlayChat.UpdatedAt = dest.UpdatedAt
	inlayChat.Version = int(dest.Version)

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
