// ABOUTME: WebSocket connection hub for managing active clients.
// ABOUTME: Handles client registration, login, unregistration, and broadcast.

package ws

import (
	"context"
	"sync"
	"time"
)

// Hub maintains the set of active clients and manages their lifecycle.
type Hub struct {
	config *HubConfig

	// Anonymous connections (not yet logged in)
	anonymous     map[*Client]bool
	anonymousLock sync.RWMutex

	// Authenticated connections
	clients     map[*Client]bool
	clientsLock sync.RWMutex

	// Logged-in users (key: platform_userID)
	users    map[string]*Client
	userLock sync.RWMutex

	// Channels for events
	Register   chan *Client
	Unregister chan *Client
	Login      chan *LoginEvent
	Broadcast  chan []byte
}

// LoginEvent represents a user login event.
type LoginEvent struct {
	Client         *Client
	UserID         string
	Platform       string
	TokenExpiresAt int64
}

// NewHub creates a new Hub with default config.
func NewHub() *Hub {
	return NewHubWithConfig(DefaultHubConfig())
}

// NewHubWithConfig creates a new Hub with custom config.
func NewHubWithConfig(cfg *HubConfig) *Hub {
	return &Hub{
		config:     cfg,
		anonymous:  make(map[*Client]bool),
		clients:    make(map[*Client]bool),
		users:      make(map[string]*Client),
		Register:   make(chan *Client, 256),
		Unregister: make(chan *Client, 256),
		Login:      make(chan *LoginEvent, 256),
		Broadcast:  make(chan []byte, 256),
	}
}

// Run starts the hub's event loop. It blocks until context is cancelled.
func (h *Hub) Run(ctx context.Context) {
	anonymousTicker := time.NewTicker(h.config.AnonymousCleanup)
	heartbeatTicker := time.NewTicker(h.config.HeartbeatCleanup)
	defer anonymousTicker.Stop()
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

		case <-anonymousTicker.C:
			h.cleanupAnonymous()

		case <-heartbeatTicker.C:
			h.cleanupInactiveClients()

		case client := <-h.Register:
			h.handleRegister(client)

		case client := <-h.Unregister:
			h.handleUnregister(client)

		case event := <-h.Login:
			h.handleLogin(event)

		case message := <-h.Broadcast:
			h.handleBroadcast(message)
		}
	}
}

// shutdown closes all client connections on hub shutdown.
func (h *Hub) shutdown() {
	// Close anonymous connections
	h.anonymousLock.Lock()
	for client := range h.anonymous {
		close(client.Send)
		delete(h.anonymous, client)
	}
	h.anonymousLock.Unlock()

	// Close authenticated connections
	h.clientsLock.Lock()
	for client := range h.clients {
		close(client.Send)
		delete(h.clients, client)
	}
	h.clientsLock.Unlock()
}

func (h *Hub) handleRegister(client *Client) {
	h.anonymousLock.Lock()
	defer h.anonymousLock.Unlock()
	h.anonymous[client] = true
}

func (h *Hub) handleUnregister(client *Client) {
	// Remove from anonymous
	h.anonymousLock.Lock()
	if _, ok := h.anonymous[client]; ok {
		close(client.Send)
		delete(h.anonymous, client)
		h.anonymousLock.Unlock()
		return
	}
	h.anonymousLock.Unlock()

	// Remove from clients
	h.clientsLock.Lock()
	if _, ok := h.clients[client]; ok {
		close(client.Send)
		delete(h.clients, client)
	}
	h.clientsLock.Unlock()

	// Also remove from users if logged in
	if client.UserID != "" && client.Platform != "" {
		h.userLock.Lock()
		key := userKey(client.Platform, client.UserID)
		if c, ok := h.users[key]; ok && c == client {
			delete(h.users, key)
		}
		h.userLock.Unlock()
	}
}

func (h *Hub) handleLogin(event *LoginEvent) {
	client := event.Client
	key := userKey(event.Platform, event.UserID)

	// Remove from anonymous
	h.anonymousLock.Lock()
	delete(h.anonymous, client)
	h.anonymousLock.Unlock()

	// Update client info
	client.UserID = event.UserID
	client.Platform = event.Platform
	client.TokenExpiresAt = event.TokenExpiresAt

	// Add to users map
	h.userLock.Lock()
	h.users[key] = client
	h.userLock.Unlock()

	// Add to clients
	h.clientsLock.Lock()
	h.clients[client] = true
	h.clientsLock.Unlock()
}

func (h *Hub) handleBroadcast(message []byte) {
	h.clientsLock.RLock()
	defer h.clientsLock.RUnlock()

	for client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// Client buffer full, skip
		}
	}
}

// AnonymousCount returns the number of anonymous connections.
func (h *Hub) AnonymousCount() int {
	h.anonymousLock.RLock()
	defer h.anonymousLock.RUnlock()
	return len(h.anonymous)
}

// ClientCount returns the number of authenticated clients.
func (h *Hub) ClientCount() int {
	h.clientsLock.RLock()
	defer h.clientsLock.RUnlock()
	return len(h.clients)
}

// UserCount returns the number of logged-in users.
func (h *Hub) UserCount() int {
	h.userLock.RLock()
	defer h.userLock.RUnlock()
	return len(h.users)
}

// GetUserClient returns the client for a user.
func (h *Hub) GetUserClient(platform, userID string) *Client {
	h.userLock.RLock()
	defer h.userLock.RUnlock()
	return h.users[userKey(platform, userID)]
}

func userKey(platform, userID string) string {
	return platform + "_" + userID
}

func (h *Hub) cleanupAnonymous() {
	now := time.Now().Unix()
	timeout := int64(h.config.AnonymousTimeout.Seconds())

	h.anonymousLock.RLock()
	var inactive []*Client
	for client := range h.anonymous {
		if client.FirstTime+timeout <= now {
			inactive = append(inactive, client)
		}
	}
	h.anonymousLock.RUnlock()

	for _, client := range inactive {
		h.Unregister <- client
	}
}

func (h *Hub) cleanupInactiveClients() {
	now := time.Now().Unix()
	heartbeatTimeout := int64(h.config.HeartbeatTimeout.Seconds())

	h.clientsLock.RLock()
	var inactive []*Client
	for client := range h.clients {
		if client.HeartbeatTime+heartbeatTimeout <= now {
			inactive = append(inactive, client)
		}
	}
	h.clientsLock.RUnlock()

	for _, client := range inactive {
		h.Unregister <- client
	}
}
