package handlers

import (
	mocks "L0/internal/mocks/transport/kafka/handlers"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var validOrderJson = `{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}`

var invalidOrderJson = `{
  "order_uid": "",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WBILMTESTTRACK",
      "price": 453,
      "rid": "ab4219087a764ae0btest",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}`

var errorMapperOrderJson = `{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6test",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}`

func TestHandle(t *testing.T) {
	adder := new(mocks.Adder)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewMessageHandler(adder, log)
	t.Run("error unmarshal", func(t *testing.T) {
		msg := kafka.Message{
			Key:   []byte("foo"),
			Value: []byte("bar"),
		}

		err := h.Handle(context.Background(), msg)
		require.Error(t, err, "expected error unmarshalling message")
		adder.AssertExpectations(t)
	})
	t.Run("invalid order", func(t *testing.T) {
		msg := kafka.Message{
			Key:   []byte("foo"),
			Value: []byte(invalidOrderJson),
		}

		err := h.Handle(context.Background(), msg)
		require.Error(t, err, "expected error validate message")
		adder.AssertExpectations(t)
	})
	t.Run("error mapper", func(t *testing.T) {
		msg := kafka.Message{
			Key:   []byte("foo"),
			Value: []byte(errorMapperOrderJson),
		}

		err := h.Handle(context.Background(), msg)
		require.Error(t, err, "expected error mapper message")
		adder.AssertExpectations(t)
	})
	t.Run("error mapper", func(t *testing.T) {
		msg := kafka.Message{
			Key:   []byte("foo"),
			Value: []byte(errorMapperOrderJson),
		}

		err := h.Handle(context.Background(), msg)
		require.Error(t, err, "expected error mapper message")
		adder.AssertExpectations(t)
	})
	t.Run("error adding", func(t *testing.T) {
		msg := kafka.Message{
			Key:   []byte("b563feb7b2b84b6test"),
			Value: []byte(validOrderJson),
		}
		adder.On("Add", mock.Anything, mock.Anything).Return(mock.Anything, errors.New("error adder")).Once()

		err := h.Handle(context.Background(), msg)
		require.Error(t, err, "expected error adder message")
		adder.AssertExpectations(t)
	})
	t.Run("success adding", func(t *testing.T) {
		msg := kafka.Message{
			Key:   []byte("b563feb7b2b84b6test"),
			Value: []byte(validOrderJson),
		}
		adder.On("Add", mock.Anything, mock.Anything).Return(mock.Anything, nil).Once()
		err := h.Handle(context.Background(), msg)
		require.NoError(t, err, "expected no error")
		adder.AssertExpectations(t)
	})
}
