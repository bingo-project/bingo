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
	"github.com/bingo-project/bingo/internal/pkg/notification"
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
}

func (s *NotificationSubscriber) Stop() {
	s.cancel()
}

func (s *NotificationSubscriber) subscribeBroadcast() {
	pubsub := s.redis.Subscribe(s.ctx, notification.RedisBroadcastChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for {
		select {
		case <-s.ctx.Done():
			return
		case msg := <-ch:
			if msg == nil {
				continue
			}
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
	data, err := json.Marshal(push)
	if err != nil {
		log.Errorw("failed to marshal push message", "err", err)

		return
	}

	// Broadcast to all connected clients
	s.hub.Broadcast <- data
}
