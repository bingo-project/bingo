// ABOUTME: User notification preference model.
// ABOUTME: Stores per-category and per-channel notification settings as JSON.

package model

import (
	"encoding/json"
	"time"
)

// ChannelPreference defines per-channel settings.
type ChannelPreference struct {
	InApp bool `json:"in_app"`
	Email bool `json:"email"`
}

// NotificationPreferences defines all category preferences.
type NotificationPreferences struct {
	System      ChannelPreference `json:"system"`
	Security    ChannelPreference `json:"security"`
	Transaction ChannelPreference `json:"transaction"`
	Social      ChannelPreference `json:"social"`
}

// DefaultPreferences returns the default notification preferences.
func DefaultPreferences() NotificationPreferences {
	return NotificationPreferences{
		System:      ChannelPreference{InApp: true, Email: false},
		Security:    ChannelPreference{InApp: true, Email: true},
		Transaction: ChannelPreference{InApp: true, Email: true},
		Social:      ChannelPreference{InApp: true, Email: false},
	}
}

// NtfPreferenceM represents user notification preferences.
type NtfPreferenceM struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	UserID      string    `gorm:"type:varchar(64);uniqueIndex:uk_user_id" json:"userId"`
	Preferences string    `gorm:"type:json" json:"preferences"`
	CreatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}

func (*NtfPreferenceM) TableName() string {
	return "ntf_preference"
}

// GetPreferences parses and returns the notification preferences.
func (m *NtfPreferenceM) GetPreferences() NotificationPreferences {
	if m.Preferences == "" {
		return DefaultPreferences()
	}
	var prefs NotificationPreferences
	if err := json.Unmarshal([]byte(m.Preferences), &prefs); err != nil {
		return DefaultPreferences()
	}
	return prefs
}

// SetPreferences serializes and sets the notification preferences.
func (m *NtfPreferenceM) SetPreferences(prefs NotificationPreferences) error {
	data, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	m.Preferences = string(data)
	return nil
}
