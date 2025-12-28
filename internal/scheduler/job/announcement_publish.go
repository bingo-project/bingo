// ABOUTME: Asynq job handler for publishing scheduled announcements.
// ABOUTME: Updates announcement status and publishes to Redis for real-time delivery.

package job

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"

	"github.com/bingo-project/bingo/internal/pkg/facade"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/internal/pkg/task"
	"github.com/bingo-project/bingo/pkg/store/where"
)

func HandleAnnouncementPublishTask(ctx context.Context, t *asynq.Task) error {
	var payload task.AnnouncementPublishPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		log.Errorw("Failed to unmarshal announcement publish payload", "err", err)

		return err
	}

	log.C(ctx).Infow("Processing announcement publish task", "announcement_id", payload.AnnouncementID)

	// Get announcement from store
	announcement, err := store.S.NtfAnnouncement().Get(ctx, where.F("id", payload.AnnouncementID))
	if err != nil {
		log.C(ctx).Errorw("Failed to get announcement", "err", err, "id", payload.AnnouncementID)

		return err
	}

	// Skip if already published or not in scheduled status
	if announcement.Status != string(model.AnnouncementStatusScheduled) {
		log.C(ctx).Infow("Announcement not in scheduled status, skipping",
			"id", payload.AnnouncementID,
			"status", announcement.Status)

		return nil
	}

	// Update status to published
	now := time.Now()
	announcement.Status = string(model.AnnouncementStatusPublished)
	announcement.PublishedAt = &now
	if err := store.S.NtfAnnouncement().Update(ctx, announcement); err != nil {
		log.C(ctx).Errorw("Failed to update announcement status", "err", err, "id", payload.AnnouncementID)

		return err
	}

	// Publish to Redis for real-time notification
	// Use same format as biz layer immediate publish
	notifyPayload := map[string]any{
		"method": "ntf.announcement",
		"data": map[string]any{
			"uuid":      announcement.UUID,
			"title":     announcement.Title,
			"content":   announcement.Content,
			"actionUrl": announcement.ActionURL,
		},
	}

	payloadBytes, err := json.Marshal(notifyPayload)
	if err != nil {
		log.C(ctx).Errorw("Failed to marshal notification payload", "err", err)

		return err
	}

	// Use same channel as biz layer: ntf:broadcast
	if err := facade.Redis.Publish(ctx, "ntf:broadcast", string(payloadBytes)).Err(); err != nil {
		log.C(ctx).Errorw("Failed to publish announcement notification", "err", err)

		return err
	}

	log.C(ctx).Infow("Announcement published successfully", "id", payload.AnnouncementID, "uuid", announcement.UUID)

	return nil
}
