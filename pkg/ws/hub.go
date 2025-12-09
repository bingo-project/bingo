// ABOUTME: WebSocket connection hub for managing active clients.
// ABOUTME: Handles client registration, login, unregistration, and broadcast.

package ws

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"bingo/pkg/jsonrpc"
)

// ClientDisconnectCallback is called when a client disconnects.
type ClientDisconnectCallback func(client *Client)

// Hub maintains the set of active clients and manages their lifecycle.
type Hub struct {
	config             *HubConfig
	logger             Logger
	onClientDisconnect ClientDisconnectCallback

	// Anonymous connections (not yet logged in)
	anonymous     map[*Client]bool
	anonymousLock sync.RWMutex

	// Authenticated connections
	clients     map[*Client]bool
	clientsLock sync.RWMutex

	// Logged-in users (key: platform_userID)
	users    map[string]*Client
	userLock sync.RWMutex

	// Clients by ID for quick lookup
	clientsByID     map[string]*Client
	clientsByIDLock sync.RWMutex

	// Topic subscriptions
	topics     map[string]map[*Client]bool
	topicsLock sync.RWMutex

	// Channels for events
	Register    chan *Client
	Unregister  chan *Client
	Login       chan *LoginEvent
	Broadcast   chan []byte
	Subscribe   chan *SubscribeEvent
	Unsubscribe chan *UnsubscribeEvent
}

// LoginEvent represents a user login event.
type LoginEvent struct {
	Client         *Client
	UserID         string
	Platform       string
	TokenExpiresAt int64
}

// SubscribeEvent represents a topic subscription event.
type SubscribeEvent struct {
	Client *Client
	Topics []string
	Result chan []string
}

// UnsubscribeEvent represents a topic unsubscription event.
type UnsubscribeEvent struct {
	Client *Client
	Topics []string
}

// HubOption is a functional option for configuring Hub.
type HubOption func(*Hub)

// WithLogger sets a custom logger for the hub.
func WithLogger(l Logger) HubOption {
	return func(h *Hub) {
		h.logger = l
	}
}

// WithClientDisconnectCallback sets a callback for client disconnect events.
// Use this to clean up resources when a client disconnects.
func WithClientDisconnectCallback(cb ClientDisconnectCallback) HubOption {
	return func(h *Hub) {
		h.onClientDisconnect = cb
	}
}

// NewHub creates a new Hub with default config.
func NewHub(opts ...HubOption) *Hub {
	return NewHubWithConfig(DefaultHubConfig(), opts...)
}

// NewHubWithConfig creates a new Hub with custom config.
func NewHubWithConfig(cfg *HubConfig, opts ...HubOption) *Hub {
	h := &Hub{
		config:      cfg,
		logger:      nopLogger{},
		anonymous:   make(map[*Client]bool),
		clients:     make(map[*Client]bool),
		users:       make(map[string]*Client),
		clientsByID: make(map[string]*Client),
		topics:      make(map[string]map[*Client]bool),
		Register:    make(chan *Client, 256),
		Unregister:  make(chan *Client, 256),
		Login:       make(chan *LoginEvent, 256),
		Broadcast:   make(chan []byte, 256),
		Subscribe:   make(chan *SubscribeEvent, 256),
		Unsubscribe: make(chan *UnsubscribeEvent, 256),
	}
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// Run starts the hub's event loop. It blocks until context is canceled.
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

		case event := <-h.Subscribe:
			subscribed := h.doSubscribe(event.Client, event.Topics)
			if event.Result != nil {
				event.Result <- subscribed
			}

		case event := <-h.Unsubscribe:
			h.doUnsubscribe(event.Client, event.Topics)
		}
	}
}

// shutdown closes all client connections on hub shutdown.
func (h *Hub) shutdown() {
	// Close anonymous connections
	h.anonymousLock.Lock()
	for client := range h.anonymous {
		h.safeCloseSend(client)
		delete(h.anonymous, client)
	}
	h.anonymousLock.Unlock()

	// Close authenticated connections
	h.clientsLock.Lock()
	for client := range h.clients {
		h.safeCloseSend(client)
		delete(h.clients, client)
	}
	h.clientsLock.Unlock()
}

// safeCloseSend closes the client's Send channel safely using sync.Once.
func (h *Hub) safeCloseSend(client *Client) {
	client.closeOnce.Do(func() {
		close(client.Send)
	})
}

func (h *Hub) handleRegister(client *Client) {
	h.anonymousLock.Lock()
	h.anonymous[client] = true
	h.anonymousLock.Unlock()

	// Track by ID
	if client.ID != "" {
		h.clientsByIDLock.Lock()
		h.clientsByID[client.ID] = client
		h.clientsByIDLock.Unlock()
	}

	h.logger.Debugw("WebSocket client connected", "addr", client.Addr, "id", client.ID)
}

func (h *Hub) handleUnregister(client *Client) {
	// Remove from clientsByID
	if client.ID != "" {
		h.clientsByIDLock.Lock()
		delete(h.clientsByID, client.ID)
		h.clientsByIDLock.Unlock()
	}

	// Call disconnect callback to cleanup resources (e.g., rate limiters)
	if h.onClientDisconnect != nil {
		h.onClientDisconnect(client)
	}

	// Remove from anonymous
	h.anonymousLock.Lock()
	if _, ok := h.anonymous[client]; ok {
		delete(h.anonymous, client)
		h.anonymousLock.Unlock()
		h.safeCloseSend(client)
		h.logger.Debugw("WebSocket anonymous client disconnected", "addr", client.Addr)

		return
	}
	h.anonymousLock.Unlock()

	// Remove from clients
	h.clientsLock.Lock()
	delete(h.clients, client)
	h.clientsLock.Unlock()

	// Remove from users if logged in
	if client.UserID != "" && client.Platform != "" {
		h.userLock.Lock()
		key := userKey(client.Platform, client.UserID)
		if c, ok := h.users[key]; ok && c == client {
			delete(h.users, key)
		}
		h.userLock.Unlock()
	}

	// Unsubscribe from all topics
	h.unsubscribeAll(client)

	h.safeCloseSend(client)
	h.logger.Infow("WebSocket client disconnected", "addr", client.Addr, "user_id", client.UserID, "platform", client.Platform)
}

func (h *Hub) unsubscribeAll(client *Client) {
	client.topicsLock.RLock()
	topics := make([]string, 0, len(client.topics))
	for topic := range client.topics {
		topics = append(topics, topic)
	}
	client.topicsLock.RUnlock()

	if len(topics) > 0 {
		h.doUnsubscribe(client, topics)
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
	now := time.Now().Unix()
	client.UserID = event.UserID
	client.Platform = event.Platform
	client.LoginTime = now
	client.TokenExpiresAt = event.TokenExpiresAt
	client.Heartbeat(now)

	// Check for existing session
	h.userLock.Lock()
	oldClient := h.users[key]
	h.users[key] = client
	h.userLock.Unlock()

	// Add to clients
	h.clientsLock.Lock()
	h.clients[client] = true
	h.clientsLock.Unlock()

	h.logger.Infow("WebSocket client logged in", "addr", client.Addr, "user_id", event.UserID, "platform", event.Platform)

	// Kick old client if exists
	if oldClient != nil && oldClient != client {
		h.kickClient(oldClient, "您的账号已在其他设备登录")
	}
}

func (h *Hub) kickClient(client *Client, reason string) {
	h.logger.Infow("WebSocket client kicked", "addr", client.Addr, "user_id", client.UserID, "platform", client.Platform, "reason", reason)

	// Send kick notification
	push := jsonrpc.NewPush("session.kicked", map[string]string{
		"reason": reason,
	})
	data, _ := json.Marshal(push)

	select {
	case client.Send <- data:
	default:
	}

	// Kick after delay
	time.AfterFunc(100*time.Millisecond, func() {
		h.Unregister <- client
	})
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

// GetClient returns a client by ID.
func (h *Hub) GetClient(clientID string) *Client {
	h.clientsByIDLock.RLock()
	defer h.clientsByIDLock.RUnlock()

	return h.clientsByID[clientID]
}

// GetClientsByUser returns all clients for a user across all platforms.
func (h *Hub) GetClientsByUser(userID string) []*Client {
	h.userLock.RLock()
	defer h.userLock.RUnlock()

	var clients []*Client
	suffix := "_" + userID
	for key, client := range h.users {
		if len(key) > len(suffix) && key[len(key)-len(suffix):] == suffix {
			clients = append(clients, client)
		}
	}

	return clients
}

// KickClient disconnects a client by ID.
func (h *Hub) KickClient(clientID string, reason string) bool {
	client := h.GetClient(clientID)
	if client == nil {
		return false
	}
	h.kickClient(client, reason)

	return true
}

// KickUser disconnects all clients for a user.
func (h *Hub) KickUser(userID string, reason string) int {
	clients := h.GetClientsByUser(userID)
	for _, client := range clients {
		h.kickClient(client, reason)
	}

	return len(clients)
}

// HubStats contains hub statistics.
type HubStats struct {
	TotalConnections      int64
	AuthenticatedConns    int64
	AnonymousConns        int64
	ConnectionsByPlatform map[string]int
}

// Stats returns current hub statistics.
func (h *Hub) Stats() *HubStats {
	h.anonymousLock.RLock()
	anonymous := int64(len(h.anonymous))
	h.anonymousLock.RUnlock()

	h.clientsLock.RLock()
	authenticated := int64(len(h.clients))
	byPlatform := make(map[string]int)
	for client := range h.clients {
		if client.Platform != "" {
			byPlatform[client.Platform]++
		}
	}
	h.clientsLock.RUnlock()

	return &HubStats{
		TotalConnections:      anonymous + authenticated,
		AuthenticatedConns:    authenticated,
		AnonymousConns:        anonymous,
		ConnectionsByPlatform: byPlatform,
	}
}

func userKey(platform, userID string) string {
	return platform + "_" + userID
}

// TopicCount returns the number of topics with subscribers.
func (h *Hub) TopicCount() int {
	h.topicsLock.RLock()
	defer h.topicsLock.RUnlock()

	return len(h.topics)
}

// PushToTopic sends a message to all subscribers of a topic.
func (h *Hub) PushToTopic(topic, method string, data any) {
	push := jsonrpc.NewPush(method, data)
	msg, err := json.Marshal(push)
	if err != nil {
		return
	}

	h.topicsLock.RLock()
	clients := h.topics[topic]
	h.topicsLock.RUnlock()

	for client := range clients {
		select {
		case client.Send <- msg:
		default:
		}
	}
}

// PushToUser sends a message to a specific user on a specific platform.
func (h *Hub) PushToUser(platform, userID, method string, data any) {
	client := h.GetUserClient(platform, userID)
	if client == nil {
		return
	}

	push := jsonrpc.NewPush(method, data)
	msg, err := json.Marshal(push)
	if err != nil {
		return
	}

	select {
	case client.Send <- msg:
	default:
	}
}

// PushToUserAllPlatforms sends a message to a user on all connected platforms.
func (h *Hub) PushToUserAllPlatforms(userID, method string, data any) {
	push := jsonrpc.NewPush(method, data)
	msg, err := json.Marshal(push)
	if err != nil {
		return
	}

	suffix := "_" + userID

	h.userLock.RLock()
	defer h.userLock.RUnlock()

	for key, client := range h.users {
		if len(key) > len(suffix) && key[len(key)-len(suffix):] == suffix {
			select {
			case client.Send <- msg:
			default:
			}
		}
	}
}

func (h *Hub) doSubscribe(client *Client, topics []string) []string {
	h.topicsLock.Lock()
	defer h.topicsLock.Unlock()

	var subscribed []string
	for _, topic := range topics {
		if h.topics[topic] == nil {
			h.topics[topic] = make(map[*Client]bool)
		}
		h.topics[topic][client] = true

		client.topicsLock.Lock()
		if client.topics == nil {
			client.topics = make(map[string]bool)
		}
		client.topics[topic] = true
		client.topicsLock.Unlock()

		subscribed = append(subscribed, topic)
	}

	h.logger.Debugw("WebSocket client subscribed", "addr", client.Addr, "user_id", client.UserID, "topics", subscribed)

	return subscribed
}

func (h *Hub) doUnsubscribe(client *Client, topics []string) {
	h.topicsLock.Lock()
	defer h.topicsLock.Unlock()

	for _, topic := range topics {
		if clients, ok := h.topics[topic]; ok {
			delete(clients, client)
			if len(clients) == 0 {
				delete(h.topics, topic)
			}
		}

		client.topicsLock.Lock()
		delete(client.topics, topic)
		client.topicsLock.Unlock()
	}

	h.logger.Debugw("WebSocket client unsubscribed", "addr", client.Addr, "user_id", client.UserID, "topics", topics)
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
	var expired []*Client

	for client := range h.clients {
		// Check heartbeat timeout
		if client.HeartbeatTime+heartbeatTimeout <= now {
			inactive = append(inactive, client)

			continue
		}

		// Check token expiration
		if client.TokenExpiresAt > 0 && client.TokenExpiresAt <= now {
			expired = append(expired, client)
		}
	}
	h.clientsLock.RUnlock()

	// Kick inactive clients
	for _, client := range inactive {
		h.Unregister <- client
	}

	// Notify and kick expired clients
	for _, client := range expired {
		h.expireClient(client)
	}
}

func (h *Hub) expireClient(client *Client) {
	// Remove from clients map first to prevent duplicate expiration
	h.clientsLock.Lock()
	if _, ok := h.clients[client]; !ok {
		h.clientsLock.Unlock()

		return // Already removed
	}
	delete(h.clients, client)
	h.clientsLock.Unlock()

	// Remove from users map
	if client.UserID != "" && client.Platform != "" {
		h.userLock.Lock()
		key := userKey(client.Platform, client.UserID)
		if c, ok := h.users[key]; ok && c == client {
			delete(h.users, key)
		}
		h.userLock.Unlock()
	}

	h.logger.Infow("WebSocket client token expired", "addr", client.Addr, "user_id", client.UserID, "platform", client.Platform)

	push := jsonrpc.NewPush("session.expired", map[string]string{
		"reason": "Token 已过期，请重新登录",
	})
	data, _ := json.Marshal(push)

	select {
	case client.Send <- data:
	default:
	}

	time.AfterFunc(100*time.Millisecond, func() {
		h.safeCloseSend(client)
	})
}
