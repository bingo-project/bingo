# 通知系统实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现完整的通知系统，包括用户通知偏好、通知中心、公告管理和实时推送。

**Architecture:** 三层架构（Handler → Biz → Store），Redis Pub/Sub 实现服务间通信，WebSocket 实现实时推送，Asynq 处理定时发布任务。

**Tech Stack:** Go 1.24+, Gin, GORM, Redis, Asynq, WebSocket (bingo-project/websocket)

**Design Doc:** [2025-12-28-notification-design.md](./2025-12-28-notification-design.md)

---

## Phase 1: 数据层基础

建立数据模型、迁移文件和 Store 层。

### Task 1.1: 创建通知相关的 Model

**Files:**
- Create: `internal/pkg/model/ntf_message.go`
- Create: `internal/pkg/model/ntf_announcement.go`
- Create: `internal/pkg/model/ntf_preference.go`

**Step 1: 创建 ntf_message.go**

```go
// ABOUTME: Notification message model for individual user notifications.
// ABOUTME: Stores personal notifications like login alerts, transaction updates.

package model

import (
	"time"

	"github.com/bingo-project/bingo/internal/pkg/facade"
)

// NotificationCategory defines the notification category type.
type NotificationCategory string

const (
	NotificationCategorySystem      NotificationCategory = "system"
	NotificationCategorySecurity    NotificationCategory = "security"
	NotificationCategoryTransaction NotificationCategory = "transaction"
	NotificationCategorySocial      NotificationCategory = "social"
)

// NtfMessageM represents a notification message.
type NtfMessageM struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UUID      string    `gorm:"type:varchar(64);uniqueIndex:uk_uuid" json:"uuid"`
	UserID    string    `gorm:"type:varchar(64);index:idx_user_id;not null" json:"userId"`
	Category  string    `gorm:"type:varchar(32);index:idx_category;not null" json:"category"`
	Type      string    `gorm:"type:varchar(64);not null" json:"type"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	ActionURL string    `gorm:"type:varchar(512);not null;default:''" json:"actionUrl"`
	IsRead    bool      `gorm:"type:tinyint(1);not null;default:0" json:"isRead"`
	ReadAt    *time.Time `gorm:"type:timestamp;default:null" json:"readAt"`
	CreatedAt time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index:idx_created_at" json:"createdAt"`
}

func (*NtfMessageM) TableName() string {
	return "ntf_message"
}

// BeforeCreate generates UUID before creating.
func (m *NtfMessageM) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == "" {
		m.UUID = facade.Snowflake.Generate().String()
	}
	return nil
}
```

**Step 2: 创建 ntf_announcement.go**

```go
// ABOUTME: Announcement model for system-wide broadcasts.
// ABOUTME: Supports draft, scheduled, and published states.

package model

import (
	"time"

	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/facade"
)

// AnnouncementStatus defines the announcement status type.
type AnnouncementStatus string

const (
	AnnouncementStatusDraft     AnnouncementStatus = "draft"
	AnnouncementStatusScheduled AnnouncementStatus = "scheduled"
	AnnouncementStatusPublished AnnouncementStatus = "published"
)

// NtfAnnouncementM represents a system announcement.
type NtfAnnouncementM struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	UUID        string     `gorm:"type:varchar(64);uniqueIndex:uk_uuid" json:"uuid"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Content     string     `gorm:"type:text" json:"content"`
	ActionURL   string     `gorm:"type:varchar(512);not null;default:''" json:"actionUrl"`
	Status      string     `gorm:"type:varchar(32);index:idx_status;not null;default:'draft'" json:"status"`
	ScheduledAt *time.Time `gorm:"type:timestamp;default:null" json:"scheduledAt"`
	PublishedAt *time.Time `gorm:"type:timestamp;default:null" json:"publishedAt"`
	ExpiresAt   *time.Time `gorm:"type:timestamp;default:null" json:"expiresAt"`
	CreatedAt   time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index:idx_created_at" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)" json:"updatedAt"`
}

func (*NtfAnnouncementM) TableName() string {
	return "ntf_announcement"
}

// BeforeCreate generates UUID before creating.
func (m *NtfAnnouncementM) BeforeCreate(tx *gorm.DB) error {
	if m.UUID == "" {
		m.UUID = facade.Snowflake.Generate().String()
	}
	return nil
}

// NtfAnnouncementReadM represents a user's read status for an announcement.
type NtfAnnouncementReadM struct {
	UserID         string    `gorm:"type:varchar(64);primaryKey" json:"userId"`
	AnnouncementID uint64    `gorm:"primaryKey" json:"announcementId"`
	ReadAt         time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)" json:"readAt"`
}

func (*NtfAnnouncementReadM) TableName() string {
	return "ntf_announcement_read"
}
```

**Step 3: 创建 ntf_preference.go**

```go
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
```

**Step 4: 验证编译**

Run: `go build ./internal/pkg/model/...`
Expected: 编译成功，无错误

**Step 5: Commit**

```bash
git add internal/pkg/model/ntf_*.go
git commit -m "feat(model): add notification models

- NtfMessageM for personal notifications
- NtfAnnouncementM and NtfAnnouncementReadM for announcements
- NtfPreferenceM for user notification preferences"
```

---

### Task 1.2: 创建数据库迁移文件

**Files:**
- Create: `internal/pkg/database/migration/2025_12_28_100000_create_ntf_message_table.go`
- Create: `internal/pkg/database/migration/2025_12_28_100001_create_ntf_announcement_table.go`
- Create: `internal/pkg/database/migration/2025_12_28_100002_create_ntf_preference_table.go`

**Step 1: 创建 ntf_message 迁移**

```go
package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateNtfMessageTable struct {
	ID        uint64     `gorm:"primaryKey"`
	UUID      string     `gorm:"type:varchar(64);uniqueIndex:uk_uuid"`
	UserID    string     `gorm:"type:varchar(64);index:idx_user_id;not null"`
	Category  string     `gorm:"type:varchar(32);index:idx_category;not null"`
	Type      string     `gorm:"type:varchar(64);not null"`
	Title     string     `gorm:"type:varchar(255);not null"`
	Content   string     `gorm:"type:text"`
	ActionURL string     `gorm:"type:varchar(512);not null;default:''"`
	IsRead    bool       `gorm:"type:tinyint(1);not null;default:0"`
	ReadAt    *time.Time `gorm:"type:timestamp;default:null"`
	CreatedAt time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index:idx_created_at"`
}

func (CreateNtfMessageTable) TableName() string {
	return "ntf_message"
}

func (CreateNtfMessageTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateNtfMessageTable{})
}

func (CreateNtfMessageTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateNtfMessageTable{})
}

func init() {
	migrate.Add("2025_12_28_100000_create_ntf_message_table", CreateNtfMessageTable{}.Up, CreateNtfMessageTable{}.Down)
}
```

**Step 2: 创建 ntf_announcement 迁移**

```go
package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateNtfAnnouncementTable struct {
	ID          uint64     `gorm:"primaryKey"`
	UUID        string     `gorm:"type:varchar(64);uniqueIndex:uk_uuid"`
	Title       string     `gorm:"type:varchar(255);not null"`
	Content     string     `gorm:"type:text"`
	ActionURL   string     `gorm:"type:varchar(512);not null;default:''"`
	Status      string     `gorm:"type:varchar(32);index:idx_status;not null;default:'draft'"`
	ScheduledAt *time.Time `gorm:"type:timestamp;default:null"`
	PublishedAt *time.Time `gorm:"type:timestamp;default:null"`
	ExpiresAt   *time.Time `gorm:"type:timestamp;default:null"`
	CreatedAt   time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3);index:idx_created_at"`
	UpdatedAt   time.Time  `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateNtfAnnouncementTable) TableName() string {
	return "ntf_announcement"
}

type CreateNtfAnnouncementReadTable struct {
	UserID         string    `gorm:"type:varchar(64);primaryKey"`
	AnnouncementID uint64    `gorm:"primaryKey"`
	ReadAt         time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
}

func (CreateNtfAnnouncementReadTable) TableName() string {
	return "ntf_announcement_read"
}

func (CreateNtfAnnouncementTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateNtfAnnouncementTable{})
	_ = migrator.AutoMigrate(&CreateNtfAnnouncementReadTable{})
}

func (CreateNtfAnnouncementTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateNtfAnnouncementReadTable{})
	_ = migrator.DropTable(&CreateNtfAnnouncementTable{})
}

func init() {
	migrate.Add("2025_12_28_100001_create_ntf_announcement_table", CreateNtfAnnouncementTable{}.Up, CreateNtfAnnouncementTable{}.Down)
}
```

**Step 3: 创建 ntf_preference 迁移**

```go
package migration

import (
	"time"

	"github.com/bingo-project/bingoctl/pkg/migrate"
	"gorm.io/gorm"
)

type CreateNtfPreferenceTable struct {
	ID          uint64    `gorm:"primaryKey"`
	UserID      string    `gorm:"type:varchar(64);uniqueIndex:uk_user_id"`
	Preferences string    `gorm:"type:json"`
	CreatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)"`
	UpdatedAt   time.Time `gorm:"type:DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)"`
}

func (CreateNtfPreferenceTable) TableName() string {
	return "ntf_preference"
}

func (CreateNtfPreferenceTable) Up(migrator gorm.Migrator) {
	_ = migrator.AutoMigrate(&CreateNtfPreferenceTable{})
}

func (CreateNtfPreferenceTable) Down(migrator gorm.Migrator) {
	_ = migrator.DropTable(&CreateNtfPreferenceTable{})
}

func init() {
	migrate.Add("2025_12_28_100002_create_ntf_preference_table", CreateNtfPreferenceTable{}.Up, CreateNtfPreferenceTable{}.Down)
}
```

**Step 4: 运行迁移验证**

Run: `bingo migrate up`
Expected: 成功创建 3 张表

**Step 5: Commit**

```bash
git add internal/pkg/database/migration/2025_12_28_*.go
git commit -m "feat(migration): add notification tables

- ntf_message for personal notifications
- ntf_announcement and ntf_announcement_read for announcements
- ntf_preference for user preferences"
```

---

### Task 1.3: 创建 Store 层

**Files:**
- Create: `internal/pkg/store/ntf_message.go`
- Create: `internal/pkg/store/ntf_announcement.go`
- Create: `internal/pkg/store/ntf_preference.go`
- Modify: `internal/pkg/store/store.go`

**Step 1: 创建 ntf_message store**

```go
// ABOUTME: Store layer for notification messages.
// ABOUTME: Provides CRUD operations for personal notifications.

package store

import (
	"context"

	genericstore "github.com/bingo-project/component-base/store"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type NtfMessageStore interface {
	Create(ctx context.Context, obj *model.NtfMessageM) error
	Get(ctx context.Context, opts *where.Options) (*model.NtfMessageM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.NtfMessageM, error)
	Update(ctx context.Context, obj *model.NtfMessageM, opts *where.Options) error
	Delete(ctx context.Context, opts *where.Options) error

	NtfMessageExpansion
}

type NtfMessageExpansion interface {
	GetByUUID(ctx context.Context, uuid string) (*model.NtfMessageM, error)
	CountUnread(ctx context.Context, userID string) (int64, error)
	MarkAsRead(ctx context.Context, userID string, uuid string) error
	MarkAllAsRead(ctx context.Context, userID string) error
}

type ntfMessageStore struct {
	*genericstore.Store[model.NtfMessageM]
}

func NewNtfMessageStore(ds *datastore) *ntfMessageStore {
	return &ntfMessageStore{
		Store: genericstore.NewStore[model.NtfMessageM](ds, NewLogger()),
	}
}

func (s *ntfMessageStore) GetByUUID(ctx context.Context, uuid string) (*model.NtfMessageM, error) {
	return s.Get(ctx, where.F("uuid", uuid))
}

func (s *ntfMessageStore) CountUnread(ctx context.Context, userID string) (int64, error) {
	return s.Count(ctx, where.F("user_id", userID).F("is_read", false))
}

func (s *ntfMessageStore) MarkAsRead(ctx context.Context, userID string, uuid string) error {
	return s.DB(ctx).Model(&model.NtfMessageM{}).
		Where("user_id = ? AND uuid = ?", userID, uuid).
		Updates(map[string]any{"is_read": true, "read_at": time.Now()}).Error
}

func (s *ntfMessageStore) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.DB(ctx).Model(&model.NtfMessageM{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]any{"is_read": true, "read_at": time.Now()}).Error
}
```

**Step 2: 创建 ntf_announcement store**

```go
// ABOUTME: Store layer for announcements.
// ABOUTME: Provides CRUD and read status operations for system announcements.

package store

import (
	"context"
	"time"

	genericstore "github.com/bingo-project/component-base/store"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type NtfAnnouncementStore interface {
	Create(ctx context.Context, obj *model.NtfAnnouncementM) error
	Get(ctx context.Context, opts *where.Options) (*model.NtfAnnouncementM, error)
	List(ctx context.Context, opts *where.Options) (int64, []*model.NtfAnnouncementM, error)
	Update(ctx context.Context, obj *model.NtfAnnouncementM, opts *where.Options) error
	Delete(ctx context.Context, opts *where.Options) error

	NtfAnnouncementExpansion
}

type NtfAnnouncementExpansion interface {
	GetByUUID(ctx context.Context, uuid string) (*model.NtfAnnouncementM, error)
	ListPublished(ctx context.Context, opts *where.Options) (int64, []*model.NtfAnnouncementM, error)
	IsRead(ctx context.Context, userID string, announcementID uint64) (bool, error)
	MarkAsRead(ctx context.Context, userID string, announcementID uint64) error
	CountUnreadForUser(ctx context.Context, userID string) (int64, error)
}

type ntfAnnouncementStore struct {
	*genericstore.Store[model.NtfAnnouncementM]
}

func NewNtfAnnouncementStore(ds *datastore) *ntfAnnouncementStore {
	return &ntfAnnouncementStore{
		Store: genericstore.NewStore[model.NtfAnnouncementM](ds, NewLogger()),
	}
}

func (s *ntfAnnouncementStore) GetByUUID(ctx context.Context, uuid string) (*model.NtfAnnouncementM, error) {
	return s.Get(ctx, where.F("uuid", uuid))
}

func (s *ntfAnnouncementStore) ListPublished(ctx context.Context, opts *where.Options) (int64, []*model.NtfAnnouncementM, error) {
	opts = opts.F("status", string(model.AnnouncementStatusPublished))
	// Exclude expired announcements
	db := s.DB(ctx).Where("expires_at IS NULL OR expires_at > ?", time.Now())
	return s.ListWithDB(ctx, db, opts)
}

func (s *ntfAnnouncementStore) IsRead(ctx context.Context, userID string, announcementID uint64) (bool, error) {
	var count int64
	err := s.DB(ctx).Model(&model.NtfAnnouncementReadM{}).
		Where("user_id = ? AND announcement_id = ?", userID, announcementID).
		Count(&count).Error
	return count > 0, err
}

func (s *ntfAnnouncementStore) MarkAsRead(ctx context.Context, userID string, announcementID uint64) error {
	read := &model.NtfAnnouncementReadM{
		UserID:         userID,
		AnnouncementID: announcementID,
		ReadAt:         time.Now(),
	}
	return s.DB(ctx).Create(read).Error
}

func (s *ntfAnnouncementStore) CountUnreadForUser(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := s.DB(ctx).Model(&model.NtfAnnouncementM{}).
		Where("status = ?", string(model.AnnouncementStatusPublished)).
		Where("expires_at IS NULL OR expires_at > ?", time.Now()).
		Where("id NOT IN (?)",
			s.DB(ctx).Model(&model.NtfAnnouncementReadM{}).
				Select("announcement_id").
				Where("user_id = ?", userID),
		).
		Count(&count).Error
	return count, err
}
```

**Step 3: 创建 ntf_preference store**

```go
// ABOUTME: Store layer for notification preferences.
// ABOUTME: Provides CRUD operations for user notification settings.

package store

import (
	"context"

	genericstore "github.com/bingo-project/component-base/store"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type NtfPreferenceStore interface {
	Create(ctx context.Context, obj *model.NtfPreferenceM) error
	Get(ctx context.Context, opts *where.Options) (*model.NtfPreferenceM, error)
	Update(ctx context.Context, obj *model.NtfPreferenceM, opts *where.Options) error

	NtfPreferenceExpansion
}

type NtfPreferenceExpansion interface {
	GetByUserID(ctx context.Context, userID string) (*model.NtfPreferenceM, error)
	Upsert(ctx context.Context, userID string, prefs model.NotificationPreferences) error
}

type ntfPreferenceStore struct {
	*genericstore.Store[model.NtfPreferenceM]
}

func NewNtfPreferenceStore(ds *datastore) *ntfPreferenceStore {
	return &ntfPreferenceStore{
		Store: genericstore.NewStore[model.NtfPreferenceM](ds, NewLogger()),
	}
}

func (s *ntfPreferenceStore) GetByUserID(ctx context.Context, userID string) (*model.NtfPreferenceM, error) {
	return s.Get(ctx, where.F("user_id", userID))
}

func (s *ntfPreferenceStore) Upsert(ctx context.Context, userID string, prefs model.NotificationPreferences) error {
	pref := &model.NtfPreferenceM{UserID: userID}
	if err := pref.SetPreferences(prefs); err != nil {
		return err
	}

	return s.DB(ctx).
		Where("user_id = ?", userID).
		Assign(model.NtfPreferenceM{Preferences: pref.Preferences}).
		FirstOrCreate(pref).Error
}
```

**Step 4: 修改 store.go 添加接口**

在 `IStore` 接口中添加：

```go
// Notification stores
NtfMessage() NtfMessageStore
NtfAnnouncement() NtfAnnouncementStore
NtfPreference() NtfPreferenceStore
```

在 `datastore` 结构体方法中添加：

```go
func (ds *datastore) NtfMessage() NtfMessageStore {
	return NewNtfMessageStore(ds)
}

func (ds *datastore) NtfAnnouncement() NtfAnnouncementStore {
	return NewNtfAnnouncementStore(ds)
}

func (ds *datastore) NtfPreference() NtfPreferenceStore {
	return NewNtfPreferenceStore(ds)
}
```

**Step 5: 验证编译**

Run: `go build ./internal/pkg/store/...`
Expected: 编译成功

**Step 6: Commit**

```bash
git add internal/pkg/store/ntf_*.go internal/pkg/store/store.go
git commit -m "feat(store): add notification store layer

- NtfMessageStore for personal notifications
- NtfAnnouncementStore for announcements with read status
- NtfPreferenceStore for user preferences"
```

---

## Phase 2: apiserver 通知中心

实现用户端的通知列表、偏好设置等功能。

### Task 2.1: 创建 API 请求/响应结构

**Files:**
- Create: `pkg/api/apiserver/v1/notification.go`

**Step 1: 创建通知 API 结构**

```go
// ABOUTME: API request and response types for notification endpoints.
// ABOUTME: Defines structures for notification list, preferences, and read operations.

package v1

import "time"

// NotificationItem represents a single notification in the list.
type NotificationItem struct {
	UUID      string     `json:"uuid"`
	Source    string     `json:"source"` // "message" or "announcement"
	Category  string     `json:"category"`
	Type      string     `json:"type,omitempty"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	ActionURL string     `json:"actionUrl,omitempty"`
	IsRead    bool       `json:"isRead"`
	CreatedAt time.Time  `json:"createdAt"`
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

// GetPreferencesResponse is the response for getting preferences.
type GetPreferencesResponse struct {
	Preferences NotificationPreferences `json:"preferences"`
}

// UpdatePreferencesRequest is the request for updating preferences.
type UpdatePreferencesRequest struct {
	Preferences NotificationPreferences `json:"preferences" binding:"required"`
}
```

**Step 2: Commit**

```bash
git add pkg/api/apiserver/v1/notification.go
git commit -m "feat(api): add notification API types for apiserver"
```

---

### Task 2.2: 创建 apiserver Biz 层

**Files:**
- Create: `internal/apiserver/biz/notification/notification.go`
- Create: `internal/apiserver/biz/notification/preference.go`
- Modify: `internal/apiserver/biz/biz.go`

**Step 1: 创建 notification biz**

```go
// ABOUTME: Business logic for notification center.
// ABOUTME: Handles notification list, read status, and deletion.

package notification

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/store/where"
)

type NotificationBiz interface {
	List(ctx context.Context, userID string, req *v1.ListNotificationsRequest) (*v1.ListNotificationsResponse, error)
	UnreadCount(ctx context.Context, userID string) (*v1.UnreadCountResponse, error)
	MarkAsRead(ctx context.Context, userID string, uuid string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, userID string, uuid string) error
}

type notificationBiz struct {
	ds store.IStore
}

func New(ds store.IStore) NotificationBiz {
	return &notificationBiz{ds: ds}
}

func (b *notificationBiz) List(ctx context.Context, userID string, req *v1.ListNotificationsRequest) (*v1.ListNotificationsResponse, error) {
	// Query personal notifications
	msgOpts := where.F("user_id", userID).Page(req.Page, req.PageSize).Order("created_at DESC")
	if req.Category != "" {
		msgOpts = msgOpts.F("category", req.Category)
	}
	if req.IsRead != nil {
		msgOpts = msgOpts.F("is_read", *req.IsRead)
	}

	msgTotal, messages, err := b.ds.NtfMessage().List(ctx, msgOpts)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("list messages: %v", err)
	}

	// Query announcements (only system category or no filter)
	var annTotal int64
	var announcements []*model.NtfAnnouncementM
	if req.Category == "" || req.Category == string(model.NotificationCategorySystem) {
		annOpts := where.Page(req.Page, req.PageSize).Order("created_at DESC")
		annTotal, announcements, err = b.ds.NtfAnnouncement().ListPublished(ctx, annOpts)
		if err != nil {
			return nil, errno.ErrDBRead.WithMessage("list announcements: %v", err)
		}
	}

	// Merge and build response
	items := make([]v1.NotificationItem, 0, len(messages)+len(announcements))

	for _, msg := range messages {
		items = append(items, v1.NotificationItem{
			UUID:      msg.UUID,
			Source:    "message",
			Category:  msg.Category,
			Type:      msg.Type,
			Title:     msg.Title,
			Content:   msg.Content,
			ActionURL: msg.ActionURL,
			IsRead:    msg.IsRead,
			CreatedAt: msg.CreatedAt,
		})
	}

	for _, ann := range announcements {
		isRead, _ := b.ds.NtfAnnouncement().IsRead(ctx, userID, ann.ID)
		// Filter by IsRead if specified
		if req.IsRead != nil && *req.IsRead != isRead {
			continue
		}
		items = append(items, v1.NotificationItem{
			UUID:      ann.UUID,
			Source:    "announcement",
			Category:  string(model.NotificationCategorySystem),
			Title:     ann.Title,
			Content:   ann.Content,
			ActionURL: ann.ActionURL,
			IsRead:    isRead,
			CreatedAt: ann.CreatedAt,
		})
	}

	// Sort by CreatedAt (simplified - in production use proper merge sort)
	// For now, just return combined results

	return &v1.ListNotificationsResponse{
		Data:  items,
		Total: msgTotal + annTotal,
	}, nil
}

func (b *notificationBiz) UnreadCount(ctx context.Context, userID string) (*v1.UnreadCountResponse, error) {
	msgCount, err := b.ds.NtfMessage().CountUnread(ctx, userID)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("count unread messages: %v", err)
	}

	annCount, err := b.ds.NtfAnnouncement().CountUnreadForUser(ctx, userID)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("count unread announcements: %v", err)
	}

	return &v1.UnreadCountResponse{Count: msgCount + annCount}, nil
}

func (b *notificationBiz) MarkAsRead(ctx context.Context, userID string, uuid string) error {
	// Try message first
	msg, err := b.ds.NtfMessage().GetByUUID(ctx, uuid)
	if err == nil && msg.UserID == userID {
		if err := b.ds.NtfMessage().MarkAsRead(ctx, userID, uuid); err != nil {
			return errno.ErrDBWrite.WithMessage("mark message as read: %v", err)
		}
		return nil
	}

	// Try announcement
	ann, err := b.ds.NtfAnnouncement().GetByUUID(ctx, uuid)
	if err == nil {
		if err := b.ds.NtfAnnouncement().MarkAsRead(ctx, userID, ann.ID); err != nil {
			return errno.ErrDBWrite.WithMessage("mark announcement as read: %v", err)
		}
		return nil
	}

	return errno.ErrNotFound
}

func (b *notificationBiz) MarkAllAsRead(ctx context.Context, userID string) error {
	// Mark all messages as read
	if err := b.ds.NtfMessage().MarkAllAsRead(ctx, userID); err != nil {
		return errno.ErrDBWrite.WithMessage("mark all messages as read: %v", err)
	}

	// Mark all announcements as read (get all published, mark each)
	_, announcements, err := b.ds.NtfAnnouncement().ListPublished(ctx, where.New())
	if err != nil {
		return errno.ErrDBRead.WithMessage("list published announcements: %v", err)
	}

	for _, ann := range announcements {
		isRead, _ := b.ds.NtfAnnouncement().IsRead(ctx, userID, ann.ID)
		if !isRead {
			_ = b.ds.NtfAnnouncement().MarkAsRead(ctx, userID, ann.ID)
		}
	}

	return nil
}

func (b *notificationBiz) Delete(ctx context.Context, userID string, uuid string) error {
	msg, err := b.ds.NtfMessage().GetByUUID(ctx, uuid)
	if err != nil {
		return errno.ErrNotFound
	}
	if msg.UserID != userID {
		return errno.ErrPermissionDenied
	}

	if err := b.ds.NtfMessage().Delete(ctx, where.F("uuid", uuid)); err != nil {
		return errno.ErrDBWrite.WithMessage("delete message: %v", err)
	}

	return nil
}
```

**Step 2: 创建 preference biz**

```go
// ABOUTME: Business logic for notification preferences.
// ABOUTME: Handles getting and updating user notification settings.

package notification

import (
	"context"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type PreferenceBiz interface {
	Get(ctx context.Context, userID string) (*v1.GetPreferencesResponse, error)
	Update(ctx context.Context, userID string, req *v1.UpdatePreferencesRequest) error
}

type preferenceBiz struct {
	ds store.IStore
}

func NewPreference(ds store.IStore) PreferenceBiz {
	return &preferenceBiz{ds: ds}
}

func (b *preferenceBiz) Get(ctx context.Context, userID string) (*v1.GetPreferencesResponse, error) {
	pref, err := b.ds.NtfPreference().GetByUserID(ctx, userID)
	if err != nil {
		// Return default preferences if not set
		defaults := model.DefaultPreferences()
		return &v1.GetPreferencesResponse{
			Preferences: v1.NotificationPreferences{
				System:      v1.ChannelPreference{InApp: defaults.System.InApp, Email: defaults.System.Email},
				Security:    v1.ChannelPreference{InApp: defaults.Security.InApp, Email: defaults.Security.Email},
				Transaction: v1.ChannelPreference{InApp: defaults.Transaction.InApp, Email: defaults.Transaction.Email},
				Social:      v1.ChannelPreference{InApp: defaults.Social.InApp, Email: defaults.Social.Email},
			},
		}, nil
	}

	prefs := pref.GetPreferences()
	return &v1.GetPreferencesResponse{
		Preferences: v1.NotificationPreferences{
			System:      v1.ChannelPreference{InApp: prefs.System.InApp, Email: prefs.System.Email},
			Security:    v1.ChannelPreference{InApp: prefs.Security.InApp, Email: prefs.Security.Email},
			Transaction: v1.ChannelPreference{InApp: prefs.Transaction.InApp, Email: prefs.Transaction.Email},
			Social:      v1.ChannelPreference{InApp: prefs.Social.InApp, Email: prefs.Social.Email},
		},
	}, nil
}

func (b *preferenceBiz) Update(ctx context.Context, userID string, req *v1.UpdatePreferencesRequest) error {
	prefs := model.NotificationPreferences{
		System:      model.ChannelPreference{InApp: req.Preferences.System.InApp, Email: req.Preferences.System.Email},
		Security:    model.ChannelPreference{InApp: req.Preferences.Security.InApp, Email: req.Preferences.Security.Email},
		Transaction: model.ChannelPreference{InApp: req.Preferences.Transaction.InApp, Email: req.Preferences.Transaction.Email},
		Social:      model.ChannelPreference{InApp: req.Preferences.Social.InApp, Email: req.Preferences.Social.Email},
	}

	if err := b.ds.NtfPreference().Upsert(ctx, userID, prefs); err != nil {
		return errno.ErrDBWrite.WithMessage("update preferences: %v", err)
	}

	return nil
}
```

**Step 3: 修改 biz.go 添加接口**

在 `IBiz` 接口中添加：

```go
Notifications() notification.NotificationBiz
NotificationPreferences() notification.PreferenceBiz
```

添加导入和实现方法。

**Step 4: Commit**

```bash
git add internal/apiserver/biz/notification/*.go internal/apiserver/biz/biz.go
git commit -m "feat(biz): add notification biz layer for apiserver

- NotificationBiz for list, read, delete operations
- PreferenceBiz for user notification preferences"
```

---

### Task 2.3: 创建 apiserver Handler 层

**Files:**
- Create: `internal/apiserver/handler/http/notification/notification.go`
- Create: `internal/apiserver/handler/http/notification/preference.go`
- Modify: `internal/apiserver/router/api.go`

**Step 1: 创建 notification handler**

```go
// ABOUTME: HTTP handlers for notification center endpoints.
// ABOUTME: Provides list, read, and delete operations for notifications.

package notification

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/apiserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/contextx"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type NotificationHandler struct {
	biz biz.IBiz
}

func New(biz biz.IBiz) *NotificationHandler {
	return &NotificationHandler{biz: biz}
}

// List
// @Summary    List notifications
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      category   query     string  false  "Filter by category"
// @Param      is_read    query     bool    false  "Filter by read status"
// @Param      page       query     int     false  "Page number"
// @Param      page_size  query     int     false  "Page size"
// @Success    200        {object}  v1.ListNotificationsResponse
// @Failure    400        {object}  core.ErrResponse
// @Failure    500        {object}  core.ErrResponse
// @Router     /v1/notifications [GET].
func (h *NotificationHandler) List(c *gin.Context) {
	var req v1.ListNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
		return
	}

	userID := contextx.UserID(c)
	resp, err := h.biz.Notifications().List(c, userID, &req)
	core.Response(c, resp, err)
}

// UnreadCount
// @Summary    Get unread notification count
// @Security   Bearer
// @Tags       Notification
// @Produce    json
// @Success    200  {object}  v1.UnreadCountResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/unread-count [GET].
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := contextx.UserID(c)
	resp, err := h.biz.Notifications().UnreadCount(c, userID)
	core.Response(c, resp, err)
}

// MarkAsRead
// @Summary    Mark notification as read
// @Security   Bearer
// @Tags       Notification
// @Param      uuid  path  string  true  "Notification UUID"
// @Success    200
// @Failure    404  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/{uuid}/read [PUT].
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	uuid := c.Param("uuid")
	userID := contextx.UserID(c)
	err := h.biz.Notifications().MarkAsRead(c, userID, uuid)
	core.Response(c, nil, err)
}

// MarkAllAsRead
// @Summary    Mark all notifications as read
// @Security   Bearer
// @Tags       Notification
// @Success    200
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/read-all [PUT].
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := contextx.UserID(c)
	err := h.biz.Notifications().MarkAllAsRead(c, userID)
	core.Response(c, nil, err)
}

// Delete
// @Summary    Delete notification
// @Security   Bearer
// @Tags       Notification
// @Param      uuid  path  string  true  "Notification UUID"
// @Success    200
// @Failure    403  {object}  core.ErrResponse
// @Failure    404  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/{uuid} [DELETE].
func (h *NotificationHandler) Delete(c *gin.Context) {
	uuid := c.Param("uuid")
	userID := contextx.UserID(c)
	err := h.biz.Notifications().Delete(c, userID, uuid)
	core.Response(c, nil, err)
}
```

**Step 2: 创建 preference handler**

```go
// ABOUTME: HTTP handlers for notification preference endpoints.
// ABOUTME: Provides get and update operations for user notification settings.

package notification

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/apiserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/contextx"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

type PreferenceHandler struct {
	biz biz.IBiz
}

func NewPreference(biz biz.IBiz) *PreferenceHandler {
	return &PreferenceHandler{biz: biz}
}

// Get
// @Summary    Get notification preferences
// @Security   Bearer
// @Tags       Notification
// @Produce    json
// @Success    200  {object}  v1.GetPreferencesResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/preferences [GET].
func (h *PreferenceHandler) Get(c *gin.Context) {
	userID := contextx.UserID(c)
	resp, err := h.biz.NotificationPreferences().Get(c, userID)
	core.Response(c, resp, err)
}

// Update
// @Summary    Update notification preferences
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      request  body  v1.UpdatePreferencesRequest  true  "Preferences"
// @Success    200
// @Failure    400  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/preferences [PUT].
func (h *PreferenceHandler) Update(c *gin.Context) {
	var req v1.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))
		return
	}

	userID := contextx.UserID(c)
	err := h.biz.NotificationPreferences().Update(c, userID, &req)
	core.Response(c, nil, err)
}
```

**Step 3: 修改 router/api.go 注册路由**

在已有的路由注册中添加：

```go
// Notification routes
ntfHandler := notification.New(b)
prefHandler := notification.NewPreference(b)

ntf := authGroup.Group("/notifications")
{
	ntf.GET("", ntfHandler.List)
	ntf.GET("/unread-count", ntfHandler.UnreadCount)
	ntf.PUT("/:uuid/read", ntfHandler.MarkAsRead)
	ntf.PUT("/read-all", ntfHandler.MarkAllAsRead)
	ntf.DELETE("/:uuid", ntfHandler.Delete)
	ntf.GET("/preferences", prefHandler.Get)
	ntf.PUT("/preferences", prefHandler.Update)
}
```

**Step 4: 更新 Swagger 文档**

Run: `make swag`

**Step 5: 验证编译**

Run: `make build BINS=bingo-apiserver`
Expected: 编译成功

**Step 6: Commit**

```bash
git add internal/apiserver/handler/http/notification/*.go internal/apiserver/router/api.go docs/swagger/
git commit -m "feat(apiserver): add notification HTTP handlers and routes

- List, read, delete notifications
- Get and update notification preferences
- Update Swagger documentation"
```

---

## Phase 3: admserver 公告管理

实现后台公告的 CRUD 和发布功能。

### Task 3.1: 创建 admserver API 结构

**Files:**
- Create: `pkg/api/admserver/v1/announcement.go`

**Step 1: 创建公告 API 结构**

```go
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
```

**Step 2: Commit**

```bash
git add pkg/api/admserver/v1/announcement.go
git commit -m "feat(api): add announcement API types for admserver"
```

---

### Task 3.2: 创建 Asynq 任务定义

**Files:**
- Create: `internal/pkg/task/announcement.go`

**Step 1: 创建公告任务定义**

```go
// ABOUTME: Asynq task definitions for announcement operations.
// ABOUTME: Defines task types and payloads for scheduled publishing.

package task

const (
	AnnouncementPublish = "announcement:publish"
)

type AnnouncementPublishPayload struct {
	AnnouncementID uint64 `json:"announcement_id"`
}
```

**Step 2: Commit**

```bash
git add internal/pkg/task/announcement.go
git commit -m "feat(task): add announcement publish task type"
```

---

### Task 3.3: 创建 admserver Biz 层

**Files:**
- Create: `internal/admserver/biz/notification/announcement.go`
- Modify: `internal/admserver/biz/biz.go`

**Step 1: 创建 announcement biz**

```go
// ABOUTME: Business logic for announcement management.
// ABOUTME: Handles CRUD, publishing, and scheduling operations.

package notification

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"

	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/internal/pkg/task"
	v1 "github.com/bingo-project/bingo/pkg/api/admserver/v1"
	"github.com/bingo-project/bingo/pkg/store/where"
)

const (
	RedisPubSubChannel = "ntf:broadcast"
)

type AnnouncementBiz interface {
	List(ctx context.Context, req *v1.ListAnnouncementsRequest) (*v1.ListAnnouncementsResponse, error)
	Get(ctx context.Context, uuid string) (*v1.AnnouncementItem, error)
	Create(ctx context.Context, req *v1.CreateAnnouncementRequest) (*v1.AnnouncementItem, error)
	Update(ctx context.Context, uuid string, req *v1.UpdateAnnouncementRequest) error
	Delete(ctx context.Context, uuid string) error
	Publish(ctx context.Context, uuid string) error
	Schedule(ctx context.Context, uuid string, req *v1.ScheduleAnnouncementRequest) error
	Cancel(ctx context.Context, uuid string) error
}

type announcementBiz struct {
	ds store.IStore
}

func NewAnnouncement(ds store.IStore) AnnouncementBiz {
	return &announcementBiz{ds: ds}
}

func (b *announcementBiz) List(ctx context.Context, req *v1.ListAnnouncementsRequest) (*v1.ListAnnouncementsResponse, error) {
	opts := where.Page(req.Page, req.PageSize).Order("created_at DESC")
	if req.Status != "" {
		opts = opts.F("status", req.Status)
	}

	total, items, err := b.ds.NtfAnnouncement().List(ctx, opts)
	if err != nil {
		return nil, errno.ErrDBRead.WithMessage("list announcements: %v", err)
	}

	data := make([]v1.AnnouncementItem, 0, len(items))
	for _, item := range items {
		data = append(data, b.toAnnouncementItem(item))
	}

	return &v1.ListAnnouncementsResponse{Data: data, Total: total}, nil
}

func (b *announcementBiz) Get(ctx context.Context, uuid string) (*v1.AnnouncementItem, error) {
	ann, err := b.ds.NtfAnnouncement().GetByUUID(ctx, uuid)
	if err != nil {
		return nil, errno.ErrNotFound
	}

	item := b.toAnnouncementItem(ann)
	return &item, nil
}

func (b *announcementBiz) Create(ctx context.Context, req *v1.CreateAnnouncementRequest) (*v1.AnnouncementItem, error) {
	ann := &model.NtfAnnouncementM{
		Title:     req.Title,
		Content:   req.Content,
		ActionURL: req.ActionURL,
		Status:    string(model.AnnouncementStatusDraft),
		ExpiresAt: req.ExpiresAt,
	}

	if err := b.ds.NtfAnnouncement().Create(ctx, ann); err != nil {
		return nil, errno.ErrDBWrite.WithMessage("create announcement: %v", err)
	}

	item := b.toAnnouncementItem(ann)
	return &item, nil
}

func (b *announcementBiz) Update(ctx context.Context, uuid string, req *v1.UpdateAnnouncementRequest) error {
	ann, err := b.ds.NtfAnnouncement().GetByUUID(ctx, uuid)
	if err != nil {
		return errno.ErrNotFound
	}

	if ann.Status == string(model.AnnouncementStatusPublished) {
		return errno.ErrPermissionDenied.WithMessage("cannot update published announcement")
	}

	if req.Title != "" {
		ann.Title = req.Title
	}
	if req.Content != "" {
		ann.Content = req.Content
	}
	ann.ActionURL = req.ActionURL
	ann.ExpiresAt = req.ExpiresAt

	if err := b.ds.NtfAnnouncement().Update(ctx, ann, where.F("uuid", uuid)); err != nil {
		return errno.ErrDBWrite.WithMessage("update announcement: %v", err)
	}

	return nil
}

func (b *announcementBiz) Delete(ctx context.Context, uuid string) error {
	ann, err := b.ds.NtfAnnouncement().GetByUUID(ctx, uuid)
	if err != nil {
		return errno.ErrNotFound
	}

	if ann.Status != string(model.AnnouncementStatusDraft) {
		return errno.ErrPermissionDenied.WithMessage("can only delete draft announcements")
	}

	if err := b.ds.NtfAnnouncement().Delete(ctx, where.F("uuid", uuid)); err != nil {
		return errno.ErrDBWrite.WithMessage("delete announcement: %v", err)
	}

	return nil
}

func (b *announcementBiz) Publish(ctx context.Context, uuid string) error {
	ann, err := b.ds.NtfAnnouncement().GetByUUID(ctx, uuid)
	if err != nil {
		return errno.ErrNotFound
	}

	if ann.Status == string(model.AnnouncementStatusPublished) {
		return errno.ErrPermissionDenied.WithMessage("already published")
	}

	now := time.Now()
	ann.Status = string(model.AnnouncementStatusPublished)
	ann.PublishedAt = &now
	ann.ScheduledAt = nil

	if err := b.ds.NtfAnnouncement().Update(ctx, ann, where.F("uuid", uuid)); err != nil {
		return errno.ErrDBWrite.WithMessage("publish announcement: %v", err)
	}

	// Publish to Redis for real-time push
	return b.publishToRedis(ann)
}

func (b *announcementBiz) Schedule(ctx context.Context, uuid string, req *v1.ScheduleAnnouncementRequest) error {
	ann, err := b.ds.NtfAnnouncement().GetByUUID(ctx, uuid)
	if err != nil {
		return errno.ErrNotFound
	}

	if ann.Status == string(model.AnnouncementStatusPublished) {
		return errno.ErrPermissionDenied.WithMessage("cannot schedule published announcement")
	}

	ann.Status = string(model.AnnouncementStatusScheduled)
	ann.ScheduledAt = &req.ScheduledAt

	if err := b.ds.NtfAnnouncement().Update(ctx, ann, where.F("uuid", uuid)); err != nil {
		return errno.ErrDBWrite.WithMessage("schedule announcement: %v", err)
	}

	// Enqueue Asynq task
	payload, _ := json.Marshal(task.AnnouncementPublishPayload{AnnouncementID: ann.ID})
	t := asynq.NewTask(task.AnnouncementPublish, payload)
	delay := time.Until(req.ScheduledAt)
	if _, err := facade.Queue.Client.Enqueue(t, asynq.ProcessIn(delay)); err != nil {
		return errno.ErrOperationFailed.WithMessage("enqueue task: %v", err)
	}

	return nil
}

func (b *announcementBiz) Cancel(ctx context.Context, uuid string) error {
	ann, err := b.ds.NtfAnnouncement().GetByUUID(ctx, uuid)
	if err != nil {
		return errno.ErrNotFound
	}

	if ann.Status != string(model.AnnouncementStatusScheduled) {
		return errno.ErrPermissionDenied.WithMessage("can only cancel scheduled announcements")
	}

	ann.Status = string(model.AnnouncementStatusDraft)
	ann.ScheduledAt = nil

	if err := b.ds.NtfAnnouncement().Update(ctx, ann, where.F("uuid", uuid)); err != nil {
		return errno.ErrDBWrite.WithMessage("cancel announcement: %v", err)
	}

	return nil
}

func (b *announcementBiz) toAnnouncementItem(ann *model.NtfAnnouncementM) v1.AnnouncementItem {
	return v1.AnnouncementItem{
		UUID:        ann.UUID,
		Title:       ann.Title,
		Content:     ann.Content,
		ActionURL:   ann.ActionURL,
		Status:      ann.Status,
		ScheduledAt: ann.ScheduledAt,
		PublishedAt: ann.PublishedAt,
		ExpiresAt:   ann.ExpiresAt,
		CreatedAt:   ann.CreatedAt,
		UpdatedAt:   ann.UpdatedAt,
	}
}

func (b *announcementBiz) publishToRedis(ann *model.NtfAnnouncementM) error {
	msg := map[string]any{
		"method": "ntf.announcement",
		"data": map[string]any{
			"uuid":      ann.UUID,
			"title":     ann.Title,
			"content":   ann.Content,
			"actionUrl": ann.ActionURL,
		},
	}
	payload, _ := json.Marshal(msg)
	return facade.Redis.Publish(context.Background(), RedisPubSubChannel, payload).Err()
}
```

**Step 2: 修改 biz.go 添加接口**

**Step 3: Commit**

```bash
git add internal/admserver/biz/notification/*.go internal/admserver/biz/biz.go
git commit -m "feat(biz): add announcement biz layer for admserver

- CRUD operations for announcements
- Publish and schedule functionality
- Redis Pub/Sub for real-time push"
```

---

### Task 3.4: 创建 admserver Handler 层

**Files:**
- Create: `internal/admserver/handler/http/notification/announcement.go`
- Modify: `internal/admserver/router/api.go`

类似 Task 2.3 的模式，创建 Handler 和注册路由。

**Step 1: 创建 announcement handler（参考设计文档的 API 定义）**

**Step 2: 注册路由**

**Step 3: 更新 Swagger**

**Step 4: Commit**

---

### Task 3.5: 创建 scheduler Job

**Files:**
- Create: `internal/scheduler/job/announcement_publish.go`
- Modify: `internal/scheduler/job/registry.go`

**Step 1: 创建 announcement publish job**

```go
// ABOUTME: Asynq job handler for scheduled announcement publishing.
// ABOUTME: Publishes announcements at their scheduled time.

package job

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/internal/pkg/task"
	"github.com/bingo-project/bingo/pkg/store/where"
)

const RedisPubSubChannel = "ntf:broadcast"

func HandleAnnouncementPublishTask(ctx context.Context, t *asynq.Task) error {
	var payload task.AnnouncementPublishPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	// Get announcement
	ann, err := store.S.NtfAnnouncement().Get(ctx, where.F("id", payload.AnnouncementID))
	if err != nil {
		return err
	}

	// Skip if not scheduled (might have been cancelled)
	if ann.Status != string(model.AnnouncementStatusScheduled) {
		return nil
	}

	// Update status to published
	now := time.Now()
	ann.Status = string(model.AnnouncementStatusPublished)
	ann.PublishedAt = &now
	ann.ScheduledAt = nil

	if err := store.S.NtfAnnouncement().Update(ctx, ann, where.F("id", payload.AnnouncementID)); err != nil {
		return err
	}

	// Publish to Redis
	msg := map[string]any{
		"method": "ntf.announcement",
		"data": map[string]any{
			"uuid":      ann.UUID,
			"title":     ann.Title,
			"content":   ann.Content,
			"actionUrl": ann.ActionURL,
		},
	}
	msgPayload, _ := json.Marshal(msg)
	return facade.Redis.Publish(ctx, RedisPubSubChannel, msgPayload).Err()
}
```

**Step 2: 修改 registry.go 注册任务**

```go
mux.HandleFunc(task.AnnouncementPublish, HandleAnnouncementPublishTask)
```

**Step 3: Commit**

```bash
git add internal/scheduler/job/announcement_publish.go internal/scheduler/job/registry.go
git commit -m "feat(scheduler): add announcement publish job

- Handle scheduled announcement publishing
- Update status and publish to Redis"
```

---

## Phase 4: 实时推送

实现 apiserver 订阅 Redis 并推送给 WebSocket 客户端。

### Task 4.1: 创建 apiserver Subscriber

**Files:**
- Create: `internal/apiserver/subscriber/notification.go`
- Modify: `internal/apiserver/run.go`

**Step 1: 创建 notification subscriber**

```go
// ABOUTME: Redis Pub/Sub subscriber for notification push.
// ABOUTME: Subscribes to notification channels and broadcasts to WebSocket clients.

package subscriber

import (
	"context"
	"encoding/json"

	"github.com/bingo-project/websocket"
	"github.com/bingo-project/websocket/jsonrpc"
	"github.com/redis/go-redis/v9"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
)

const (
	ChannelBroadcast = "ntf:broadcast"
	ChannelUserPrefix = "ntf:user:"
)

type NotificationSubscriber struct {
	hub    *websocket.Hub
	redis  *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func NewNotificationSubscriber(hub *websocket.Hub) *NotificationSubscriber {
	ctx, cancel := context.WithCancel(context.Background())
	return &NotificationSubscriber{
		hub:    hub,
		redis:  facade.Redis,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *NotificationSubscriber) Start() {
	go s.subscribeBroadcast()
	// User-specific subscriptions would be handled per-connection
}

func (s *NotificationSubscriber) Stop() {
	s.cancel()
}

func (s *NotificationSubscriber) subscribeBroadcast() {
	pubsub := s.redis.Subscribe(s.ctx, ChannelBroadcast)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-ch:
			s.handleBroadcast(msg.Payload)
		}
	}
}

func (s *NotificationSubscriber) handleBroadcast(payload string) {
	var msg struct {
		Method string         `json:"method"`
		Data   map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(payload), &msg); err != nil {
		log.Errorw("failed to unmarshal broadcast message", "err", err)
		return
	}

	// Create JSON-RPC push message
	push := jsonrpc.NewPush(msg.Method, msg.Data)
	data, _ := json.Marshal(push)

	// Broadcast to all connected clients
	s.hub.Broadcast <- data
}
```

**Step 2: 修改 run.go 启动 subscriber**

在初始化 WebSocket Hub 后，启动 subscriber：

```go
// Start notification subscriber
subscriber := subscriber.NewNotificationSubscriber(hub)
subscriber.Start()
defer subscriber.Stop()
```

**Step 3: Commit**

```bash
git add internal/apiserver/subscriber/*.go internal/apiserver/run.go
git commit -m "feat(apiserver): add notification Redis subscriber

- Subscribe to ntf:broadcast channel
- Broadcast to all WebSocket clients using JSON-RPC format"
```

---

## Phase 5: 通知发送封装

实现业务层调用的通知发送服务。

### Task 5.1: 创建通知发送服务

**Files:**
- Create: `internal/pkg/notification/notification.go`
- Create: `internal/pkg/notification/category.go`
- Create: `internal/pkg/notification/channel.go`

**Step 1: 创建 category.go**

```go
// ABOUTME: Notification category constants.
// ABOUTME: Defines system, security, transaction, and social categories.

package notification

type Category string

const (
	CategorySystem      Category = "system"
	CategorySecurity    Category = "security"
	CategoryTransaction Category = "transaction"
	CategorySocial      Category = "social"
)
```

**Step 2: 创建 channel.go**

```go
// ABOUTME: Notification channel constants.
// ABOUTME: Defines in-app, email, SMS, and push channels.

package notification

type Channel string

const (
	ChannelInApp Channel = "in_app"
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"  // Reserved
	ChannelPush  Channel = "push" // Reserved
)
```

**Step 3: 创建 notification.go**

```go
// ABOUTME: Notification service for sending notifications to users.
// ABOUTME: Checks preferences, persists to DB, and triggers real-time push.

package notification

import (
	"context"
	"encoding/json"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
)

type Message struct {
	UserID    string
	Category  Category
	Type      string
	Title     string
	Content   string
	ActionURL string
}

// Send sends a notification to a user based on their preferences.
func Send(ctx context.Context, msg *Message) error {
	// Get user preferences
	pref, _ := store.S.NtfPreference().GetByUserID(ctx, msg.UserID)
	prefs := model.DefaultPreferences()
	if pref != nil {
		prefs = pref.GetPreferences()
	}

	// Get category preferences
	var channelPref model.ChannelPreference
	switch msg.Category {
	case CategorySystem:
		channelPref = prefs.System
	case CategorySecurity:
		channelPref = prefs.Security
	case CategoryTransaction:
		channelPref = prefs.Transaction
	case CategorySocial:
		channelPref = prefs.Social
	default:
		channelPref = model.ChannelPreference{InApp: true}
	}

	// Send via in-app channel
	if channelPref.InApp {
		if err := sendInApp(ctx, msg); err != nil {
			return err
		}
	}

	// Send via email channel (async via Asynq)
	if channelPref.Email {
		// TODO: Enqueue email task
	}

	return nil
}

func sendInApp(ctx context.Context, msg *Message) error {
	// Persist to database
	ntfMsg := &model.NtfMessageM{
		UserID:    msg.UserID,
		Category:  string(msg.Category),
		Type:      msg.Type,
		Title:     msg.Title,
		Content:   msg.Content,
		ActionURL: msg.ActionURL,
	}
	if err := store.S.NtfMessage().Create(ctx, ntfMsg); err != nil {
		return err
	}

	// Publish to Redis for real-time push
	payload := map[string]any{
		"method": "ntf.message",
		"data": map[string]any{
			"uuid":      ntfMsg.UUID,
			"category":  ntfMsg.Category,
			"type":      ntfMsg.Type,
			"title":     ntfMsg.Title,
			"content":   ntfMsg.Content,
			"actionUrl": ntfMsg.ActionURL,
		},
	}
	data, _ := json.Marshal(payload)
	channel := "ntf:user:" + msg.UserID
	return facade.Redis.Publish(ctx, channel, data).Err()
}
```

**Step 4: Commit**

```bash
git add internal/pkg/notification/*.go
git commit -m "feat(notification): add notification send service

- Check user preferences before sending
- Persist to database
- Publish to Redis for real-time push"
```

---

## Phase 6: 测试与文档

### Task 6.1: 编写单元测试

为 Store、Biz 层编写单元测试。

### Task 6.2: 编写集成测试

测试完整的通知流程：发送 → 存储 → 推送 → 列表查询。

### Task 6.3: 更新 API 文档

确保 Swagger 文档完整。

---

## 执行顺序建议

1. **Phase 1** (数据层) - 必须先完成，其他 Phase 依赖
2. **Phase 2** (apiserver) 和 **Phase 3** (admserver) - 可以并行开发
3. **Phase 4** (实时推送) - 依赖 Phase 1
4. **Phase 5** (发送服务) - 依赖 Phase 1 和 Phase 4
5. **Phase 6** (测试) - 最后进行

每个 Phase 完成后进行代码审查。
