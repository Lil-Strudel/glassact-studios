package data

import "github.com/jackc/pgx/v5/pgxpool"

type InlayProof struct {
	StandardTable
	InlayID int `json:"inlay_id"`
}

type InlayProofModel struct {
	DB *pgxpool.Pool
}
