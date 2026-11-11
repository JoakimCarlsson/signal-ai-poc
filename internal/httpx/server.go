package httpx

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/JoakimCarlsson/signal-ai-poc/internal/config"
	"github.com/JoakimCarlsson/signal-ai-poc/internal/otel"
	"github.com/JoakimCarlsson/signal-ai-poc/internal/shop"
	"github.com/joakimcarlsson/minmux/openapi"
	"github.com/joakimcarlsson/minmux/router"
)

// apiPrefix is where every API route lives.
const apiPrefix = "/api/v1"

// Config is how the server is set up: where it listens.
type Config struct {
	// Addr is the address to listen on.
	Addr string
}

// Deps are the subsystems the handlers work through. All are required.
type Deps struct {
	Shop *shop.Store
}

// Server owns the minmux router and the underlying http.Server.
type Server struct {
	addr   string
	router *router.Router
	api    *router.Group
	docs   *openapi.Generator
	shop   *shop.Store
}

// New constructs a Server with all routes registered.
func New(cfg Config, deps Deps) *Server {
	s := &Server{
		addr:   cfg.Addr,
		router: router.New(),
		shop:   deps.Shop,
	}
	s.docs = openapi.NewGenerator(openapi.Info{
		Title:       "Shop",
		Description: "The dummy shop API signal-ai watches.",
	})
	s.docs.SecuritySchemes = map[string]*openapi.SecurityScheme{}
	s.router.Use(recoverPanics)
	s.api = s.router.Group(apiPrefix)
	s.routes()
	return s
}

// routes registers every route on the router.
func (s *Server) routes() {
	s.registerHealth()
	s.registerProducts()
	s.registerOrders()
	s.registerCustomers()
	s.registerDocs()
}

// Start runs the server until ctx is cancelled, then drains gracefully. It
// returns the error that stopped the listener, or the shutdown error.
func (s *Server) Start(ctx context.Context) error {
	srv := &http.Server{
		Addr:              s.addr,
		Handler:           otel.Tracing(otel.Logging(s.router)),
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		ReadTimeout:       config.ReadTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
	}

	errc := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "api listening", "addr", s.addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(
			context.Background(),
			config.ShutdownTimeout,
		)
		defer cancel()
		return srv.Shutdown(shutCtx)
	}
}
