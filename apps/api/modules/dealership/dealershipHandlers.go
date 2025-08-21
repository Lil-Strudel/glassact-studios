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

func (dm DealershipModule) HandleGetDealerships(w http.ResponseWriter, r *http.Request) {
	dealerships, err := dm.Db.Dealerships.GetAll()
	if err != nil {
		dm.WriteError(w, r, dm.Err.ServerError, err)
		return
	}

	dm.WriteJSON(w, r, http.StatusOK, dealerships)
}

func (dm DealershipModule) HandleGetDealershipByUUID(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := dm.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		dm.WriteError(w, r, dm.Err.BadRequest, err)
		return
	}

	dealership, found, err := dm.Db.Dealerships.GetByUUID(uuid)
	if err != nil {
		dm.WriteError(w, r, dm.Err.ServerError, err)
		return
	}

	if !found {
		dm.WriteError(w, r, dm.Err.RecordNotFound, nil)
		return
	}

	dm.WriteJSON(w, r, http.StatusOK, dealership)
}

func (dm DealershipModule) HandlePostDealership(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name    string `json:"name" validate:"required"`
		Address struct {
			Street     string  `json:"street" validate:"required"`
			StreetExt  *string `json:"street_ext"`
			City       string  `json:"city" validate:"required"`
			State      string  `json:"state" validate:"required"`
			PostalCode string  `json:"postal_code" validate:"required"`
			Country    string  `json:"country" validate:"required,iso3166_1_alpha2"`
			Latitude   float64 `json:"latitude" validate:"min=-90,max=90"`
			Longitude  float64 `json:"longitude" validate:"min=-180,max=180"`
		} `json:"address" validate:"required"`
	}

	err := dm.ReadJSONBody(w, r, &body)
	if err != nil {
		dm.WriteError(w, r, dm.Err.BadRequest, err)
		return
	}

	dealership := data.Dealership{
		Name:    body.Name,
		Address: data.Address(body.Address),
	}

	err = dm.Db.Dealerships.Insert(&dealership)
	if err != nil {
		dm.WriteError(w, r, dm.Err.ServerError, err)
		return
	}

	dm.WriteJSON(w, r, http.StatusOK, dealership)
}
