package user

import (
	"fmt"
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
	users, err := m.Db.DealershipUsers.GetAll()
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

	user, found, err := m.Db.DealershipUsers.GetByUUID(uuid)
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

// canManageDealershipUser enforces tenant scope for mutations on an existing
// dealership user. Internal users (already gated by permission) may act across
// dealerships; dealership users are confined to their own dealership.
func (m *UserModule) canManageDealershipUser(r *http.Request, target *data.DealershipUser) bool {
	requester := m.ContextGetUser(r)
	if requester.IsInternal() {
		return true
	}

	id := requester.GetDealershipID()
	return id != nil && *id == target.DealershipID
}

func (m *UserModule) HandleCreateDealershipUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name         string                  `json:"name" validate:"required"`
		Email        string                  `json:"email" validate:"required,email"`
		Avatar       string                  `json:"avatar" validate:"required,url"`
		Role         data.DealershipUserRole `json:"role" validate:"required"`
		DealershipID int                     `json:"dealership_id"`
		IsActive     bool                    `json:"is_active"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	// Resolve the target dealership. Dealership admins may only create users in
	// their own dealership (never trust a client-supplied dealership_id). Internal
	// admins act on behalf of a specific dealership supplied in the request.
	requester := m.ContextGetUser(r)
	var dealershipID int
	if requester.IsDealership() {
		id := requester.GetDealershipID()
		if id == nil {
			m.WriteError(w, r, m.Err.Forbidden, nil)
			return
		}
		dealershipID = *id
	} else {
		if body.DealershipID <= 0 {
			m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("dealership_id is required"))
			return
		}
		_, found, err := m.Db.Dealerships.GetByID(body.DealershipID)
		if err != nil {
			m.WriteError(w, r, m.Err.ServerError, err)
			return
		}
		if !found {
			m.WriteError(w, r, m.Err.BadRequest, fmt.Errorf("dealership %d not found", body.DealershipID))
			return
		}
		dealershipID = body.DealershipID
	}

	user := data.DealershipUser{
		Name:         body.Name,
		Email:        body.Email,
		Avatar:       body.Avatar,
		DealershipID: dealershipID,
		Role:         body.Role,
		IsActive:     body.IsActive,
	}

	err = m.Db.DealershipUsers.Insert(&user)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, user)
}

func (m *UserModule) HandleUpdateDealershipUser(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user, found, err := m.Db.DealershipUsers.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	if !m.canManageDealershipUser(r, user) {
		m.WriteError(w, r, m.Err.Forbidden, nil)
		return
	}

	var body struct {
		Name   string                  `json:"name"`
		Email  string                  `json:"email"`
		Avatar string                  `json:"avatar"`
		Role   data.DealershipUserRole `json:"role"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if body.Name != "" {
		user.Name = body.Name
	}
	if body.Email != "" {
		user.Email = body.Email
	}
	if body.Avatar != "" {
		user.Avatar = body.Avatar
	}
	if body.Role != "" {
		user.Role = body.Role
	}

	err = m.Db.DealershipUsers.Update(user)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, user)
}

func (m *UserModule) HandleDeleteDealershipUser(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user, found, err := m.Db.DealershipUsers.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	if !m.canManageDealershipUser(r, user) {
		m.WriteError(w, r, m.Err.Forbidden, nil)
		return
	}

	user.IsActive = false
	err = m.Db.DealershipUsers.Update(user)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, user)
}

func (m *UserModule) HandleGetInternalUsers(w http.ResponseWriter, r *http.Request) {
	users, err := m.Db.InternalUsers.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, users)
}

func (m *UserModule) HandleGetInternalUserByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user, found, err := m.Db.InternalUsers.GetByUUID(uuid)
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

func (m *UserModule) HandleCreateInternalUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string                `json:"name" validate:"required"`
		Email    string                `json:"email" validate:"required,email"`
		Avatar   string                `json:"avatar" validate:"required,url"`
		Role     data.InternalUserRole `json:"role" validate:"required"`
		IsActive bool                  `json:"is_active"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user := data.InternalUser{
		Name:     body.Name,
		Email:    body.Email,
		Avatar:   body.Avatar,
		Role:     body.Role,
		IsActive: body.IsActive,
	}

	err = m.Db.InternalUsers.Insert(&user)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, user)
}

func (m *UserModule) HandleUpdateInternalUser(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user, found, err := m.Db.InternalUsers.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	var body struct {
		Name   string                `json:"name"`
		Email  string                `json:"email"`
		Avatar string                `json:"avatar"`
		Role   data.InternalUserRole `json:"role"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	if body.Name != "" {
		user.Name = body.Name
	}
	if body.Email != "" {
		user.Email = body.Email
	}
	if body.Avatar != "" {
		user.Avatar = body.Avatar
	}
	if body.Role != "" {
		user.Role = body.Role
	}

	err = m.Db.InternalUsers.Update(user)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, user)
}

func (m *UserModule) HandleDeleteInternalUser(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	user, found, err := m.Db.InternalUsers.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	user.IsActive = false
	err = m.Db.InternalUsers.Update(user)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, user)
}
