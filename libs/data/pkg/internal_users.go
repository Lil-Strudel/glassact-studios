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

type InternalUserRole string

type internalUserRoles struct {
	Designer   InternalUserRole
	Production InternalUserRole
	Billing    InternalUserRole
	Admin      InternalUserRole
}

var InternalUserRoles = internalUserRoles{
	Designer:   InternalUserRole("designer"),
	Production: InternalUserRole("production"),
	Billing:    InternalUserRole("billing"),
	Admin:      InternalUserRole("admin"),
}

type InternalUser struct {
	StandardTable
	Name     string           `json:"name"`
	Email    string           `json:"email"`
	Avatar   string           `json:"avatar"`
	Role     InternalUserRole `json:"role"`
	IsActive bool             `json:"is_active"`
}

type InternalUserModel struct {
	DB   *pgxpool.Pool
	STDB *sql.DB
}

func internalUserFromGen(genUser model.InternalUsers) *InternalUser {
	user := InternalUser{
		StandardTable: StandardTable{
			ID:        int(genUser.ID),
			UUID:      genUser.UUID.String(),
			CreatedAt: genUser.CreatedAt,
			UpdatedAt: genUser.UpdatedAt,
			Version:   int(genUser.Version),
		},
		Name:     genUser.Name,
		Email:    genUser.Email,
		Avatar:   genUser.Avatar,
		Role:     InternalUserRole(genUser.Role),
		IsActive: genUser.IsActive,
	}

	return &user
}

func internalUserToGen(u *InternalUser) (*model.InternalUsers, error) {
	var userUUID uuid.UUID
	var err error

	if u.UUID != "" {
		userUUID, err = uuid.Parse(u.UUID)
		if err != nil {
			return nil, err
		}
	}

	genUser := model.InternalUsers{
		ID:        int32(u.ID),
		UUID:      userUUID,
		Name:      u.Name,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Role:      string(u.Role),
		IsActive:  u.IsActive,
		UpdatedAt: u.UpdatedAt,
		CreatedAt: u.CreatedAt,
		Version:   int32(u.Version),
	}

	return &genUser, nil
}

func (m InternalUserModel) Insert(user *InternalUser) error {
	genUser, err := internalUserToGen(user)
	if err != nil {
		return err
	}

	query := table.InternalUsers.INSERT(
		table.InternalUsers.Name,
		table.InternalUsers.Email,
		table.InternalUsers.Avatar,
		table.InternalUsers.Role,
		table.InternalUsers.IsActive,
	).MODEL(
		genUser,
	).RETURNING(
		table.InternalUsers.ID,
		table.InternalUsers.UUID,
		table.InternalUsers.UpdatedAt,
		table.InternalUsers.CreatedAt,
		table.InternalUsers.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalUsers
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

func (m InternalUserModel) GetByID(id int) (*InternalUser, bool, error) {
	query := postgres.SELECT(
		table.InternalUsers.AllColumns,
	).FROM(
		table.InternalUsers,
	).WHERE(
		table.InternalUsers.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalUsers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return internalUserFromGen(dest), true, nil
}

func (m InternalUserModel) GetByUUID(uuidStr string) (*InternalUser, bool, error) {
	parsedUUID, err := uuid.Parse(uuidStr)
	if err != nil {
		return nil, false, err
	}

	query := postgres.SELECT(
		table.InternalUsers.AllColumns,
	).FROM(
		table.InternalUsers,
	).WHERE(
		table.InternalUsers.UUID.EQ(postgres.UUID(parsedUUID)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalUsers
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return internalUserFromGen(dest), true, nil
}

func (m InternalUserModel) GetByEmail(email string) (*InternalUser, bool, error) {
	query := postgres.SELECT(
		table.InternalUsers.AllColumns,
	).FROM(
		table.InternalUsers,
	).WHERE(
		table.InternalUsers.Email.EQ(Citext(email)),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalUsers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return internalUserFromGen(dest), true, nil
}

func (m InternalUserModel) GetForToken(tokenScope, tokenPlaintext string) (*InternalUser, bool, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := postgres.SELECT(
		table.InternalUsers.AllColumns,
	).FROM(
		table.InternalUsers.INNER_JOIN(table.InternalTokens, table.InternalTokens.InternalUserID.EQ(table.InternalUsers.ID)),
	).WHERE(
		postgres.AND(
			table.InternalTokens.Hash.EQ(postgres.Bytea(tokenHash[:])),
			table.InternalTokens.Scope.EQ(postgres.String(tokenScope)),
			table.InternalTokens.Expiry.GT(postgres.TimestampzExp(Now())),
		),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalUsers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		switch {
		case errors.Is(err, qrm.ErrNoRows):
			return nil, false, nil
		default:
			return nil, false, err
		}
	}

	return internalUserFromGen(dest), true, nil
}

func (m InternalUserModel) GetAll() ([]*InternalUser, error) {
	query := postgres.SELECT(
		table.InternalUsers.AllColumns,
	).FROM(
		table.InternalUsers,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest []model.InternalUsers
	err := query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return nil, err
	}

	users := make([]*InternalUser, len(dest))
	for i, d := range dest {
		users[i] = internalUserFromGen(d)
	}

	return users, nil
}

func (m InternalUserModel) Update(user *InternalUser) error {
	genUser, err := internalUserToGen(user)
	if err != nil {
		return err
	}

	query := table.InternalUsers.UPDATE(
		table.InternalUsers.Name,
		table.InternalUsers.Email,
		table.InternalUsers.Avatar,
		table.InternalUsers.Role,
		table.InternalUsers.IsActive,
		table.InternalUsers.Version,
	).MODEL(
		genUser,
	).WHERE(
		postgres.AND(
			table.InternalUsers.ID.EQ(postgres.Int(int64(user.ID))),
			table.InternalUsers.Version.EQ(postgres.Int(int64(user.Version))),
		),
	).RETURNING(
		table.InternalUsers.UpdatedAt,
		table.InternalUsers.Version,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var dest model.InternalUsers
	err = query.QueryContext(ctx, m.STDB, &dest)
	if err != nil {
		return err
	}

	user.UpdatedAt = dest.UpdatedAt
	user.Version = int(dest.Version)

	return nil
}

func (m InternalUserModel) Delete(id int) error {
	query := table.InternalUsers.DELETE().WHERE(
		table.InternalUsers.ID.EQ(postgres.Int(int64(id))),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := query.ExecContext(ctx, m.STDB)
	if err != nil {
		return err
	}

	return nil
}

func (u *InternalUser) GetID() int {
	return u.ID
}

func (u *InternalUser) GetUUID() string {
	return u.UUID
}

func (u *InternalUser) GetEmail() string {
	return u.Email
}

func (u *InternalUser) GetName() string {
	return u.Name
}

func (u *InternalUser) GetAvatar() string {
	return u.Avatar
}

func (u *InternalUser) GetRole() string {
	return string(u.Role)
}

func (u *InternalUser) GetIsActive() bool {
	return u.IsActive
}

func (u *InternalUser) IsInternal() bool {
	return true
}

func (u *InternalUser) IsDealership() bool {
	return false
}

func (u *InternalUser) GetDealershipID() *int {
	return nil
}

func (u *InternalUser) Can(action string) bool {
	switch action {
	case ActionCreateProof:
		return u.Role == InternalUserRoles.Designer ||
			u.Role == InternalUserRoles.Admin
	case ActionManageKanban:
		return u.Role == InternalUserRoles.Production ||
			u.Role == InternalUserRoles.Admin
	case ActionCreateBlocker:
		return u.Role == InternalUserRoles.Production ||
			u.Role == InternalUserRoles.Admin
	case ActionCreateInvoice:
		return u.Role == InternalUserRoles.Billing ||
			u.Role == InternalUserRoles.Admin
	case ActionManageInternalUsers:
		return u.Role == InternalUserRoles.Admin
	case ActionViewAll:
		return u.Role == InternalUserRoles.Admin
	default:
		return false
	}
}
