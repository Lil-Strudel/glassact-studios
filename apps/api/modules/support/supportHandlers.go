package support

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
)

type SupportModule struct {
	*app.Application
}

func NewSupportModule(app *app.Application) *SupportModule {
	return &SupportModule{app}
}

// HandleGetArticles returns published knowledge-base articles, grouped/ordered by
// category then sort_order, for every authenticated user.
func (m *SupportModule) HandleGetArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := m.Db.SupportArticles.GetAllPublished()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, articles)
}

func (m *SupportModule) HandleGetArticle(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	article, found, err := m.Db.SupportArticles.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, article)
}

func (m *SupportModule) HandlePostArticle(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Category    string  `json:"category" validate:"required,oneof=installation ordering pricing contact general"`
		Title       string  `json:"title" validate:"required,min=1,max=255"`
		Body        string  `json:"body"`
		YoutubeURL  *string `json:"youtube_url" validate:"omitempty,url"`
		SortOrder   int     `json:"sort_order"`
		IsPublished bool    `json:"is_published"`
	}

	err := m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	article := &data.SupportArticle{
		Category:    data.SupportCategory(body.Category),
		Title:       body.Title,
		Body:        body.Body,
		YoutubeURL:  body.YoutubeURL,
		SortOrder:   body.SortOrder,
		IsPublished: body.IsPublished,
	}

	err = m.Db.SupportArticles.Insert(article)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusCreated, article)
}

func (m *SupportModule) HandlePatchArticle(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	var body struct {
		Category    *string `json:"category" validate:"omitempty,oneof=installation ordering pricing contact general"`
		Title       *string `json:"title" validate:"omitempty,min=1,max=255"`
		Body        *string `json:"body"`
		YoutubeURL  *string `json:"youtube_url" validate:"omitempty,url"`
		SortOrder   *int    `json:"sort_order"`
		IsPublished *bool   `json:"is_published"`
	}

	err = m.ReadJSONBody(w, r, &body)
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	article, found, err := m.Db.SupportArticles.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	if body.Category != nil {
		article.Category = data.SupportCategory(*body.Category)
	}
	if body.Title != nil {
		article.Title = *body.Title
	}
	if body.Body != nil {
		article.Body = *body.Body
	}
	if body.YoutubeURL != nil {
		article.YoutubeURL = body.YoutubeURL
	}
	if body.SortOrder != nil {
		article.SortOrder = *body.SortOrder
	}
	if body.IsPublished != nil {
		article.IsPublished = *body.IsPublished
	}

	err = m.Db.SupportArticles.Update(article)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, article)
}

func (m *SupportModule) HandleDeleteArticle(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	err := m.Validate.Var(uuid, "required,uuid4")
	if err != nil {
		m.WriteError(w, r, m.Err.BadRequest, err)
		return
	}

	article, found, err := m.Db.SupportArticles.GetByUUID(uuid)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}
	if !found {
		m.WriteError(w, r, m.Err.RecordNotFound, nil)
		return
	}

	err = m.Db.SupportArticles.Delete(article.ID)
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// HandleGetPriceGroups returns the active price groups so the Support page's
// pricing section can render live pricing to any authenticated user (the admin
// price-group management routes stay gated behind manage_price_groups).
func (m *SupportModule) HandleGetPriceGroups(w http.ResponseWriter, r *http.Request) {
	priceGroups, err := m.Db.PriceGroups.GetAllActive()
	if err != nil {
		m.WriteError(w, r, m.Err.ServerError, err)
		return
	}

	m.WriteJSON(w, r, http.StatusOK, priceGroups)
}
