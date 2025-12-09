// ABOUTME: App is the unified entry point for bingo applications.
// ABOUTME: Coordinates Runnable lifecycle and provides dependency access.

package app

import (
	"context"
	"sync"
	"time"

	"bingo/pkg/server"

	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// App is the main application container.
type App struct {
	runnables  []Runnable
	registrars []Registrar

	ready     chan struct{}
	readyOnce sync.Once

	shutdownTimeout time.Duration

	config any
	db     *gorm.DB

	initOnce sync.Once
	initErr  error
	initFunc func() error

	healthAddr   string
	healthServer *server.HealthServer

	mu sync.Mutex
}

// New creates a new App with the given options.
// Returns error if configuration parsing fails.
func New(opts ...Option) (*App, error) {
	app := &App{
		ready:           make(chan struct{}),
		shutdownTimeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(app)
	}
	return app, nil
}

// Add adds a Runnable to the App.
// If the Runnable also implements Registrar, it will be registered.
// Returns the App for chaining.
func (app *App) Add(r Runnable) *App {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.runnables = append(app.runnables, r)

	// If it also implements Registrar, add to registrars
	if reg, ok := r.(Registrar); ok {
		app.registrars = append(app.registrars, reg)
	}

	return app
}

// Register adds a Registrar to the App.
// Returns the App for chaining.
func (app *App) Register(r Registrar) *App {
	app.mu.Lock()
	defer app.mu.Unlock()

	app.registrars = append(app.registrars, r)
	return app
}

// Run starts all runnables and blocks until ctx is canceled.
// Automatically calls Init() if not already called.
func (app *App) Run(ctx context.Context) error {
	// Auto-init if not already done
	if err := app.Init(); err != nil {
		return err
	}

	// Start health server if configured
	if app.healthAddr != "" {
		app.healthServer = server.NewHealthServer(app.healthAddr)
		app.runnables = append([]Runnable{app.healthServer}, app.runnables...)
	}

	// Phase 1: Register (serial, in order)
	for _, reg := range app.registrars {
		if err := reg.Register(app); err != nil {
			return err
		}
	}

	// Phase 2: Start all runnables (concurrent)
	if len(app.runnables) == 0 {
		app.markReady()
		<-ctx.Done()
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	for _, r := range app.runnables {
		r := r
		g.Go(func() error {
			return r.Start(ctx)
		})
	}

	// Mark ready after all runnables started
	app.markReady()

	return g.Wait()
}

// Ready returns a channel that is closed when the App is ready.
func (app *App) Ready() <-chan struct{} {
	return app.ready
}

func (app *App) markReady() {
	app.readyOnce.Do(func() {
		if app.healthServer != nil {
			app.healthServer.SetReady(true)
		}
		close(app.ready)
	})
}

// getName returns the name of a component if it implements Named.
func getName(r any) string {
	if n, ok := r.(Named); ok {
		return n.Name()
	}
	return ""
}

// Config returns the application configuration.
func (app *App) Config() any {
	return app.config
}

// DB returns the database connection.
// Panics if not configured.
func (app *App) DB() *gorm.DB {
	if app.db == nil {
		panic("database not configured")
	}
	return app.db
}

// Init initializes dependencies (DB, Cache, etc).
// Safe to call multiple times - only executes once.
func (app *App) Init() error {
	app.initOnce.Do(func() {
		if app.initFunc != nil {
			app.initErr = app.initFunc()
		}
	})
	return app.initErr
}

// Close releases resources (DB connections, etc).
// Safe to call multiple times.
func (app *App) Close() error {
	// Future: close DB, Cache connections
	return nil
}

// HealthAddr returns the actual health server address.
// Only valid after Run() is called with WithHealthAddr.
func (app *App) HealthAddr() string {
	if app.healthServer != nil {
		return app.healthServer.Addr()
	}
	return app.healthAddr
}
