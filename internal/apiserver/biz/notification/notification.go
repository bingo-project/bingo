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

var _ NotificationBiz = (*notificationBiz)(nil)

func New(ds store.IStore) NotificationBiz {
	return &notificationBiz{ds: ds}
}

func (b *notificationBiz) List(ctx context.Context, userID string, req *v1.ListNotificationsRequest) (*v1.ListNotificationsResponse, error) {
	// Query personal notifications
	msgOpts := where.F("user_id", userID).P(req.Page, req.PageSize)
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
		annOpts := where.P(req.Page, req.PageSize)
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
	_, announcements, err := b.ds.NtfAnnouncement().ListPublished(ctx, where.NewWhere())
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
