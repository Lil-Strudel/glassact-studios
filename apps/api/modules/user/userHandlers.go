package user

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type UserModule struct {
	*app.Application
}

func NewUserModule(app *app.Application) *UserModule {
	return &UserModule{
		app,
	}
}

func (m *UserModule) HandleGetUserSelf(w http.ResponseWriter, r *http.Request) {
	user := m.ContextGetUser(r)

	m.WriteJSON(w, r, http.StatusOK, user)
}

func (m *UserModule) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := m.Db.Users.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, users)
}

func (m *UserModule) HandleGetUserByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user, found, err := m.Db.Users.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, user)
}

func (m *UserModule) HandlePostUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name         string `json:"name" validate:"required"`
		Email        string `json:"email" validate:"required,email"`
		Avatar       string `json:"avatar" validate:"required,url"`
		DealershipID int    `json:"dealership_id" validate:"required"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user := data.User{
		Name:         body.Name,
		Email:        body.Email,
		Avatar:       body.Avatar,
		DealershipID: body.DealershipID,
	}

	err = m.Db.Users.Insert(&user)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, user)
}
