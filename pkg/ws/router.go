// ABOUTME: WebSocket method router with middleware support.
// ABOUTME: Provides group-based routing similar to Gin.

package ws

import (
	"sync"

	"bingo/pkg/errorsx"
	"bingo/pkg/jsonrpc"
)

// Router routes JSON-RPC methods to handlers with middleware.
type Router struct {
	mu          sync.RWMutex
	middlewares []Middleware
	handlers    map[string]*handlerEntry
}

type handlerEntry struct {
	handler     Handler
	middlewares []Middleware
	compiled    Handler // cached compiled handler chain
}

// NewRouter creates a new Router.
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]*handlerEntry),
	}
}

// Use adds global middleware that applies to all handlers.
func (r *Router) Use(middlewares ...Middleware) *Router {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middlewares = append(r.middlewares, middlewares...)
	// Invalidate compiled handlers
	for _, entry := range r.handlers {
		entry.compiled = nil
	}
	return r
}

// Handle registers a handler for a method.
func (r *Router) Handle(method string, handler Handler, middlewares ...Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[method] = &handlerEntry{
		handler:     handler,
		middlewares: middlewares,
	}
}

// Dispatch routes a request to its handler.
func (r *Router) Dispatch(mc *MiddlewareContext) *jsonrpc.Response {
	r.mu.RLock()
	entry, ok := r.handlers[mc.Method]
	if !ok {
		r.mu.RUnlock()
		return jsonrpc.NewErrorResponse(mc.Request.ID,
			errorsx.New(404, "MethodNotFound", "Method not found: %s", mc.Method))
	}

	// Compile handler chain if not cached
	if entry.compiled == nil {
		r.mu.RUnlock()
		r.mu.Lock()
		// Double-check after acquiring write lock
		if entry.compiled == nil {
			all := make([]Middleware, 0, len(r.middlewares)+len(entry.middlewares))
			all = append(all, r.middlewares...)
			all = append(all, entry.middlewares...)
			entry.compiled = Chain(all...)(entry.handler)
		}
		r.mu.Unlock()
		r.mu.RLock()
	}

	compiled := entry.compiled
	r.mu.RUnlock()

	return compiled(mc)
}

// Methods returns all registered method names.
func (r *Router) Methods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	methods := make([]string, 0, len(r.handlers))
	for method := range r.handlers {
		methods = append(methods, method)
	}
	return methods
}
