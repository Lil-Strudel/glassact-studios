package data

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
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
	DB *pgxpool.Pool
}

func (m InlayChatModel) Insert(inlayChat *InlayChat) error {
	query := `
        INSERT INTO inlay_chats (inlay_id, user_id, sender_type, message) 
        VALUES ($1, $2, $3, $4)
        RETURNING id, uuid, created_at, updated_at, version`

	args := []any{
		inlayChat.InlayID,
		inlayChat.UserID,
		inlayChat.SenderType,
		inlayChat.Message,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(
		&inlayChat.ID,
		&inlayChat.UUID,
		&inlayChat.CreatedAt,
		&inlayChat.UpdatedAt,
		&inlayChat.Version,
	)
	if err != nil {
		return err
	}

	return nil
}

func (m InlayChatModel) GetByID(id int) (*InlayChat, bool, error) {
	query := `
        SELECT id, uuid, inlay_id, user_id, sender_type, message, created_at, updated_at, version
        FROM inlay_chats
        WHERE id = $1`

	var inlayChat InlayChat

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, id).Scan(
		&inlayChat.ID,
		&inlayChat.UUID,
		&inlayChat.InlayID,
		&inlayChat.UserID,
		&inlayChat.SenderType,
		&inlayChat.Message,
		&inlayChat.CreatedAt,
		&inlayChat.UpdatedAt,
		&inlayChat.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &inlayChat, true, nil
}

func (m InlayChatModel) GetByUUID(uuid string) (*InlayChat, bool, error) {
	query := `
        SELECT id, uuid, inlay_id, user_id, sender_type, message, created_at, updated_at, version
        FROM inlay_chats
        WHERE uuid = $1`

	var inlayChat InlayChat

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, uuid).Scan(
		&inlayChat.ID,
		&inlayChat.UUID,
		&inlayChat.InlayID,
		&inlayChat.UserID,
		&inlayChat.SenderType,
		&inlayChat.Message,
		&inlayChat.CreatedAt,
		&inlayChat.UpdatedAt,
		&inlayChat.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return &inlayChat, true, nil
}

func (m InlayChatModel) GetAll() ([]*InlayChat, error) {
	query := `
        SELECT id, uuid, inlay_id, user_id, sender_type, message, created_at, updated_at, version
		FROM inlay_chats
		WHERE id IS NOT NULL;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{}

	rows, err := m.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	inlayChats := []*InlayChat{}

	for rows.Next() {
		var inlayChat InlayChat

		err := rows.Scan(
			&inlayChat.ID,
			&inlayChat.UUID,
			&inlayChat.InlayID,
			&inlayChat.UserID,
			&inlayChat.SenderType,
			&inlayChat.Message,
			&inlayChat.CreatedAt,
			&inlayChat.UpdatedAt,
			&inlayChat.Version,
		)
		if err != nil {
			return nil, err
		}

		inlayChats = append(inlayChats, &inlayChat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return inlayChats, nil
}

func (m InlayChatModel) GetAllByInlayID(inlayID int) ([]*InlayChat, error) {
	query := `
        SELECT id, uuid, inlay_id, user_id, sender_type, message, created_at, updated_at, version
		FROM inlay_chats
		WHERE inlay_id = $1;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{
		inlayID,
	}

	rows, err := m.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	inlayChats := []*InlayChat{}

	for rows.Next() {
		var inlayChat InlayChat

		err := rows.Scan(
			&inlayChat.ID,
			&inlayChat.UUID,
			&inlayChat.InlayID,
			&inlayChat.UserID,
			&inlayChat.SenderType,
			&inlayChat.Message,
			&inlayChat.CreatedAt,
			&inlayChat.UpdatedAt,
			&inlayChat.Version,
		)
		if err != nil {
			return nil, err
		}

		inlayChats = append(inlayChats, &inlayChat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return inlayChats, nil
}

func (m InlayChatModel) Update(inlayChat *InlayChat) error {
	query := `
        UPDATE inlay_chats 
        SET inlay_id = $3, user_id = $4, sender_type = $5, message = $6, version = version + 1
        WHERE id = $1 AND version = $2
        RETURNING version`

	args := []any{
		inlayChat.ID,
		inlayChat.Version,
		inlayChat.InlayID,
		inlayChat.UserID,
		inlayChat.SenderType,
		inlayChat.Message,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, query, args...).Scan(&inlayChat.Version)
	if err != nil {
		return err
	}

	return nil
}

func (m InlayChatModel) Delete(id int) error {
	query := `
        DELETE FROM inlay_chats
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}
