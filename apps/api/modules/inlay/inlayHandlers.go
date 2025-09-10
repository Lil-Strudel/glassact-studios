package inlay

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
)

type InlayModule struct {
	*app.Application
}

func NewInlayModule(app *app.Application) *InlayModule {
	return &InlayModule{
		app,
	}
}

func (m InlayModule) HandleGetInlays(w http.ResponseWriter, r *http.Request) {
	// inlays, err := m.Db.Inlays.GetAll()
	// if err != nil {
	// 	m.WriteError(w, r, m.Err.ServerError, err)
	// 	return
	// }
	//
	// m.WriteJSON(w, r, http.StatusOK, inlays)
	m.WriteJSON(w, r, http.StatusOK, nil)
}

func (m InlayModule) HandleGetInlayByUUID(w http.ResponseWriter, r *http.Request) {
	// uuid := r.PathValue("uuid")
	//
	// err := m.Validate.Var(uuid, "required,uuid4")
	// if err != nil {
	// 	m.WriteError(w, r, m.Err.BadRequest, err)
	// 	return
	// }
	//
	// inlay, found, err := m.Db.Inlays.GetByUUID(uuid)
	// if err != nil {
	// 	m.WriteError(w, r, m.Err.ServerError, err)
	// 	return
	// }
	//
	// if !found {
	// 	m.WriteError(w, r, m.Err.RecordNotFound, nil)
	// 	return
	// }
	//
	// m.WriteJSON(w, r, http.StatusOK, inlay)
	m.WriteJSON(w, r, http.StatusOK, nil)
}

func (m InlayModule) HandlePostInlay(w http.ResponseWriter, r *http.Request) {
	// var body struct {
	// 	ProjectID   int            `json:"project_id" validate:"required"`
	// 	Name        string         `json:"name" validate:"required"`
	// 	PreviewURL  string         `json:"preview_url" validate:"required"`
	// 	PriceGroup  int            `json:"price_group" validate:"required"`
	// 	Type        data.InlayType `json:"type" validate:"required"`
	// 	CatalogInfo *struct {
	// 		CatalogItemID int `json:"catalog_item_id" validate:"required"`
	// 	} `json:"catalog_info,omitempty" validate:"required_if=Type catalog"`
	// 	CustomInfo *struct {
	// 		Description string  `json:"description" validate:"required"`
	// 		Width       float64 `json:"width" validate:"required"`
	// 		Height      float64 `json:"height" validate:"required"`
	// 	} `json:"custom_info,omitempty" validate:"required_if=Type custom"`
	// }
	//
	// err := m.ReadJSONBody(w, r, &body)
	// if err != nil {
	// 	m.WriteError(w, r, m.Err.BadRequest, err)
	// 	return
	// }
	//
	// inlay := data.Inlay{
	// 	ProjectID:  body.ProjectID,
	// 	Name:       body.Name,
	// 	PreviewURL: body.PreviewURL,
	// 	PriceGroup: body.PriceGroup,
	// 	Type:       body.Type,
	// 	CatalogInfo: &data.InlayCatalogInfo{
	// 		CatalogItemID: body.CatalogInfo.CatalogItemID,
	// 	},
	// 	CustomInfo: &data.InlayCustomInfo{
	// 		Description: body.CustomInfo.Description,
	// 		Width:       body.CustomInfo.Width,
	// 		Height:      body.CustomInfo.Height,
	// 	},
	// }
	//
	// err = m.Db.Inlays.Insert(&inlay)
	// if err != nil {
	// 	m.WriteError(w, r, m.Err.ServerError, err)
	// 	return
	// }
	//
	// m.WriteJSON(w, r, http.StatusOK, inlay)
	m.WriteJSON(w, r, http.StatusOK, nil)
}
