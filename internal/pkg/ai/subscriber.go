// ABOUTME: Redis Pub/Sub subscriber for AI provider reload triggers.
// ABOUTME: Listens for reload notifications to refresh provider registry.

package ai

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/bingo-project/bingo/internal/pkg/log"
)

const (
	// AIReloadChannel is the Redis Pub/Sub channel for reload triggers.
	AIReloadChannel = "ai:reload:providers"
)

// Subscriber listens for AI reload triggers via Redis Pub/Sub.
type Subscriber struct {
	redis  *redis.Client
	loader *Loader
	ctx    context.Context
	cancel context.CancelFunc
}

// NewSubscriber creates a new reload subscriber.
func NewSubscriber(redis *redis.Client, loader *Loader) *Subscriber {
	ctx, cancel := context.WithCancel(context.Background())

	return &Subscriber{
		redis:  redis,
		loader: loader,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins listening for reload messages.
// This blocks - run in a goroutine.
func (s *Subscriber) Start() {
	log.Infow("AI reload subscriber started", "channel", AIReloadChannel)

	pubsub := s.redis.Subscribe(s.ctx, AIReloadChannel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case <-s.ctx.Done():
			log.Infow("AI reload subscriber stopped")

			return
		case msg, ok := <-ch:
			if !ok {
				log.Warnw("AI reload subscriber channel closed")

				return
			}

			log.Infow("AI reload trigger received", "payload", msg.Payload)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := s.loader.Reload(ctx); err != nil {
				log.Errorw("AI reload failed", "err", err)
			} else {
				log.Infow("AI reload completed successfully")
			}
			cancel()
		}
	}
}

// Stop gracefully shuts down the subscriber.
func (s *Subscriber) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}
