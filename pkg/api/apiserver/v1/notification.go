// ABOUTME: API request and response types for notification endpoints.
// ABOUTME: Defines structures for notification list, preferences, and read operations.

package v1

import "time"

// NotificationItem represents a single notification in the list.
type NotificationItem struct {
	UUID      string    `json:"uuid"`
	Source    string    `json:"source"` // "message" or "announcement"
	Category  string    `json:"category"`
	Type      string    `json:"type,omitempty"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ActionURL string    `json:"actionUrl,omitempty"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListNotificationsRequest is the request for listing notifications.
type ListNotificationsRequest struct {
	Category string `form:"category"`
	IsRead   *bool  `form:"is_read"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// ListNotificationsResponse is the response for listing notifications.
type ListNotificationsResponse struct {
	Data  []NotificationItem `json:"data"`
	Total int64              `json:"total"`
}

// UnreadCountResponse is the response for unread count.
type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

// ChannelPreference defines per-channel settings.
type ChannelPreference struct {
	InApp bool `json:"inApp"`
	Email bool `json:"email"`
}

// NotificationPreferences defines all category preferences.
type NotificationPreferences struct {
	System      ChannelPreference `json:"system"`
	Security    ChannelPreference `json:"security"`
	Transaction ChannelPreference `json:"transaction"`
	Social      ChannelPreference `json:"social"`
}

// UpdatePreferencesRequest is the request for updating preferences.
// The request body is the NotificationPreferences directly, without wrapper.
type UpdatePreferencesRequest = NotificationPreferences
