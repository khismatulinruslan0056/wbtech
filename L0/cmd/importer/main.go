package main

import (
	"L0/internal/app/importer"
	"L0/internal/config"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustLoad("./.env")
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	app := importer.New(log, cfg)
	app.Run()

}
