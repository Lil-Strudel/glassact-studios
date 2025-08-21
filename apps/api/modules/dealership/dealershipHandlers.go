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
	dealership := data.Dealership{}
	dm.WriteJSON(w, r, http.StatusOK, []data.Dealership{dealership})
}
