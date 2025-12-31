// ABOUTME: Mock store implementations for testing.
// ABOUTME: Provides in-memory implementations of store interfaces.

package store

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/store/where"
	"gorm.io/gorm"
)

// Store implements store.IStore for testing.
type Store struct {
	aiProvider *AiProviderStore
	aiModel    *AiModelStore
}

var _ store.IStore = (*Store)(nil)

// NewStore creates a new mock store.
func NewStore() *Store {
	return &Store{
		aiProvider: NewAiProviderStore(),
		aiModel:    NewAiModelStore(),
	}
}

// DB returns a mock database instance.
func (m *Store) DB(ctx context.Context, wheres ...where.Where) *gorm.DB {
	return nil
}

// TX executes a function in a transaction context.
func (m *Store) TX(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

// Admin returns the system admin store.
func (m *Store) Admin() store.AdminStore {
	return nil
}

// Schedule returns the system schedule store.
func (m *Store) Schedule() store.ScheduleStore {
	return nil
}

// SysConfig returns the system config store.
func (m *Store) SysConfig() store.ConfigStore {
	return nil
}

// AppVersion returns the app version store.
func (m *Store) AppVersion() store.AppVersionStore {
	return nil
}

// SysRole returns the system role store.
func (m *Store) SysRole() store.SysRoleStore {
	return nil
}

// SysApi returns the system API store.
func (m *Store) SysApi() store.SysApiStore {
	return nil
}

// SysMenu returns the system menu store.
func (m *Store) SysMenu() store.SysMenuStore {
	return nil
}

// SysRoleMenu returns the system role-menu store.
func (m *Store) SysRoleMenu() store.SysRoleMenuStore {
	return nil
}

// Bot returns the bot store.
func (m *Store) Bot() store.BotStore {
	return nil
}

// BotChannel returns the bot channel store.
func (m *Store) BotChannel() store.BotChannelStore {
	return nil
}

// BotAdmin returns the bot admin store.
func (m *Store) BotAdmin() store.BotAdminStore {
	return nil
}

// User returns the user store.
func (m *Store) User() store.UserStore {
	return nil
}

// UserAccount returns the user account store.
func (m *Store) UserAccount() store.UserAccountStore {
	return nil
}

// AuthProvider returns the auth provider store.
func (m *Store) AuthProvider() store.AuthProviderStore {
	return nil
}

// App returns the app store.
func (m *Store) App() store.AppStore {
	return nil
}

// ApiKey returns the API key store.
func (m *Store) ApiKey() store.ApiKeyStore {
	return nil
}

// NtfMessage returns the notification message store.
func (m *Store) NtfMessage() store.NtfMessageStore {
	return nil
}

// NtfAnnouncement returns the notification announcement store.
func (m *Store) NtfAnnouncement() store.NtfAnnouncementStore {
	return nil
}

// NtfPreference returns the notification preference store.
func (m *Store) NtfPreference() store.NtfPreferenceStore {
	return nil
}

// AiProvider returns the AI provider store.
func (m *Store) AiProvider() store.AiProviderStore {
	return m.aiProvider
}

// AiModel returns the AI model store.
func (m *Store) AiModel() store.AiModelStore {
	return m.aiModel
}

// AiQuotaTier returns the AI quota tier store.
func (m *Store) AiQuotaTier() store.AiQuotaTierStore {
	return nil
}

// AiUserQuota returns the AI user quota store.
func (m *Store) AiUserQuota() store.AiUserQuotaStore {
	return nil
}

// AiSession returns the AI session store.
func (m *Store) AiSession() store.AiSessionStore {
	return nil
}

// AiMessage returns the AI message store.
func (m *Store) AiMessage() store.AiMessageStore {
	return nil
}
