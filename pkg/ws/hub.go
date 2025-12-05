// ABOUTME: WebSocket connection hub for managing active clients.
// ABOUTME: Handles client registration, login, unregistration, and broadcast.

package ws

import (
	"context"
	"fmt"
	"sync"
)

// Hub maintains the set of active clients and manages their lifecycle.
type Hub struct {
	// Registered clients
	clients     map[*Client]bool
	clientsLock sync.RWMutex

	// Logged-in users (key: appID_userID)
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
	Client *Client
	UserID string
	AppID  uint32
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
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
	for {
		select {
		case <-ctx.Done():
			h.shutdown()
			return

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
	h.clientsLock.Lock()
	defer h.clientsLock.Unlock()

	for client := range h.clients {
		close(client.Send)
		delete(h.clients, client)
	}
}

func (h *Hub) handleRegister(client *Client) {
	h.clientsLock.Lock()
	defer h.clientsLock.Unlock()
	h.clients[client] = true
}

func (h *Hub) handleUnregister(client *Client) {
	h.clientsLock.Lock()
	if _, ok := h.clients[client]; ok {
		close(client.Send) // Signal WritePump to exit
		delete(h.clients, client)
	}
	h.clientsLock.Unlock()

	// Also remove from users if logged in
	h.userLock.Lock()
	defer h.userLock.Unlock()
	key := userKey(client.AppID, client.UserID)
	if c, ok := h.users[key]; ok && c == client {
		delete(h.users, key)
	}
}

func (h *Hub) handleLogin(event *LoginEvent) {
	client := event.Client

	// Update client info
	client.UserID = event.UserID
	client.AppID = event.AppID

	// Add to users map
	h.userLock.Lock()
	defer h.userLock.Unlock()
	key := userKey(event.AppID, event.UserID)
	h.users[key] = client
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

// ClientCount returns the number of connected clients.
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
func (h *Hub) GetUserClient(appID uint32, userID string) *Client {
	h.userLock.RLock()
	defer h.userLock.RUnlock()
	return h.users[userKey(appID, userID)]
}

func userKey(appID uint32, userID string) string {
	return fmt.Sprintf("%d_%s", appID, userID)
}
