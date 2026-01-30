package data

import (
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Models struct {
	Accounts    AccountModel
	Dealerships DealershipModel
	Inlays      InlayModel
	InlayChats  InlayChatModel
	InlayProofs InlayProofModel
	Projects    ProjectModel
	Tokens      TokenModel
	Users       UserModel
	Pool        *pgxpool.Pool
}

func NewModels(db *pgxpool.Pool, stdb *sql.DB) Models {
	return Models{
		Accounts:    AccountModel{DB: db, STDB: stdb},
		Dealerships: DealershipModel{DB: db, STDB: stdb},
		Inlays:      InlayModel{DB: db, STDB: stdb},
		InlayChats:  InlayChatModel{DB: db},
		InlayProofs: InlayProofModel{DB: db},
		Projects:    ProjectModel{DB: db, STDB: stdb},
		Tokens:      TokenModel{DB: db, STDB: stdb},
		Users:       UserModel{DB: db, STDB: stdb},
		Pool:        db,
	}
}
