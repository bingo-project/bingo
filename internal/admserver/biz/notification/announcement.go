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

var _ AnnouncementBiz = (*announcementBiz)(nil)

func NewAnnouncement(ds store.IStore) AnnouncementBiz {
	return &announcementBiz{ds: ds}
}

func (b *announcementBiz) List(ctx context.Context, req *v1.ListAnnouncementsRequest) (*v1.ListAnnouncementsResponse, error) {
	opts := where.P(req.Page, req.PageSize)
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

	if err := b.ds.NtfAnnouncement().Update(ctx, ann); err != nil {
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

	if err := b.ds.NtfAnnouncement().Update(ctx, ann); err != nil {
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

	if err := b.ds.NtfAnnouncement().Update(ctx, ann); err != nil {
		return errno.ErrDBWrite.WithMessage("schedule announcement: %v", err)
	}

	// Enqueue Asynq task with delay
	payload := task.AnnouncementPublishPayload{AnnouncementID: ann.ID}
	delay := time.Until(req.ScheduledAt)
	if _, err := task.T.Queue(ctx, task.AnnouncementPublish, payload).Dispatch(asynq.ProcessIn(delay)); err != nil {
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

	if err := b.ds.NtfAnnouncement().Update(ctx, ann); err != nil {
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
