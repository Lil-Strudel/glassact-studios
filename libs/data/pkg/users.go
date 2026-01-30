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

type UserRole string

type userRoles struct {
	Admin UserRole
	User  UserRole
}

var UserRoles = userRoles{
	Admin: UserRole("admin"),
	User:  UserRole("user"),
}

type User struct {
	StandardTable
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	Avatar       string   `json:"avatar"`
	DealershipID int      `json:"dealership_id"`
	Role         UserRole `json:"role"`
}

type UserModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func userFromGen(genUser model.Users) *User {
	user := User{
		StandardTable: StandardTable{
			ID:        int(genUser.ID),
			UUID:      genUser.UUID.String(),
			CreatedAt: genUser.CreatedAt,
			UpdatedAt: genUser.UpdatedAt,
			Version:   int(genUser.Version),
		},
		Name:         genUser.Name,
		Email:        genUser.Email,
		Avatar:       genUser.Avatar,
		DealershipID: int(genUser.DealershipID),
		Role:         UserRole(genUser.Role),
	}

	return &user
}

func userToGen(u *User) (*model.Users, error) {
	var userUUID uuid.UUID
	var err error

	if u.UUID != "" {
		userUUID, err = uuid.Parse(u.UUID)
		if err != nil {
			return nil, err
		}
	}

	genUser := model.Users{
		ID:           int32(u.ID),
		UUID:         userUUID,
		Name:         u.Name,
		Email:        u.Email,
		Avatar:       u.Avatar,
		DealershipID: int32(u.DealershipID),
		Role:         string(u.Role),
		UpdatedAt:    u.UpdatedAt,
		CreatedAt:    u.CreatedAt,
		Version:      int32(u.Version),
	}

	return &genUser, nil
}

func (m UserModel) Insert(user *User) error {
	genUser, err := userToGen(user)
	if err != nil {
		return err
	}

	query := table.Users.INSERT(
		table.Users.Name,
		table.Users.Email,
		table.Users.Avatar,
		table.Users.DealershipID,
		table.Users.Role,
	).MODEL(
		genUser,
	).RETURNING(
		table.Users.ID,
		table.Users.UUID,
		table.Users.UpdatedAt,
		table.Users.CreatedAt,
		table.Users.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Users
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

func (m UserModel) GetByID(id int) (*User, bool, error) {
	query := postgres.SELECT(
		table.Users.AllColumns,
	).FROM(
		table.Users,
	).WHERE(
		table.Users.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Users
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return userFromGen(dest), true, nil
}

func (m UserModel) GetByUUID(uuidStr string) (*User, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.Users.AllColumns,
	).FROM(
		table.Users,
	).WHERE(
		table.Users.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Users
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return userFromGen(dest), true, nil
}

func (m UserModel) GetByEmail(email string) (*User, bool, error) {
	query := postgres.SELECT(
		table.Users.AllColumns,
	).FROM(
		table.Users,
	).WHERE(
		table.Users.Email.EQ(postgres.String(email)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Users
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return userFromGen(dest), true, nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, bool, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := postgres.SELECT(
		table.Users.AllColumns,
	).FROM(
		table.Users.INNER_JOIN(table.Tokens, table.Tokens.UserID.EQ(table.Users.ID)),
	).WHERE(
		postgres.AND(
			table.Tokens.Hash.EQ(postgres.Bytea(tokenHash[:])),
			table.Tokens.Scope.EQ(postgres.String(tokenScope)),
			table.Tokens.Expiry.GT(postgres.TimestampzExp(postgres.Raw("now()"))),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Users
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return userFromGen(dest), true, nil
}

func (m UserModel) GetAll() ([]*User, error) {
	query := postgres.SELECT(
		table.Users.AllColumns,
	).FROM(
		table.Users,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.Users
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	users := make([]*User, len(dest))
	for i, d := range dest {
		users[i] = userFromGen(d)
	}

	return users, nil
}

func (m UserModel) Update(user *User) error {
	genUser, err := userToGen(user)
	if err != nil {
		return err
	}

	query := table.Users.UPDATE(
		table.Users.Name,
		table.Users.Email,
		table.Users.Avatar,
		table.Users.Role,
		table.Users.Version,
	).MODEL(
		genUser,
	).WHERE(
		postgres.AND(
			table.Users.ID.EQ(postgres.Int(int64(user.ID))),
			table.Users.Version.EQ(postgres.Int(int64(user.Version))),
		),
	).RETURNING(
		table.Users.UpdatedAt,
		table.Users.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.Users
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	user.UpdatedAt = dest.UpdatedAt
	user.Version = int(dest.Version)

	return nil
}

func (m UserModel) Delete(id int) error {
	query := table.Users.DELETE().WHERE(
		table.Users.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}
