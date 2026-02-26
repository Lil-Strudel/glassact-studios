package dealership

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type DealershipModule struct {
	*app.Application
}

func NewDealershipModule(app *app.Application) *DealershipModule {
	return &DealershipModule{
		app,
	}
}

func (m DealershipModule) HandleGetDealerships(w http.ResponseWriter, r *http.Request) {
	dealerships, err := m.Db.Dealerships.GetAll()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, dealerships)
}

func (m DealershipModule) HandleGetDealershipByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	dealership, found, err := m.Db.Dealerships.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, dealership)
}

func (m DealershipModule) HandlePostDealership(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name    string `json:"name" validate:"required"`
		Address struct {
			Street     string  `json:"street" validate:"required"`
			StreetExt  string  `json:"street_ext"`
			City       string  `json:"city" validate:"required"`
			State      string  `json:"state" validate:"required"`
			PostalCode string  `json:"postal_code" validate:"required"`
			Country    string  `json:"country" validate:"required,iso3166_1_alpha2"`
			Latitude   float64 `json:"latitude" validate:"required,latitude"`
			Longitude  float64 `json:"longitude" validate:"required,longitude"`
		} `json:"address" validate:"required"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	dealership := data.Dealership{
		Name:    body.Name,
		Address: data.Address(body.Address),
	}

	err = m.Db.Dealerships.Insert(&dealership)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, dealership)
}
