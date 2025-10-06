package main

import (
	_ "L0/docs"
	"L0/internal/app/api"
	"L0/internal/config"
	"fmt"
	"log/slog"
	"os"
)

// @title L0 Order Service API
// @version 1.0
// @description A simple api to get order data by its ID.
// @host localhost:8081
// @BasePath /
func main() {
	cfg := config.MustLoad("./.env")
	fmt.Println(cfg)
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	application, err := api.New(log, cfg)
	if err != nil {
		log.Error("failed to init app", "err", err)
		os.Exit(1)
	}

	application.Run()

}
