// ABOUTME: AuthProvider model defines OAuth provider configuration.
// ABOUTME: Supports multiple OAuth platforms with configurable endpoints and PKCE.

package model

type AuthProvider struct {
	Base

	Name         string             `gorm:"type:varchar(255);not null;default:'';comment:Auth provider name"`
	Status       AuthProviderStatus `gorm:"type:varchar(20);not null;default:'disabled';comment:Status: enabled/disabled"`
	IsDefault    int                `gorm:"type:tinyint;not null;default:0;comment:Is default provider, 0-not, 1-yes"`
	AppID        string             `gorm:"type:varchar(255);not null;default:'';comment:App ID"`
	ClientID     string             `gorm:"type:varchar(255);not null;default:'';comment:Client ID"`
	ClientSecret string             `gorm:"type:varchar(1024);not null;default:'';comment:Client secret"`
	TokenType    string             `gorm:"type:varchar(1024);not null;default:'';comment:Token type"`
	RedirectURL  string             `gorm:"type:varchar(1024);not null;default:'';comment:Redirect URL"`
	AuthURL      string             `gorm:"type:varchar(1024);not null;default:'';comment:Auth URL"`
	TokenURL     string             `gorm:"type:varchar(1024);not null;default:'';comment:Token URL"`
	LogoutURI    string             `gorm:"type:varchar(1024);not null;default:'';comment:Logout URI"`
	Info         string             `gorm:"type:json;comment:Ext info"`
	UserInfoURL  string             `gorm:"column:user_info_url;type:varchar(500)"`
	FieldMapping string             `gorm:"column:field_mapping;type:text"`
	TokenInQuery bool               `gorm:"column:token_in_query;default:false"`
	ExtraHeaders string             `gorm:"column:extra_headers;type:text"`
	Scopes       string             `gorm:"column:scopes;type:varchar(500)"`
	PKCEEnabled  bool               `gorm:"column:pkce_enabled;default:false"`
}

func (*AuthProvider) TableName() string {
	return "uc_auth_provider"
}

// AuthProviderStatus enabled/disabled.
type AuthProviderStatus string

const (
	AuthProviderStatusEnabled  AuthProviderStatus = "enabled"
	AuthProviderStatusDisabled AuthProviderStatus = "disabled"

	AuthProviderGoogle  = "google"
	AuthProviderApple   = "apple"
	AuthProviderGithub  = "github"
	AuthProviderDiscord = "discord"
	AuthProviderTwitter = "twitter"
	AuthProviderWallet  = "wallet"
)
