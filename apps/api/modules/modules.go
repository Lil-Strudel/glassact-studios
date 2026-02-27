package modules

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/auth"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/catalog"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/chat"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/dealership"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/inlay"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/pricegroup"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/project"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/proof"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/upload"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/user"
	data "github.com/Lil-Strudel/glassact-studios/libs/data/pkg"
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

	canCreateProject := alice.New(app.Authenticate, app.RequirePermission(data.ActionCreateProject))
	canPlaceOrder := alice.New(app.Authenticate, app.RequirePermission(data.ActionPlaceOrder))

	projectModule := project.NewProjectModule(app)
	mux.Handle("GET /api/project", protected.ThenFunc(projectModule.HandleGetProjects))
	mux.Handle("POST /api/project", canCreateProject.ThenFunc(projectModule.HandlePostProject))
	mux.Handle("POST /api/project/with-inlays", canCreateProject.ThenFunc(projectModule.HandlePostProjectWithInlays))
	mux.Handle("GET /api/project/{uuid}", protected.ThenFunc(projectModule.HandleGetProjectByUUID))
	mux.Handle("PATCH /api/project/{uuid}", protected.ThenFunc(projectModule.HandlePatchProject))
	mux.Handle("DELETE /api/project/{uuid}", protected.ThenFunc(projectModule.HandleDeleteProject))
	mux.Handle("POST /api/project/{uuid}/submit", canCreateProject.ThenFunc(projectModule.HandleSubmitProject))
	mux.Handle("POST /api/project/{uuid}/place-order", canPlaceOrder.ThenFunc(projectModule.HandlePlaceOrder))

	inlayModule := inlay.NewInlayModule(app)
	mux.Handle("GET /api/project/{uuid}/inlays", protected.ThenFunc(inlayModule.HandleGetInlaysByProject))
	mux.Handle("POST /api/project/{uuid}/inlays/catalog", canCreateProject.ThenFunc(inlayModule.HandlePostCatalogInlay))
	mux.Handle("POST /api/project/{uuid}/inlays/custom", canCreateProject.ThenFunc(inlayModule.HandlePostCustomInlay))
	mux.Handle("GET /api/inlay/{uuid}", protected.ThenFunc(inlayModule.HandleGetInlayByUUID))
	mux.Handle("PATCH /api/inlay/{uuid}", protected.ThenFunc(inlayModule.HandlePatchInlay))
	mux.Handle("DELETE /api/inlay/{uuid}", protected.ThenFunc(inlayModule.HandleDeleteInlay))
	mux.Handle("PATCH /api/inlay/{uuid}/exclude", canCreateProject.ThenFunc(inlayModule.HandleExcludeInlay))

	chatModule := chat.NewChatModule(app)
	mux.Handle("GET /api/inlay/{uuid}/chats", protected.ThenFunc(chatModule.HandleGetInlayChats))
	mux.Handle("POST /api/inlay/{uuid}/chats", protected.ThenFunc(chatModule.HandlePostInlayChat))

	canCreateProof := alice.New(app.Authenticate, app.RequirePermission(data.ActionCreateProof))
	canApproveProof := alice.New(app.Authenticate, app.RequirePermission(data.ActionApproveProof))

	proofModule := proof.NewProofModule(app)
	mux.Handle("GET /api/inlay/{uuid}/proofs", protected.ThenFunc(proofModule.HandleGetProofsByInlay))
	mux.Handle("POST /api/inlay/{uuid}/proofs", canCreateProof.ThenFunc(proofModule.HandleCreateProof))
	mux.Handle("GET /api/proof/{uuid}", protected.ThenFunc(proofModule.HandleGetProof))
	mux.Handle("POST /api/proof/{uuid}/approve", canApproveProof.ThenFunc(proofModule.HandleApproveProof))
	mux.Handle("POST /api/proof/{uuid}/decline", canApproveProof.ThenFunc(proofModule.HandleDeclineProof))

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
	mux.Handle("GET /file/{path...}", unprotected.ThenFunc(uploadModule.HandleGetFile))

	canManageCatalog := alice.New(app.Authenticate, app.RequirePermission(data.ActionManageCatalog))
	canManagePriceGroups := alice.New(app.Authenticate, app.RequirePermission(data.ActionManagePriceGroups))

	catalogModule := catalog.NewCatalogModule(app)
	mux.Handle("GET /api/catalog", canManageCatalog.ThenFunc(catalogModule.HandleGetCatalog))
	mux.Handle("POST /api/catalog", canManageCatalog.ThenFunc(catalogModule.HandlePostCatalog))
	mux.Handle("PATCH /api/catalog/{uuid}", canManageCatalog.ThenFunc(catalogModule.HandlePatchCatalog))
	mux.Handle("DELETE /api/catalog/{uuid}", canManageCatalog.ThenFunc(catalogModule.HandleDeleteCatalog))

	mux.Handle("POST /api/catalog/{uuid}/tags", canManageCatalog.ThenFunc(catalogModule.HandlePostTag))
	mux.Handle("DELETE /api/catalog/{uuid}/tags/{tag}", canManageCatalog.ThenFunc(catalogModule.HandleDeleteTag))

	mux.Handle("GET /api/catalog/browse", protected.ThenFunc(catalogModule.HandleBrowseCatalog))
	mux.Handle("GET /api/catalog/categories", protected.ThenFunc(catalogModule.HandleGetCategories))
	mux.Handle("GET /api/catalog/tags", protected.ThenFunc(catalogModule.HandleGetAllTags))
	mux.Handle("GET /api/catalog/{uuid}", protected.ThenFunc(catalogModule.HandleGetCatalogItem))
	mux.Handle("GET /api/catalog/{uuid}/tags", protected.ThenFunc(catalogModule.HandleGetTags))

	priceGroupModule := pricegroup.NewPriceGroupModule(app)
	mux.Handle("GET /api/price-groups", canManagePriceGroups.ThenFunc(priceGroupModule.HandleGetPriceGroups))
	mux.Handle("POST /api/price-groups", canManagePriceGroups.ThenFunc(priceGroupModule.HandlePostPriceGroup))
	mux.Handle("GET /api/price-groups/{uuid}", canManagePriceGroups.ThenFunc(priceGroupModule.HandleGetPriceGroup))
	mux.Handle("PATCH /api/price-groups/{uuid}", canManagePriceGroups.ThenFunc(priceGroupModule.HandlePatchPriceGroup))
	mux.Handle("DELETE /api/price-groups/{uuid}", canManagePriceGroups.ThenFunc(priceGroupModule.HandleDeletePriceGroup))

	mux.Handle("/", unprotected.ThenFunc(app.HandleNotFound))
	return standard.Then(mux)
}
