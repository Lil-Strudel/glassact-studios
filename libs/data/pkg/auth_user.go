package data

const (
	ScopeAccess  = "access"
	ScopeLogin   = "login"
	ScopeRefresh = "refresh"
)

type AuthUser interface {
	GetID() int
	GetUUID() string
	GetEmail() string
	GetName() string
	GetAvatar() string
	GetRole() string
	GetIsActive() bool
	IsInternal() bool
	IsDealership() bool
	GetDealershipID() *int
	Can(action string) bool
}
