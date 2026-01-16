package modules

import (
	"net/http"

	"github.com/Lil-Strudel/glassact-studios/apps/api/app"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/auth"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/dealership"
	"github.com/Lil-Strudel/glassact-studios/apps/api/modules/inlay"
	inlayChat "github.com/Lil-Strudel/glassact-studios/apps/api/modules/inlay-chat"
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
	mux.Handle("GET /api/user", protected.ThenFunc(userModule.HandleGetUsers))
	mux.Handle("GET /api/user/self", protected.ThenFunc(userModule.HandleGetUserSelf))
	mux.Handle("GET /api/user/{uuid}", protected.ThenFunc(userModule.HandleGetUserByUUID))
	mux.Handle("POST /api/user", protected.ThenFunc(userModule.HandlePostUser))

	uploadModule := upload.NewUploadModule(app)
	mux.Handle("POST /api/upload", protected.ThenFunc(uploadModule.HandlePostUpload))

	mux.Handle("/", unprotected.ThenFunc(app.HandleNotFound))
	return standard.Then(mux)
}
