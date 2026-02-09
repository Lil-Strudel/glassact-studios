package modules

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/auth"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/catalog"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/dealership"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/inlay"
	inlayChat "github.com/Lil-Strudel/glassact-studios/apps/api/modules/inlay-chat"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/pricegroup"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/project"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/upload"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/user"
	"github.com/justinas/alice"
)

func GetRoutes(app *app.Application) http.Handler {
	mux := http.NewServeMux()

	unprotected := alice.New()
	protected := alice.New(app.Authenticate)
	standard := alice.New(app.RecoverPanic, app.LogRequest)

	authModule := auth.NewAuthModule(app)
	mux.Handle("GET /api/auth/google", unprotected.ThenFunc(authModule.HandleGetGoogleAuth))
	mux.Handle("GET /api/auth/google/callback", unprotected.ThenFunc(authModule.HandleGetGoogleAuthCallback))

	mux.Handle("GET /api/auth/microsoft", unprotected.ThenFunc(authModule.HandleGetMicrosoftAuth))
	mux.Handle("GET /api/auth/microsoft/callback", unprotected.ThenFunc(authModule.HandleGetMicrosoftAuthCallback))

	mux.Handle("POST /api/auth/magic-link", unprotected.ThenFunc(authModule.HandlePostMagicLinkAuth))
	mux.Handle("GET /api/auth/magic-link/callback", unprotected.ThenFunc(authModule.HandleGetMagicLinkCallback))

	mux.Handle("POST /api/auth/token/access", unprotected.ThenFunc(authModule.HandlePostTokenAccess))
	mux.Handle("GET /api/auth/logout", unprotected.ThenFunc(authModule.HandleGetLogout))

	dealershipModule := dealership.NewDealershipModule(app)
	mux.Handle("GET /api/dealership", protected.ThenFunc(dealershipModule.HandleGetDealerships))
	mux.Handle("GET /api/dealership/{uuid}", protected.ThenFunc(dealershipModule.HandleGetDealershipByUUID))
	mux.Handle("POST /api/dealership", protected.ThenFunc(dealershipModule.HandlePostDealership))

	inlayModule := inlay.NewInlayModule(app)
	mux.Handle("GET /api/inlay", protected.ThenFunc(inlayModule.HandleGetInlays))
	mux.Handle("GET /api/inlay/{uuid}", protected.ThenFunc(inlayModule.HandleGetInlayByUUID))
	mux.Handle("POST /api/inlay", protected.ThenFunc(inlayModule.HandlePostInlay))

	inlayChatModule := inlayChat.NewInlayChatModule(app)
	mux.Handle("GET /api/inlay-chat", protected.ThenFunc(inlayChatModule.HandleGetInlayChats))
	mux.Handle("GET /api/inlay-chat/inlay/{uuid}", protected.ThenFunc(inlayChatModule.HandleGetInlayChatsByInlayUUID))
	mux.Handle("GET /api/inlay-chat/{uuid}", protected.ThenFunc(inlayChatModule.HandleGetInlayChatByUUID))
	mux.Handle("POST /api/inlay-chat", protected.ThenFunc(inlayChatModule.HandlePostInlayChat))

	projectModule := project.NewProjectModule(app)
	mux.Handle("GET /api/project", protected.ThenFunc(projectModule.HandleGetProjects))
	mux.Handle("GET /api/project/{uuid}", protected.ThenFunc(projectModule.HandleGetProjectByUUID))
	mux.Handle("POST /api/project", protected.ThenFunc(projectModule.HandlePostProject))
	mux.Handle("POST /api/project/with-inlays", protected.ThenFunc(projectModule.HandlePostProjectWithInlays))

	userModule := user.NewUserModule(app)
	mux.Handle("GET /api/user/self", protected.ThenFunc(userModule.HandleGetUserSelf))
	mux.Handle("GET /api/dealership-user", protected.ThenFunc(userModule.HandleGetUsers))
	mux.Handle("GET /api/dealership-user/{uuid}", protected.ThenFunc(userModule.HandleGetUserByUUID))
	mux.Handle("POST /api/dealership-user", protected.ThenFunc(userModule.HandleCreateDealershipUser))
	mux.Handle("PATCH /api/dealership-user/{uuid}", protected.ThenFunc(userModule.HandleUpdateDealershipUser))
	mux.Handle("DELETE /api/dealership-user/{uuid}", protected.ThenFunc(userModule.HandleDeleteDealershipUser))
	mux.Handle("POST /api/internal-user", protected.ThenFunc(userModule.HandleCreateInternalUser))
	mux.Handle("PATCH /api/internal-user/{uuid}", protected.ThenFunc(userModule.HandleUpdateInternalUser))
	mux.Handle("DELETE /api/internal-user/{uuid}", protected.ThenFunc(userModule.HandleDeleteInternalUser))

	uploadModule := upload.NewUploadModule(app)
	mux.Handle("POST /api/upload", protected.ThenFunc(uploadModule.HandlePostUpload))

	// Catalog routes
	internalAdmin := alice.New(app.Authenticate, app.RequireRole("admin"))

	catalogModule := catalog.NewCatalogModule(app)

	// Internal admin only
	mux.Handle("GET /api/catalog", internalAdmin.ThenFunc(catalogModule.HandleGetCatalog))
	mux.Handle("POST /api/catalog", internalAdmin.ThenFunc(catalogModule.HandlePostCatalog))
	mux.Handle("PATCH /api/catalog/{uuid}", internalAdmin.ThenFunc(catalogModule.HandlePatchCatalog))
	mux.Handle("DELETE /api/catalog/{uuid}", internalAdmin.ThenFunc(catalogModule.HandleDeleteCatalog))

	mux.Handle("POST /api/catalog/{uuid}/tags", internalAdmin.ThenFunc(catalogModule.HandlePostTag))
	mux.Handle("DELETE /api/catalog/{uuid}/tags/{tag}", internalAdmin.ThenFunc(catalogModule.HandleDeleteTag))

	// Public routes (authenticated dealership users) - MUST come before wildcard {uuid} route
	mux.Handle("GET /api/catalog/browse", protected.ThenFunc(catalogModule.HandleBrowseCatalog))
	mux.Handle("GET /api/catalog/categories", protected.ThenFunc(catalogModule.HandleGetCategories))
	mux.Handle("GET /api/catalog/tags", protected.ThenFunc(catalogModule.HandleGetAllTags))
	mux.Handle("GET /api/catalog/{uuid}", protected.ThenFunc(catalogModule.HandleGetCatalogItem))
	mux.Handle("GET /api/catalog/{uuid}/tags", protected.ThenFunc(catalogModule.HandleGetTags))

	// Price Group routes (internal admin only)
	priceGroupModule := pricegroup.NewPriceGroupModule(app)
	mux.Handle("GET /api/price-groups", internalAdmin.ThenFunc(priceGroupModule.HandleGetPriceGroups))
	mux.Handle("POST /api/price-groups", internalAdmin.ThenFunc(priceGroupModule.HandlePostPriceGroup))
	mux.Handle("GET /api/price-groups/{uuid}", internalAdmin.ThenFunc(priceGroupModule.HandleGetPriceGroup))
	mux.Handle("PATCH /api/price-groups/{uuid}", internalAdmin.ThenFunc(priceGroupModule.HandlePatchPriceGroup))
	mux.Handle("DELETE /api/price-groups/{uuid}", internalAdmin.ThenFunc(priceGroupModule.HandleDeletePriceGroup))

	mux.Handle("/", unprotected.ThenFunc(app.HandleNotFound))
	return standard.Then(mux)
}
