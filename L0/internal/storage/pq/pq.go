package pq

import (
	"L0/internal/config"
	"L0/internal/models"
	"L0/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/lib/pq"
)

type Storage struct {
	DB  storage.DBTX
	log *slog.Logger
}

func NewStorage(conf *config.DsnPQ, log *slog.Logger) (*Storage, error) {
	const op = "NewStorage"
	dsn, err := DSN(conf)
	if err != nil {
		return nil, errHandler("", op, err)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, errHandler("", op, err)
	}
	if err = db.Ping(); err != nil {
		return nil, errHandler("", op, err)
	}

	return &Storage{DB: db, log: log}, nil
}

func (s *Storage) BeginTx(ctx context.Context, optsTx *sql.TxOptions) (storage.Tx, error) {
	const op = "Storage.BeginTx"
	stor, ok := s.DB.(*sql.DB)
	if !ok {
		return nil, errHandler("", op, storage.IncorrectTypeDB)
	}

	var tx *sql.Tx
	err := s.Retry(ctx, func(innerCtx context.Context) error {
		var txErr error
		tx, txErr = stor.BeginTx(ctx, optsTx)
		if txErr != nil {
			return errHandler("", op, txErr)
		}
		return txErr
	})

	return tx, err
}

func (s *Storage) WithTx(tx storage.Tx) storage.Storage {
	return &Storage{tx, s.log}
}

func (s *Storage) AddOrder(ctx context.Context, order *models.Order) (string, error) {
	const op = "Storage.AddOrder"

	query := `INSERT INTO orders (id, tracknumber, entry, locale, internalsignature, customerid,
                    deliveryservice, shardkey, smid, datecreated, oofshard)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	var OrderID string
	err := s.DB.QueryRowContext(ctx, query, order.OrderUID, order.TrackNumber, order.Entry, order.Locale,
		order.InternalSignature, order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID,
		order.DateCreated, order.OofShard).Scan(&OrderID)
	if err != nil {
		err = errHandler("order", op, err)
		return "", err
	}

	return OrderID, nil
}

func (s *Storage) AddItem(ctx context.Context, item *models.Item) (int64, error) {
	const op = "Storage.AddItems"

	query := `INSERT INTO items (orderid, chrtid, price, rid, name, sale,
                    size, totalprice, nmid, brand, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) returning id`

	var itemID int64
	err := s.DB.QueryRowContext(ctx, query, item.OrderUID, item.ChrtID, item.Price, item.RID,
		item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID,
		item.Brand, item.Status).Scan(&itemID)
	if err != nil {
		err = errHandler("item", op, err)
		return -1, err
	}

	return itemID, nil
}

func (s *Storage) AddDelivery(ctx context.Context, delivery *models.Delivery) (int64, error) {
	const op = "Storage.AddDelivery"

	var deliveryID int64
	query := `INSERT INTO deliveries (orderid, name, phone, zip, city, address, region, email)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	err := s.DB.QueryRowContext(ctx, query, delivery.OrderUID, delivery.Name, delivery.Phone, delivery.Zip,
		delivery.City, delivery.Address, delivery.Region, delivery.Email).Scan(&deliveryID)
	if err != nil {
		err = errHandler("delivery", op, err)
		return -1, err
	}

	return deliveryID, nil
}

func (s *Storage) AddPayment(ctx context.Context, payment *models.Payment) (int64, error) {
	const op = "Storage.AddPayment"
	var paymentID int64
	query := `INSERT INTO payments (orderid, transaction, requestid, currency, provider, amount, payment_dt,
                      bank, deliverycost, goodstotal, customfee)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	err := s.DB.QueryRowContext(ctx, query, payment.OrderUID, payment.Transaction, payment.RequestID, payment.Currency,
		payment.Provider, payment.Amount, payment.PaymentDT, payment.Bank, payment.DeliveryCost,
		payment.GoodsTotal, payment.CustomFee).Scan(&paymentID)
	if err != nil {
		err = errHandler("payment", op, err)
		return -1, err
	}

	return paymentID, nil
}

func (s *Storage) AddItems(ctx context.Context, items []*models.Item) ([]int64, error) {
	const op = "Storage.AddItems"

	query := `INSERT INTO items (orderid, chrtid, price, rid, name, sale,
                    size, totalprice, nmid, brand, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING chrtid`

	stmt, err := s.DB.PrepareContext(ctx, query)
	if err != nil {
		err = errHandler("items", op, err)
		return nil, err
	}

	itemsIDs := make([]int64, 0, len(items))
	var itemID int64

	for _, item := range items {
		err = stmt.QueryRowContext(ctx, item.OrderUID, item.ChrtID, item.Price, item.RID,
			item.Name, item.Sale, item.Size, item.TotalPrice, item.NmID,
			item.Brand, item.Status).Scan(&itemID)
		if err != nil {
			err = errHandler("item", op, err)
			return nil, err
		}

		itemsIDs = append(itemsIDs, itemID)
	}

	return itemsIDs, nil
}

func errHandler(object, op string, err error) error {
	var pqErr *pq.Error

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return fmt.Errorf("%s: %w", op, storage.ErrNotFound)

	case errors.Is(err, context.Canceled):
		return fmt.Errorf("%s: %w", op, storage.OperationCancelled)

	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("%s: %w", op, storage.TimeoutExpired)

	case errors.Is(err, sql.ErrConnDone):
		return fmt.Errorf("%s: %w", op, storage.ConnectionDBClosed)

	case errors.Is(err, sql.ErrTxDone):
		return fmt.Errorf("%s: %w", op, storage.TransactionBeenCompleted)

	case errors.As(err, &pqErr) && pqErr.Code == "23505":
		switch object {
		case "item":
			return fmt.Errorf("%s: %w", op, storage.ErrItemExists)
		case "order":
			return fmt.Errorf("%s: %w", op, storage.ErrOrderExists)
		case "payment":
			return fmt.Errorf("%s: %w", op, storage.ErrPaymentExists)
		case "delivery":
			return fmt.Errorf("%s: %w", op, storage.ErrDeliveryExists)
		}
	}
	return fmt.Errorf("%s: %w", op, storage.UnknownErr)
}

func DSN(conf *config.DsnPQ) (string, error) {
	const op = "DSN"
	if conf.Host == "" || conf.Port <= 0 || conf.User == "" || conf.Password == "" || conf.Name == "" {
		return "", errHandler("", op, storage.IncorrectConfDSN)
	}
	return fmt.Sprintf(
		"host=%v port=%v user=%s password=%s dbname=%s sslmode=disable",
		conf.Host,
		conf.Port,
		conf.User,
		conf.Password,
		conf.Name,
	), nil
}

func (s *Storage) GetOrder(ctx context.Context, OrderID string) (*models.Order, error) {
	const op = "Storage.GetOrder"

	order := &models.Order{}

	err := s.Retry(ctx, func(innerCtx context.Context) error {
		var queryErr error

		query := `SELECT * FROM orders WHERE id = $1`

		row := s.DB.QueryRowContext(innerCtx, query, OrderID)
		queryErr = row.Scan(&order.OrderUID,
			&order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID,
			&order.DateCreated, &order.OofShard)
		if queryErr != nil {
			return errHandler("order", op, queryErr)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}
	return order, nil
}

func (s *Storage) GetDelivery(ctx context.Context, OrderID string) (*models.Delivery, error) {
	const op = "Storage.GetDelivery"
	delivery := &models.Delivery{}

	err := s.Retry(ctx, func(innerCtx context.Context) error {
		var queryErr error
		query := `SELECT * FROM deliveries WHERE orderid = $1`

		queryErr = s.DB.QueryRowContext(innerCtx, query, OrderID).Scan(&delivery.ID, &delivery.OrderUID,
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address,
			&delivery.Region, &delivery.Email)
		if queryErr != nil {
			return errHandler("delivery", op, queryErr)
		}
		return queryErr
	})
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}
	return delivery, nil
}

func (s *Storage) GetPayment(ctx context.Context, OrderID string) (*models.Payment, error) {
	const op = "Storage.GetPayment"

	payment := &models.Payment{}
	err := s.Retry(ctx, func(innerCtx context.Context) error {
		var queryErr error

		query := `SELECT * FROM payments WHERE orderid = $1`

		queryErr = s.DB.QueryRowContext(innerCtx, query, OrderID).Scan(&payment.ID, &payment.OrderUID,
			&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount,
			&payment.PaymentDT, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal,
			&payment.CustomFee)
		if queryErr != nil {
			return errHandler("payment", op, queryErr)
		}

		return queryErr
	})
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}
	return payment, nil
}

func (s *Storage) GetItemsByID(ctx context.Context, OrderID string) ([]*models.Item, error) {
	const op = "Storage.GetItems"
	items := make([]*models.Item, 0)

	err := s.Retry(ctx, func(innerCtx context.Context) error {
		tmp := make([]*models.Item, 0)
		var queryErr error
		query := `SELECT * FROM items WHERE orderid = $1`
		rows, queryErr := s.DB.QueryContext(innerCtx, query, OrderID)
		if queryErr != nil {
			return errHandler("item", op, queryErr)
		}

		defer func() {
			if errD := rows.Close(); errD != nil {
				s.log.Error("problem with closing sql.rows",
					"op", op,
					"err", errD)
			}
		}()
		for rows.Next() {
			item := new(models.Item)
			queryErr = rows.Scan(&item.ID, &item.OrderUID, &item.ChrtID, &item.Price, &item.RID, &item.Name,
				&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
			if queryErr != nil {
				return errHandler("item", op, queryErr)
			}

			tmp = append(tmp, item)
		}

		if queryErr = rows.Err(); queryErr != nil {
			return errHandler("item", op, queryErr)
		}

		items = tmp
		return queryErr
	})
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%s %w", op, storage.ErrNotFound)
	}

	return items, err
}

func (s *Storage) GetAllItems(ctx context.Context) ([]*models.Item, error) {
	const op = "Storage.GetItems"
	items := make([]*models.Item, 0)

	err := s.Retry(ctx, func(innerCtx context.Context) error {
		var queryErr error
		tmp := make([]*models.Item, 0)
		query := `SELECT * FROM items`
		rows, queryErr := s.DB.QueryContext(innerCtx, query)
		if queryErr != nil {
			return errHandler("item", op, queryErr)
		}

		defer func() {
			if errD := rows.Close(); errD != nil {
				s.log.Error("problem with closing sql.rows",
					"op", op,
					"err", errD)
			}
		}()

		for rows.Next() {
			item := new(models.Item)
			queryErr = rows.Scan(&item.ID, &item.OrderUID, &item.ChrtID, &item.Price, &item.RID, &item.Name,
				&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
			if queryErr != nil {
				return errHandler("item", op, queryErr)
			}

			tmp = append(tmp, item)
		}

		if queryErr = rows.Err(); queryErr != nil {
			return errHandler("item", op, queryErr)
		}
		items = tmp
		return queryErr
	})

	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%s %w", op, storage.ErrNotFound)
	}
	return items, err
}

func (s *Storage) GetAllOrders(ctx context.Context) ([]*models.Order, error) {
	const op = "Storage.GetAllOrders"
	orders := make([]*models.Order, 0)

	err := s.Retry(ctx, func(innerCtx context.Context) error {
		var queryErr error
		tmp := make([]*models.Order, 0)
		query := `SELECT * FROM orders`
		rows, queryErr := s.DB.QueryContext(innerCtx, query)

		if queryErr != nil {
			return errHandler("order", op, queryErr)
		}

		defer func() {
			if errD := rows.Close(); errD != nil {
				s.log.Error("problem with closing sql.rows",
					"op", op,
					"err", errD)
			}
		}()

		for rows.Next() {
			order := new(models.Order)
			queryErr = rows.Scan(&order.OrderUID,
				&order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
				&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID,
				&order.DateCreated, &order.OofShard)
			if queryErr != nil {
				return errHandler("order", op, queryErr)
			}

			tmp = append(tmp, order)
		}

		if queryErr = rows.Err(); queryErr != nil {
			return errHandler("order", op, queryErr)
		}
		orders = tmp
		return queryErr
	})
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}
	if len(orders) == 0 {
		return nil, fmt.Errorf("%s %w", op, storage.ErrNotFound)
	}
	return orders, err
}

func (s *Storage) GetALLDeliveries(ctx context.Context) ([]*models.Delivery, error) {
	const op = "Storage.GetALLDeliveries"
	deliveries := make([]*models.Delivery, 0)
	err := s.Retry(ctx, func(innerCtx context.Context) error {
		var queryErr error
		tmp := make([]*models.Delivery, 0)
		query := `SELECT * FROM deliveries`
		rows, queryErr := s.DB.QueryContext(innerCtx, query)

		if queryErr != nil {
			return errHandler("delivery", op, queryErr)
		}

		defer func() {
			if errD := rows.Close(); errD != nil {
				s.log.Error("problem with closing sql.rows",
					"op", op,
					"err", errD)
			}
		}()

		for rows.Next() {
			delivery := new(models.Delivery)
			queryErr = rows.Scan(&delivery.ID, &delivery.OrderUID,
				&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address,
				&delivery.Region, &delivery.Email)
			if queryErr != nil {
				return errHandler("delivery", op, queryErr)
			}

			tmp = append(tmp, delivery)
		}

		if queryErr = rows.Err(); queryErr != nil {
			return errHandler("delivery", op, queryErr)
		}
		deliveries = tmp
		return queryErr
	})
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}
	if len(deliveries) == 0 {
		return nil, fmt.Errorf("%s %w", op, storage.ErrNotFound)
	}
	return deliveries, err
}

func (s *Storage) GetAllPayments(ctx context.Context) ([]*models.Payment, error) {
	const op = "Storage.GetALLDeliveries"
	payments := make([]*models.Payment, 0)

	err := s.Retry(ctx, func(innerCtx context.Context) error {
		var queryErr error
		tmp := make([]*models.Payment, 0)
		query := `SELECT * FROM payments`
		rows, queryErr := s.DB.QueryContext(innerCtx, query)

		if queryErr != nil {
			return errHandler("payment", op, queryErr)
		}

		defer func() {
			if errD := rows.Close(); errD != nil {
				s.log.Error("problem with closing sql.rows",
					"op", op,
					"err", errD)
			}
		}()

		for rows.Next() {
			payment := new(models.Payment)
			queryErr = rows.Scan(&payment.ID, &payment.OrderUID,
				&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount,
				&payment.PaymentDT, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal,
				&payment.CustomFee)
			if queryErr != nil {
				return errHandler("payment", op, queryErr)
			}

			tmp = append(tmp, payment)
		}

		if queryErr = rows.Err(); queryErr != nil {
			return errHandler("payment", op, queryErr)
		}
		payments = tmp
		return queryErr
	})
	if err != nil {
		return nil, fmt.Errorf("%s %w", op, err)
	}
	if len(payments) == 0 {
		return nil, fmt.Errorf("%s %w", op, storage.ErrNotFound)
	}
	return payments, err
}

func (s *Storage) Retry(ctx context.Context, fn func(ctx context.Context) error) error {
	const op = "Storage.Retry"

	operation := func() error {
		attemptCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		err := fn(attemptCtx)

		if errors.Is(err, storage.ErrNotFound) {
			return backoff.Permanent(err)
		}

		if errors.Is(err, sql.ErrNoRows) {
			return backoff.Permanent(storage.ErrNotFound)
		}

		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) {
			return backoff.Permanent(err)
		}
		return err
	}

	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = 50 * time.Millisecond
	bo.RandomizationFactor = 0.5
	bo.Multiplier = 2.0
	bo.MaxInterval = 500 * time.Millisecond
	bo.MaxElapsedTime = 2 * time.Second
	errR := backoff.Retry(operation, bo)
	if errR != nil {
		return fmt.Errorf("%s: %w", op, errR)
	}

	return nil

}

func (s *Storage) Close() error {
	const op = "storage.Close"
	stor, _ := s.DB.(*sql.DB)

	if err := stor.Close(); err != nil {
		s.log.Error("DB connection close failed", "err", err)
		return err
	}

	return nil
}
