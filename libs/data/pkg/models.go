package data

import (
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Models struct {
	CatalogItems            CatalogItemModel
	Dashboard               DashboardModel
	DealershipAccounts      DealershipAccountModel
	DealershipTokens        DealershipTokenModel
	DealershipUsers         DealershipUserModel
	Dealerships             DealershipModel
	GlassColors             GlassColorModel
	Grouts                  GroutModel
	InlayChats              InlayChatModel
	InlayMilestones         InlayMilestoneModel
	InlayProofs             InlayProofModel
	InlayUpdates            InlayUpdateModel
	Inlays                  InlayModel
	InternalAccounts        InternalAccountModel
	InternalTokens          InternalTokenModel
	InternalUsers           InternalUserModel
	Invoices                InvoiceModel
	Notifications           NotificationModel
	NotificationPreferences NotificationPreferencesModel
	OrderSnapshots          OrderSnapshotModel
	PriceGroups             PriceGroupModel
	ProjectChats            ProjectChatModel
	Projects                ProjectModel
	SupportArticles         SupportArticleModel
	Pool                    *pgxpool.Pool
	STDB                    *sql.DB
}

func NewModels(db *pgxpool.Pool, stdb *sql.DB) Models {
	return Models{
		CatalogItems:            CatalogItemModel{DB: db, STDB: stdb},
		Dashboard:               DashboardModel{DB: db, STDB: stdb},
		DealershipAccounts:      DealershipAccountModel{DB: db, STDB: stdb},
		DealershipTokens:        DealershipTokenModel{DB: db, STDB: stdb},
		DealershipUsers:         DealershipUserModel{DB: db, STDB: stdb},
		Dealerships:             DealershipModel{DB: db, STDB: stdb},
		GlassColors:             GlassColorModel{DB: db, STDB: stdb},
		Grouts:                  GroutModel{DB: db, STDB: stdb},
		InlayChats:              InlayChatModel{DB: db, STDB: stdb},
		InlayMilestones:         InlayMilestoneModel{DB: db, STDB: stdb},
		InlayProofs:             InlayProofModel{DB: db, STDB: stdb},
		InlayUpdates:            InlayUpdateModel{DB: db, STDB: stdb},
		Inlays:                  InlayModel{DB: db, STDB: stdb},
		InternalAccounts:        InternalAccountModel{DB: db, STDB: stdb},
		InternalTokens:          InternalTokenModel{DB: db, STDB: stdb},
		InternalUsers:           InternalUserModel{DB: db, STDB: stdb},
		Invoices:                InvoiceModel{DB: db, STDB: stdb},
		Notifications:           NotificationModel{DB: db, STDB: stdb},
		NotificationPreferences: NotificationPreferencesModel{DB: db, STDB: stdb},
		OrderSnapshots:          OrderSnapshotModel{DB: db, STDB: stdb},
		PriceGroups:             PriceGroupModel{DB: db, STDB: stdb},
		ProjectChats:            ProjectChatModel{DB: db, STDB: stdb},
		Projects:                ProjectModel{DB: db, STDB: stdb},
		SupportArticles:         SupportArticleModel{DB: db, STDB: stdb},
		Pool:                    db,
		STDB:                    stdb,
	}
}
