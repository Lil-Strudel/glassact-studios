package data

import (
	"context"
	"crypto/sha256"
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

type DealershipUserRole string

type dealershipUserRoles struct {
	Viewer    DealershipUserRole
	Submitter DealershipUserRole
	Approver  DealershipUserRole
	Admin     DealershipUserRole
}

var DealershipUserRoles = dealershipUserRoles{
	Viewer:    DealershipUserRole("viewer"),
	Submitter: DealershipUserRole("submitter"),
	Approver:  DealershipUserRole("approver"),
	Admin:     DealershipUserRole("admin"),
}

type DealershipUser struct {
	StandardTable
	DealershipID int                `json:"dealership_id"`
	Name         string             `json:"name"`
	Email        string             `json:"email"`
	Avatar       string             `json:"avatar"`
	Role         DealershipUserRole `json:"role"`
	IsActive     bool               `json:"is_active"`
}

type DealershipUserModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func dealershipUserFromGen(genUser model.DealershipUsers) *DealershipUser {
	user := DealershipUser{
		StandardTable: StandardTable{
			ID:        int(genUser.ID),
			UUID:      genUser.UUID.String(),
			CreatedAt: genUser.CreatedAt,
			UpdatedAt: genUser.UpdatedAt,
			Version:   int(genUser.Version),
		},
		DealershipID: int(genUser.DealershipID),
		Name:         genUser.Name,
		Email:        genUser.Email,
		Avatar:       genUser.Avatar,
		Role:         DealershipUserRole(genUser.Role),
		IsActive:     genUser.IsActive,
	}

	return &user
}

func dealershipUserToGen(u *DealershipUser) (*model.DealershipUsers, error) {
	var userUUID uuid.UUID
	var err error

	if u.UUID != "" {
		userUUID, err = uuid.Parse(u.UUID)
		if err != nil {
			return nil, err
		}
	}

	genUser := model.DealershipUsers{
		ID:           int32(u.ID),
		UUID:         userUUID,
		DealershipID: int32(u.DealershipID),
		Name:         u.Name,
		Email:        u.Email,
		Avatar:       u.Avatar,
		Role:         string(u.Role),
		IsActive:     u.IsActive,
		UpdatedAt:    u.UpdatedAt,
		CreatedAt:    u.CreatedAt,
		Version:      int32(u.Version),
	}

	return &genUser, nil
}

func (m DealershipUserModel) Insert(user *DealershipUser) error {
	genUser, err := dealershipUserToGen(user)
	if err != nil {
		return err
	}

	query := table.DealershipUsers.INSERT(
		table.DealershipUsers.DealershipID,
		table.DealershipUsers.Name,
		table.DealershipUsers.Email,
		table.DealershipUsers.Avatar,
		table.DealershipUsers.Role,
		table.DealershipUsers.IsActive,
	).MODEL(
		genUser,
	).RETURNING(
		table.DealershipUsers.ID,
		table.DealershipUsers.UUID,
		table.DealershipUsers.UpdatedAt,
		table.DealershipUsers.CreatedAt,
		table.DealershipUsers.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipUsers
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	user.ID = int(dest.ID)
	user.UUID = dest.UUID.String()
	user.UpdatedAt = dest.UpdatedAt
	user.CreatedAt = dest.CreatedAt
	user.Version = int(dest.Version)

	return nil
}

func (m DealershipUserModel) GetByID(id int) (*DealershipUser, bool, error) {
	query := postgres.SELECT(
		table.DealershipUsers.AllColumns,
	).FROM(
		table.DealershipUsers,
	).WHERE(
		table.DealershipUsers.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipUsers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return dealershipUserFromGen(dest), true, nil
}

func (m DealershipUserModel) GetByUUID(uuidStr string) (*DealershipUser, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.DealershipUsers.AllColumns,
	).FROM(
		table.DealershipUsers,
	).WHERE(
		table.DealershipUsers.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipUsers
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return dealershipUserFromGen(dest), true, nil
}

func (m DealershipUserModel) GetByEmail(email string) (*DealershipUser, bool, error) {
	query := postgres.SELECT(
		table.DealershipUsers.AllColumns,
	).FROM(
		table.DealershipUsers,
	).WHERE(
		table.DealershipUsers.Email.EQ(Citext(email)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipUsers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return dealershipUserFromGen(dest), true, nil
}

func (m DealershipUserModel) GetForToken(tokenScope, tokenPlaintext string) (*DealershipUser, bool, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := postgres.SELECT(
		table.DealershipUsers.AllColumns,
	).FROM(
		table.DealershipUsers.INNER_JOIN(table.DealershipTokens, table.DealershipTokens.DealershipUserID.EQ(table.DealershipUsers.ID)),
	).WHERE(
		postgres.AND(
			table.DealershipTokens.Hash.EQ(postgres.Bytea(tokenHash[:])),
			table.DealershipTokens.Scope.EQ(postgres.String(tokenScope)),
			table.DealershipTokens.Expiry.GT(postgres.TimestampzExp(Now())),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipUsers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return dealershipUserFromGen(dest), true, nil
}

func (m DealershipUserModel) GetAll() ([]*DealershipUser, error) {
	query := postgres.SELECT(
		table.DealershipUsers.AllColumns,
	).FROM(
		table.DealershipUsers,
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

func (m DealershipUserModel) Update(user *DealershipUser) error {
	genUser, err := dealershipUserToGen(user)
	if err != nil {
		return err
	}

	query := table.DealershipUsers.UPDATE(
		table.DealershipUsers.Name,
		table.DealershipUsers.Email,
		table.DealershipUsers.Avatar,
		table.DealershipUsers.Role,
		table.DealershipUsers.IsActive,
		table.DealershipUsers.Version,
	).MODEL(
		genUser,
	).WHERE(
		postgres.AND(
			table.DealershipUsers.ID.EQ(postgres.Int(int64(user.ID))),
			table.DealershipUsers.Version.EQ(postgres.Int(int64(user.Version))),
		),
	).RETURNING(
		table.DealershipUsers.UpdatedAt,
		table.DealershipUsers.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.DealershipUsers
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	user.UpdatedAt = dest.UpdatedAt
	user.Version = int(dest.Version)

	return nil
}

func (m DealershipUserModel) Delete(id int) error {
	query := table.DealershipUsers.DELETE().WHERE(
		table.DealershipUsers.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
