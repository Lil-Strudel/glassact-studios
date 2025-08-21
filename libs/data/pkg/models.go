package data

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type Models struct {
	Accounts    AccountModel
	Dealerships DealershipModel
	Tokens      TokenModel
	Users       UserModel
	Pool        *pgxpool.Pool
}

func NewModels(db *pgxpool.Pool) Models {
	return Models{
		Accounts:    AccountModel{DB: db},
		Dealerships: DealershipModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Users:       UserModel{DB: db},
		Pool:        db,
	}
}
