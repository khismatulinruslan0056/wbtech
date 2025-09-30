package importer

import (
	"L0/internal/config"
	"L0/internal/importer"
	"context"
	"log/slog"
	"math/rand"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	log      *slog.Logger
	producer *importer.Producer
	service  *importer.ImporterService
}

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewSource(time.Now().Unix()))
}

func New(log *slog.Logger, cfg *config.Config) *App {
	tpl := importer.NewTemplate()
	generator := importer.NewGenerator(rnd, tpl)
	producer := importer.NewProducer(&cfg.Kafka, log)
	service := importer.NewImportService(producer, log, generator)

	return &App{
		log:      log,
		producer: producer,
		service:  service,
	}
}

func (a *App) Run() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	defer func() {
		if err := a.producer.Close(); err != nil {
			a.log.Error("Failed to close producer cleanly", "error", err)
		}
	}()

	if err := a.service.Run(ctx); err != nil {
		a.log.Error("Importer api run failed", "error", err)
		return
	}

	if ctx.Err() != nil {
		a.log.Info("Shutdown signal received, importer stopped gracefully.")
	} else {
		a.log.Info("Importer finished its job successfully.")
	}
}
