package handlers

import (
	"L0/internal/service"
	"L0/internal/transport/dto"
	"L0/internal/transport/httpserver/mapper"
	"L0/internal/transport/validation"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type MessageHandler struct {
	adder Adder
	log   *slog.Logger
}

//go:generate go run github.com/vektra/mockery/v2 --name Adder --output ../../../mocks/transport/kafka/handlers --case underscore
type Adder interface {
	Add(context.Context, *service.Order) (string, error)
}

func NewMessageHandler(adder Adder, log *slog.Logger) *MessageHandler {
	return &MessageHandler{adder: adder, log: log}
}

type NonRetriableError struct {
	Err error
}

func (e NonRetriableError) Error() string {
	return e.Err.Error()
}

func NewNonRetriableError(err error) NonRetriableError {
	return NonRetriableError{Err: err}
}

func (h *MessageHandler) Handle(ctx context.Context, msg kafka.Message) error {
	const op = "MessageHandler.Handle"

	order := &dto.Order{}

	err := json.Unmarshal(msg.Value, order)
	if err != nil {
		h.log.Error("failed to unmarshal json", "err", err, "op", op)
		return NewNonRetriableError(fmt.Errorf("%s: %w", op, err))
	}

	if err = validation.ValidateOrder(order); err != nil {
		h.log.Error("failed to validate order", "err", err, "op", op)
		return NewNonRetriableError(fmt.Errorf("%s: %w", op, err))
	}
	orderService, err := mapper.DTOToService(order)
	if err != nil {
		h.log.Error("failed to mapper order", "err", err, "op", op)
		return NewNonRetriableError(fmt.Errorf("%s: %w", op, err))
	}
	id, err := h.adder.Add(ctx, orderService)
	if err != nil {
		h.log.Error("failed to add order", "err", err, "op", op)
		return fmt.Errorf("%s: %w", op, err)
	}

	h.log.Info("The order has been processed and added to the database.",
		"id", id)
	return nil
}
