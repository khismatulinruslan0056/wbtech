package handlers

import (
	"L0/internal/service"
	"L0/internal/storage"
	h "L0/internal/transport/httpserver/common"
	"L0/internal/transport/httpserver/mapper"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

//go:generate go run github.com/vektra/mockery/v2 --name Getter --output ../../../mocks/transport/handlers --case underscore
type Getter interface {
	Get(context.Context, string) (*service.Order, error)
}

// GetOrder godoc
// @Summary      Get an order by ID
// @Description  Retrieves a complete order object from the cache or database using its unique ID.
// @Tags         Orders
// @Produce      json
// @Param        orderID   path      string  true  "Order Unique ID"
// @Success      200  {object}  dto.PublicOrder "Successfully retrieved order"
// @Failure      400  {object}  dto.ErrorResponse "Invalid order ID provided"
// @Failure      404  {object}  dto.ErrorResponse "Order with the specified ID was not found"
// @Failure      500  {object}  dto.ErrorResponse "An internal server error occurred"
// @Router       /order/{orderID} [get]
func GetOrder(log *slog.Logger, getter Getter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const op = "GetOrder"
		if http.MethodGet != r.Method {
			log.Error("method not allowed", "err", h.MethodNotAllowed)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		orderID, ok := r.Context().Value(h.OrderIDKey).(string)
		if !ok || orderID == "" {
			log.Error("orderID must be a string and can't be empty", "err", h.InvalidOrderID)
			http.Error(w, "orderID must be a string and can't be empty", http.StatusBadRequest)
			return
		}

		orderService, err := getter.Get(r.Context(), orderID)
		if err != nil {
			log.Error("failed to get order", "op", op, "error", err)
			if errors.Is(err, storage.ErrNotFound) {
				http.Error(w, "Order not found", http.StatusNotFound)
			} else {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		order, err := mapper.ServiceToPublicDTO(orderService)
		if err != nil {
			log.Error("failed to convert order to dto", "op", op, "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		if err = json.NewEncoder(&buf).Encode(order); err != nil {
			log.Error("failed to encode response", "op", op, "error", err)
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
	})
}
