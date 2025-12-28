// ABOUTME: API request and response types for announcement management.
// ABOUTME: Defines structures for announcement CRUD and publishing operations.

package v1

import "time"

// AnnouncementItem represents a single announcement.
type AnnouncementItem struct {
	UUID        string     `json:"uuid"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	ActionURL   string     `json:"actionUrl,omitempty"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduledAt,omitempty"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// ListAnnouncementsRequest is the request for listing announcements.
type ListAnnouncementsRequest struct {
	Status   string `form:"status"`
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
}

// ListAnnouncementsResponse is the response for listing announcements.
type ListAnnouncementsResponse struct {
	Data  []AnnouncementItem `json:"data"`
	Total int64              `json:"total"`
}

// CreateAnnouncementRequest is the request for creating an announcement.
type CreateAnnouncementRequest struct {
	Title     string     `json:"title" binding:"required,max=255"`
	Content   string     `json:"content" binding:"required"`
	ActionURL string     `json:"actionUrl"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// UpdateAnnouncementRequest is the request for updating an announcement.
type UpdateAnnouncementRequest struct {
	Title     string     `json:"title" binding:"max=255"`
	Content   string     `json:"content"`
	ActionURL string     `json:"actionUrl"`
	ExpiresAt *time.Time `json:"expiresAt"`
}

// ScheduleAnnouncementRequest is the request for scheduling an announcement.
type ScheduleAnnouncementRequest struct {
	ScheduledAt time.Time `json:"scheduledAt" binding:"required"`
}
