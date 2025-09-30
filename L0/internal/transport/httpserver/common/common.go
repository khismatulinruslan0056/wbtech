package common

import "errors"

type contextKey string

const OrderIDKey = contextKey("orderID")

var (
	InvalidOrderID   = errors.New("invalid or empty order ID")
	MethodNotAllowed = errors.New("method not allowed")
)
