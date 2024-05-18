package model

type AuthProvider struct {
	Base

	Name         string             `gorm:"type:varchar(255);not null;default:'';comment:Auth provider name"`
	Status       AuthProviderStatus `gorm:"type:tinyint;not null;default:1;comment:Status, 1-enabled, 2-disabled"`
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
}

func (*AuthProvider) TableName() string {
	return "uc_auth_provider"
}

// AuthProviderStatus 1-enabled, 2-disabled.
type AuthProviderStatus int

const (
	AuthProviderStatusEnabled  AuthProviderStatus = 1
	AuthProviderStatusDisabled AuthProviderStatus = 2
)
