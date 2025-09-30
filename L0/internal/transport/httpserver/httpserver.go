package httpserver

import (
	"L0/internal/config"
	"L0/internal/transport/httpserver/router"
	"context"
	"errors"
	"log/slog"
	"net/http"
)

type HTTPServer struct {
	server *http.Server
	log    *slog.Logger
}

func New(cfg *config.HTTPServer, router *router.Router, log *slog.Logger) *HTTPServer {
	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	log.Info("httpserver: initialized", "address", cfg.Address, "timeout", cfg.Timeout, "timeout", cfg.IdleTimeout)

	return &HTTPServer{server: srv, log: log}
}

func (s *HTTPServer) Start() error {
	s.log.Info("starting server", "addr", s.server.Addr)
	err := s.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.log.Error("server failed to start", "err", err)
		return err
	}
	s.log.Info("server stopped gracefully")
	return nil
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	s.log.Info("shutting down server...")
	err := s.server.Shutdown(ctx)
	if err != nil {
		s.log.Error("shutdown error", "err", err)
	} else {
		s.log.Info("shutdown completed successfully")
	}
	return err
}

func (s *HTTPServer) Addr() string {
	return s.server.Addr
}
