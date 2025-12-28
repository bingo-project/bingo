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
	channel := RedisUserChannelPrefix + msg.UserID

	return facade.Redis.Publish(ctx, channel, data).Err()
}
