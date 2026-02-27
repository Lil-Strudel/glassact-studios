package data

import "fmt"

func GetAuthUserForToken(models *Models, tokenScope, tokenPlaintext string) (AuthUser, string, error) {
	var dealershipScopeStr, internalScopeStr string

	switch tokenScope {
	case ScopeAccess:
		dealershipScopeStr = DealershipScopeAccess
		internalScopeStr = InternalScopeAccess
	case ScopeLogin:
		dealershipScopeStr = DealershipScopeLogin
		internalScopeStr = InternalScopeLogin
	case ScopeRefresh:
		dealershipScopeStr = DealershipScopeRefresh
		internalScopeStr = InternalScopeRefresh
	default:
		return nil, "", fmt.Errorf("unknown scope: %s", tokenScope)
	}

	dealershipUser, found, err := models.DealershipUsers.GetForToken(dealershipScopeStr, tokenPlaintext)
	if err != nil {
		return nil, "", fmt.Errorf("dealership token lookup failed: %w", err)
	}

	if found && dealershipUser.IsActive {
		return dealershipUser, "dealership", nil
	}

	internalUser, found, err := models.InternalUsers.GetForToken(internalScopeStr, tokenPlaintext)
	if err != nil {
		return nil, "", fmt.Errorf("internal token lookup failed: %w", err)
	}

	if found && internalUser.IsActive {
		return internalUser, "internal", nil
	}

	return nil, "", fmt.Errorf("token not found or user inactive")
}
