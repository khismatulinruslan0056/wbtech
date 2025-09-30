package api

import (
	"L0/internal/cache/lfu"
	"L0/internal/config"
	"L0/internal/service"
	"L0/internal/storage/pq"
	"L0/internal/transport/httpserver"
	"L0/internal/transport/httpserver/router"
	"L0/internal/transport/kafka/consumer"
	"L0/internal/transport/kafka/handlers"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type App struct {
	cfg *config.Config
	log *slog.Logger

	storage    *pq.Storage
	cache      *lfu.Cache
	httpServer *httpserver.HTTPServer
	consumer   *consumer.Consumer
}

func New(log *slog.Logger, cfg *config.Config) (*App, error) {
	storage, err := pq.NewStorage(&cfg.DsnPQ, log)
	if err != nil {
		return nil, err
	}

	cache := lfu.NewCache(&cfg.Cache)
	service := service.New(storage, cache, log)

	if err = service.WarmUpCache(context.Background()); err != nil {
		log.Error("warming up cache failed", "error", err)
	}

	go cache.CheckTTL(context.Background())

	handler := handlers.NewMessageHandler(service, log)
	consumer := consumer.NewConsumer(&cfg.Kafka, handler, 100, log)

	router := router.New(service, log)
	server := httpserver.New(&cfg.HTTPServer, router, log)

	return &App{
		cfg:        cfg,
		log:        log,
		storage:    storage,
		cache:      cache,
		httpServer: server,
		consumer:   consumer,
	}, nil
}

func (a *App) Run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.log.Info("HTTP server is starting...", "addr", a.httpServer.Addr())
		if err := a.httpServer.Start(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				a.log.Error("HTTP server failed", "err", err)
				stop()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.log.Info("Kafka consumer is starting...")
		a.consumer.Run(ctx)
		a.log.Info("Kafka consumer has stopped.")
	}()

	<-ctx.Done()
	a.log.Info("Shutdown signal received, starting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.log.Error("Failed to gracefully shutdown HTTP server", "err", err)
	}

	if err := a.consumer.Close(); err != nil {
		a.log.Error("Failed to close consumer", "err", err)
	}

	if err := a.storage.Close(); err != nil {
		a.log.Error("Failed to close storage", "err", err)
	}

	wg.Wait()
	a.log.Info("Application has been gracefully shut down.")
}
