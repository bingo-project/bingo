package v1

import (
	"time"

	"github.com/bingo-project/component-base/util/gormutil"
)

type AuthProviderBrief struct {
	Name        string `json:"name"`        // Auth provider name
	IsDefault   int    `json:"isDefault"`   // Is default provider, 0-not, 1-yes
	RedirectURL string `json:"redirectUrl"` // Redirect URL
	AuthURL     string `json:"authUrl"`     // Auth URL
}

type AuthProviderInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Name         string `json:"name"`         // Auth provider name
	Status       int    `json:"status"`       // Status, 1-enabled, 2-disabled
	IsDefault    int    `json:"isDefault"`    // Is default provider, 0-not, 1-yes
	AppID        string `json:"appId"`        // App ID
	ClientID     string `json:"clientId"`     // Client ID
	ClientSecret string `json:"clientSecret"` // Client secret
	TokenType    string `json:"tokenType"`    // Token type
	RedirectURL  string `json:"redirectUrl"`  // Redirect URL
	AuthURL      string `json:"authUrl"`      // Auth URL
	TokenURL     string `json:"tokenUrl"`     // Token URL
	LogoutURI    string `json:"logoutUri"`    // Logout URI
	Info         string `json:"info"`         // Ext info
}

type ListAuthProviderRequest struct {
	gormutil.ListOptions

	Name      *string `json:"name"`      // Auth provider name
	Status    *int    `json:"status"`    // Status, 1-enabled, 2-disabled
	IsDefault *int    `json:"isDefault"` // Is default provider, 0-not, 1-yes
}

type ListAuthProviderResponse struct {
	Total int64              `json:"total"`
	Data  []AuthProviderInfo `json:"data"`
}

type CreateAuthProviderRequest struct {
	Name         string `json:"name"`         // Auth provider name
	Status       int    `json:"status"`       // Status, 1-enabled, 2-disabled
	IsDefault    int    `json:"isDefault"`    // Is default provider, 0-not, 1-yes
	AppID        string `json:"appId"`        // App ID
	ClientID     string `json:"clientId"`     // Client ID
	ClientSecret string `json:"clientSecret"` // Client secret
	TokenType    string `json:"tokenType"`    // Token type
	RedirectURL  string `json:"redirectUrl"`  // Redirect URL
	AuthURL      string `json:"authUrl"`      // Auth URL
	TokenURL     string `json:"tokenUrl"`     // Token URL
	LogoutURI    string `json:"logoutUri"`    // Logout URI
	Info         string `json:"info"`         // Ext info
}

type UpdateAuthProviderRequest struct {
	Name         *string `json:"name"`         // Auth provider name
	Status       *int    `json:"status"`       // Status, 1-enabled, 2-disabled
	IsDefault    *int    `json:"isDefault"`    // Is default provider, 0-not, 1-yes
	AppID        *string `json:"appId"`        // App ID
	ClientID     *string `json:"clientId"`     // Client ID
	ClientSecret *string `json:"clientSecret"` // Client secret
	TokenType    *string `json:"tokenType"`    // Token type
	RedirectURL  *string `json:"redirectUrl"`  // Redirect URL
	AuthURL      *string `json:"authUrl"`      // Auth URL
	TokenURL     *string `json:"tokenUrl"`     // Token URL
	LogoutURI    *string `json:"logoutUri"`    // Logout URI
	Info         *string `json:"info"`         // Ext info
}
