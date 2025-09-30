package storage

import (
	"L0/internal/models"
	"context"
	"database/sql"
	"errors"
)

var (
	IncorrectConfDSN         = errors.New("incorrect conf DSN")
	ErrOrderExists           = errors.New("order exists")
	ErrItemExists            = errors.New("item exists")
	ErrPaymentExists         = errors.New("payment exists")
	ErrDeliveryExists        = errors.New("delivery exists")
	IncorrectTypeDB          = errors.New("incorrect type DB")
	ConnectionDBClosed       = errors.New("connection to db closed")
	TransactionBeenCompleted = errors.New("transaction has already been completed")
	UnknownErr               = errors.New("unknown error")
	OperationCancelled       = errors.New("operation were cancelled")
	TimeoutExpired           = errors.New("timeout expired")
	ErrNotFound              = errors.New("object not found")
)

//go:generate go run github.com/vektra/mockery/v2 --name Storage --output ../mocks/storage --case underscore
type Storage interface {
	BeginTx(ctx context.Context, optsTx *sql.TxOptions) (Tx, error)
	WithTx(tx Tx) Storage

	AddOrder(ctx context.Context, order *models.Order) (string, error)
	AddItem(ctx context.Context, item *models.Item) (int64, error)
	AddItems(ctx context.Context, item []*models.Item) ([]int64, error)
	AddDelivery(ctx context.Context, delivery *models.Delivery) (int64, error)
	AddPayment(ctx context.Context, payment *models.Payment) (int64, error)

	GetOrder(ctx context.Context, OrderID string) (*models.Order, error)
	GetItemsByID(ctx context.Context, OrderID string) ([]*models.Item, error)
	GetPayment(ctx context.Context, OrderID string) (*models.Payment, error)
	GetDelivery(ctx context.Context, OrderID string) (*models.Delivery, error)

	GetAllItems(context.Context) ([]*models.Item, error)
	GetAllOrders(ctx context.Context) ([]*models.Order, error)
	GetALLDeliveries(context.Context) ([]*models.Delivery, error)
	GetAllPayments(context.Context) ([]*models.Payment, error)

	Retry(ctx context.Context, fn func(ctx context.Context) error) error
}

//go:generate go run github.com/vektra/mockery/v2 --name Tx --output ../mocks/storage --case underscore
type Tx interface {
	Commit() error
	Rollback() error
	DBTX
}

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
