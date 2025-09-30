package pq

import (
	"L0/internal/config"
	"L0/internal/models"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pqContainer    testcontainers.Container
	correctTime, _ = time.Parse(time.RFC3339, "2021-11-26T06:12:07Z")

	ordersModel = []*models.Order{
		{
			OrderUID:          "b563feb7b2b84b6test",
			TrackNumber:       "WBILMTESTTRACK1",
			Entry:             "WBIL",
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "test",
			DeliveryService:   "meest",
			Shardkey:          "9",
			SmID:              99,
			DateCreated:       correctTime,
			OofShard:          "1",
		},
		{
			OrderUID:          "b563feb7b2b84b6test2",
			TrackNumber:       "WBILMTESTTRACK2",
			Entry:             "WBIL",
			Locale:            "en",
			InternalSignature: "",
			CustomerID:        "test1",
			DeliveryService:   "meest1",
			Shardkey:          "1",
			SmID:              99,
			DateCreated:       correctTime,
			OofShard:          "2",
		},
	}
	deliveryModel = []*models.Delivery{
		{
			OrderUID: "b563feb7b2b84b6test",
			Name:     "Test Testov",
			Phone:    "+9720000000",
			Zip:      "2639809",
			City:     "Kiryat Mozkin",
			Address:  "Ploshad Mira 15",
			Region:   "Kraiot",
			Email:    "test@gmail.com",
		},
		{
			OrderUID: "b563feb7b2b84b6test2",
			Name:     "Test Testov2",
			Phone:    "+9720000002",
			Zip:      "2639809",
			City:     "Kiryat Moskvin",
			Address:  "Ploshad Pira 15",
			Region:   "Kayot",
			Email:    "tes2t@gmail.com",
		},
	}
	paymentModel = []*models.Payment{
		{
			OrderUID:     "b563feb7b2b84b6test",
			Transaction:  "b563feb7b2b84b6test",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDT:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		{
			OrderUID:     "b563feb7b2b84b6test2",
			Transaction:  "b563feb7b2b84b6test2",
			RequestID:    "",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1812,
			PaymentDT:    1632907727,
			Bank:         "alpha",
			DeliveryCost: 1501,
			GoodsTotal:   347,
			CustomFee:    0,
		},
	}
	itemsModel = [][]*models.Item{
		{
			{
				OrderUID:   "b563feb7b2b84b6test",
				ChrtID:     9934930,
				Price:      453,
				RID:        "ab4219087a764ae0btest",
				Name:       "Mascaras",
				Sale:       30,
				Size:       "0",
				TotalPrice: 317,
				NmID:       2389212,
				Brand:      "Vivienne Sabo",
				Status:     202,
			},
		},
		{
			{
				OrderUID:   "b563feb7b2b84b6test2",
				ChrtID:     9933930,
				Price:      452,
				RID:        "ab4219087a764ae0btest2",
				Name:       "Mascarad",
				Sale:       0,
				Size:       "0",
				TotalPrice: 452,
				NmID:       2389222,
				Brand:      "Vivienne Sabor",
				Status:     202,
			},
		},
	}
)

func setupTestDatabase(t *testing.T) (*Storage, func()) {
	var err error
	ctx := context.Background()
	containerName := "reusable-postgres-tests"

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		Name:         containerName,
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second),
		),
	}

	pqContainer, err = testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
			Reuse:            true,
		})
	require.NoError(t, err, "failed to start container")

	host, err := pqContainer.Host(ctx)
	require.NoError(t, err, "failed to get host")

	mappedPort, err := pqContainer.MappedPort(ctx, "5432")
	require.NoError(t, err, "failed to get mapped port")

	cfg := &config.DsnPQ{
		Port:     mappedPort.Int(),
		User:     "testuser",
		Password: "testpass",
		Name:     "testdb",
		Host:     host,
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	db, err := NewStorage(cfg, log)

	pqDB, _ := db.DB.(*sql.DB)

	err = goose.SetDialect("postgres")
	require.NoError(t, err)

	err = goose.UpByOne(pqDB, filepath.Join("..", "..", "..", "migrations"))
	require.NoError(t, err, "failed to apply migrations")

	cleanup := func() {
		cleanupDatabase(t, pqDB)
		if err = db.Close(); err != nil {
			t.Errorf("failed to close database connection")
		}
		if err = pqContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}

	}

	return db, cleanup
}

func cleanupDatabase(t *testing.T, db *sql.DB) {
	rows, err := db.Query(`
		SELECT tablename 
		FROM pg_tables 
		WHERE schemaname = 'public' 
		AND tablename != 'goose_db_version'
	`)
	require.NoError(t, err)
	defer func() {
		if err = rows.Close(); err != nil {
			t.Errorf("failed to close rows")
		}
	}()

	var tables []string
	for rows.Next() {
		var table string
		err := rows.Scan(&table)
		require.NoError(t, err)
		tables = append(tables, table)
	}

	for _, table := range tables {
		_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		require.NoError(t, err)
	}
}

func TestBeginTx(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx := context.Background()
		tx, err := db.BeginTx(ctx, nil)
		require.NoError(t, err)
		require.NotNil(t, tx, "tx shouldn't be nil")
		defer func() {
			if err := tx.Rollback(); err != nil {
				t.Errorf("failed to rollback transaction: %s", err)
			}

		}()

		sqlTx, ok := tx.(*sql.Tx)
		require.True(t, ok)
		_, execErr := sqlTx.Exec("SELECT 1")
		require.NoError(t, execErr)

	})
	t.Run("fail", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		cleanup()

		ctx := context.Background()
		tx, err := db.BeginTx(ctx, nil)
		require.Error(t, err, "transaction should fail")
		require.Nil(t, tx, "tx should be nil")

	})
}

func TestAddOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		expectedOrderID := ordersModel[0].OrderUID
		ctx := context.Background()
		orderID, err := db.AddOrder(ctx, ordersModel[0])
		require.NoError(t, err, "failed to add order")
		require.Equal(t, expectedOrderID, orderID, "order should be same as expected")
	})
	t.Run("fail", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		cleanup()
		ctx := context.Background()
		orderID, err := db.AddOrder(ctx, ordersModel[0])
		require.Error(t, err, "order should fail")
		require.Equal(t, "", orderID, "order should be same as expected")
	})
	t.Run("exist", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		expectedOrderID := ordersModel[0].OrderUID
		ctx := context.Background()
		orderID, err := db.AddOrder(ctx, ordersModel[0])
		require.NoError(t, err, "failed to add order")
		require.Equal(t, expectedOrderID, orderID, "order should be same as expected")
		orderID, err = db.AddOrder(ctx, ordersModel[0])
		require.Error(t, err, "order should fail")
		require.Equal(t, "", orderID, "order should be same as expected")
	})
	t.Run("cancel", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		orderID, err := db.AddOrder(ctx, ordersModel[0])
		require.Error(t, err, "order should fail")
		require.Equal(t, "", orderID, "order should be same as expected")
	})
	t.Run("timeout cancel", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		time.Sleep(100 * time.Millisecond)
		orderID, err := db.AddOrder(ctx, ordersModel[0])
		require.Error(t, err, "order should fail")
		require.Equal(t, "", orderID, "order should be same as expected")
	})
}

func TestAddItem(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx := context.Background()
		_, _ = db.AddOrder(ctx, ordersModel[0])
		item := itemsModel[0][0]
		item.ID = 1
		itemID, err := db.AddItem(ctx, item)
		require.NoError(t, err, "failed to add item")
		require.Equal(t, int64(1), itemID, "item should be same as expected")
	})
	t.Run("fail", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		cleanup()
		ctx := context.Background()
		itemID, err := db.AddItem(ctx, itemsModel[0][0])
		require.Error(t, err, "item should fail")
		require.Equal(t, int64(-1), itemID, "item should be same as expected")
	})
	t.Run("exist", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx := context.Background()
		_, _ = db.AddOrder(ctx, ordersModel[0])
		item := itemsModel[0][0]
		item.ID = 1
		itemID, err := db.AddItem(ctx, item)
		require.NoError(t, err, "failed to add item")
		require.Equal(t, int64(1), itemID, "item should be same as expected")
		itemID, err = db.AddItem(ctx, item)
		require.Error(t, err, "item should fail")
		require.Equal(t, int64(-1), itemID, "item should be same as expected")
	})

}

func TestAddDelivery(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx := context.Background()
		_, _ = db.AddOrder(ctx, ordersModel[0])
		delivery := deliveryModel[0]
		delivery.ID = 1
		deliveryID, err := db.AddDelivery(ctx, delivery)
		require.NoError(t, err, "failed to add delivery")
		require.Equal(t, int64(1), deliveryID, "delivery should be same as expected")
	})
	t.Run("fail", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		cleanup()
		ctx := context.Background()
		deliveryID, err := db.AddDelivery(ctx, deliveryModel[0])
		require.Error(t, err, "delivery should fail")
		require.Equal(t, int64(-1), deliveryID, "delivery should be same as expected")
	})
	t.Run("exist", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx := context.Background()
		_, _ = db.AddOrder(ctx, ordersModel[0])
		delivery := deliveryModel[0]
		delivery.ID = 1
		deliveryID, err := db.AddDelivery(ctx, delivery)
		require.NoError(t, err, "failed to add delivery")
		require.Equal(t, int64(1), deliveryID, "delivery should be same as expected")
		deliveryID, err = db.AddDelivery(ctx, delivery)
		require.Error(t, err, "delivery should fail")
		require.Equal(t, int64(-1), deliveryID, "delivery should be same as expected")
	})
}

func TestAddPayment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx := context.Background()
		_, _ = db.AddOrder(ctx, ordersModel[0])
		payment := paymentModel[0]
		payment.ID = 1
		paymentID, err := db.AddPayment(ctx, payment)
		require.NoError(t, err, "failed to add payment")
		require.Equal(t, int64(1), paymentID, "payment should be same as expected")
	})
	t.Run("fail", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx := context.Background()
		payment := paymentModel[0]
		payment.ID = 1
		paymentID, err := db.AddPayment(ctx, payment)
		require.Error(t, err, "item should fail")
		require.Equal(t, int64(-1), paymentID, "item should be same as expected")
	})
	t.Run("exist", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx := context.Background()
		_, _ = db.AddOrder(ctx, ordersModel[0])
		payment := paymentModel[0]
		payment.ID = 1
		paymentID, err := db.AddPayment(ctx, payment)
		require.NoError(t, err, "failed to add payment")
		require.Equal(t, int64(1), paymentID, "payment should be same as expected")
		paymentID, err = db.AddPayment(ctx, payment)
		require.Error(t, err, "payment should fail")
		require.Equal(t, int64(-1), paymentID, "payment should be same as expected")
	})

}

func TestAddItems(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx := context.Background()
		_, _ = db.AddOrder(ctx, ordersModel[0])
		items := itemsModel[0]

		itemsID, err := db.AddItems(ctx, items)
		require.NoError(t, err, "failed to add items")
		require.Equal(t, []int64{int64(items[0].ChrtID)}, itemsID, "items should be same as expected")
	})
	t.Run("fail", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		cleanup()
		ctx := context.Background()
		itemsID, err := db.AddItems(ctx, itemsModel[0])
		require.Error(t, err, "items should fail")
		require.Nil(t, itemsID, "items should be same as expected")
	})
	t.Run("exist", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()
		ctx := context.Background()
		_, _ = db.AddOrder(ctx, ordersModel[0])
		items := itemsModel[0]

		itemsID, err := db.AddItems(ctx, items)
		require.NoError(t, err, "failed to add items")
		require.Equal(t, []int64{int64(items[0].ChrtID)}, itemsID, "items should be same as expected")
		itemsID, err = db.AddItems(ctx, items)
		require.Error(t, err, "items should fail")
		require.Nil(t, itemsID, "items should be same as expected")
	})

}

func TestDSN(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		expectedDSN := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"

		cfg := &config.DsnPQ{
			Port:     5432,
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
			Host:     "localhost",
		}
		dsn, err := DSN(cfg)
		require.NoError(t, err, "failed to construct dsn")
		require.Equal(t, expectedDSN, dsn, "wrong DSN")
	})
	t.Run("failed", func(t *testing.T) {
		expectedDSN := ""

		cfg := &config.DsnPQ{
			Port:     -1,
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
			Host:     "localhost",
		}
		dsn, err := DSN(cfg)
		require.Error(t, err, "expected error")
		require.Equal(t, expectedDSN, dsn, "wrong DSN")
	})

}

func TestGetOrder(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()
		orderID, err := db.AddOrder(ctx, ordersModel[0])
		require.NoError(t, err, "failed to add order1")

		_, err = db.AddOrder(ctx, ordersModel[1])
		require.NoError(t, err, "failed to add order2")

		order, err := db.GetOrder(ctx, orderID)

		require.NoError(t, err, "failed to get order")
		require.Equal(t, ordersModel[0], order, "order should be same as expected")
	})
	t.Run("not found", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()

		order, err := db.GetOrder(ctx, ordersModel[0].OrderUID)

		require.Error(t, err, "failed to get order")
		require.Nil(t, order, "order should be same as expected")
	})

}

func TestGetDelivery(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()
		expectedDelivery := deliveryModel[0]
		expectedDelivery.ID = 1

		orderID, err := db.AddOrder(ctx, ordersModel[0])
		require.NoError(t, err, "failed to add order1")
		_, err = db.AddOrder(ctx, ordersModel[1])
		require.NoError(t, err, "failed to add order2")

		_, err = db.AddDelivery(ctx, deliveryModel[0])
		require.NoError(t, err, "failed to add delivery1")

		_, err = db.AddDelivery(ctx, deliveryModel[1])
		require.NoError(t, err, "failed to add delivery2")

		delivery, err := db.GetDelivery(ctx, orderID)

		require.NoError(t, err, "failed to get delivery")
		require.Equal(t, expectedDelivery, delivery, "delivery should be same as expected")
	})
	t.Run("not found", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()

		delivery, err := db.GetDelivery(ctx, ordersModel[0].OrderUID)

		require.Error(t, err, "failed to get delivery")
		require.Nil(t, delivery, "delivery should be same as expected")
	})
}

func TestGetPayment(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()
		expectedPayment := paymentModel[0]
		expectedPayment.ID = 1

		orderID, err := db.AddOrder(ctx, ordersModel[0])
		require.NoError(t, err, "failed to add order1")
		_, err = db.AddOrder(ctx, ordersModel[1])
		require.NoError(t, err, "failed to add order2")

		_, err = db.AddPayment(ctx, paymentModel[0])
		require.NoError(t, err, "failed to add payment1")

		_, err = db.AddPayment(ctx, paymentModel[1])
		require.NoError(t, err, "failed to add payment2")

		payment, err := db.GetPayment(ctx, orderID)

		require.NoError(t, err, "failed to get payment")
		require.Equal(t, expectedPayment, payment, "payment should be same as expected")
	})
	t.Run("not found", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()

		payment, err := db.GetPayment(ctx, ordersModel[0].OrderUID)

		require.Error(t, err, "failed to get payment")
		require.Nil(t, payment, "payment should be same as expected")
	})
}

func TestGetItemsByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()
		expectedItems := itemsModel[0]
		expectedItems[0].ID = 1

		orderID, err := db.AddOrder(ctx, ordersModel[0])
		require.NoError(t, err, "failed to add order1")
		_, err = db.AddOrder(ctx, ordersModel[1])
		require.NoError(t, err, "failed to add order2")

		_, err = db.AddItems(ctx, itemsModel[0])
		require.NoError(t, err, "failed to add items1")

		_, err = db.AddItems(ctx, itemsModel[1])
		require.NoError(t, err, "failed to add items2")

		items, err := db.GetItemsByID(ctx, orderID)

		require.NoError(t, err, "failed to get items")
		require.Equal(t, expectedItems, items, "items should be same as expected")
	})
	t.Run("not found", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()

		items, err := db.GetItemsByID(ctx, ordersModel[0].OrderUID)

		require.Error(t, err, "failed to get items")
		require.Nil(t, items, "items should be same as expected")
	})
}

func TestGetAllItems(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()
		expectedItems := append(itemsModel[0], itemsModel[1]...)
		itemsModel[0][0].ID = 1
		itemsModel[1][0].ID = 2
		_, err := db.AddOrder(ctx, ordersModel[0])
		require.NoError(t, err, "failed to add order1")
		_, err = db.AddOrder(ctx, ordersModel[1])
		require.NoError(t, err, "failed to add order2")

		_, err = db.AddItems(ctx, itemsModel[0])
		require.NoError(t, err, "failed to add items1")

		_, err = db.AddItems(ctx, itemsModel[1])
		require.NoError(t, err, "failed to add items2")

		items, err := db.GetAllItems(ctx)

		require.NoError(t, err, "failed to get items")
		require.Equal(t, expectedItems, items, "items should be same as expected")
	})
	t.Run("not found", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()

		items, err := db.GetAllItems(ctx)

		require.Error(t, err, "failed to get items")
		require.Nil(t, items, "items should be same as expected")
	})
}

func TestGetAllOrders(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()
		expectedOrders := ordersModel
		_, err := db.AddOrder(ctx, ordersModel[0])
		require.NoError(t, err, "failed to add order1")
		_, err = db.AddOrder(ctx, ordersModel[1])
		require.NoError(t, err, "failed to add order2")

		orders, err := db.GetAllOrders(ctx)

		require.NoError(t, err, "failed to get orders")
		require.Equal(t, expectedOrders, orders, "orders should be same as expected")
	})
	t.Run("not found", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()

		orders, err := db.GetAllOrders(ctx)

		require.Error(t, err, "failed to get orders")
		require.Nil(t, orders, "orders should be same as expected")
	})
}

func TestGetALLDeliveries(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()
		expectedDeliveries := deliveryModel
		expectedDeliveries[0].ID = 1
		expectedDeliveries[1].ID = 2
		_, err := db.AddOrder(ctx, ordersModel[0])
		require.NoError(t, err, "failed to add order1")
		_, err = db.AddOrder(ctx, ordersModel[1])
		require.NoError(t, err, "failed to add order2")

		_, err = db.AddDelivery(ctx, deliveryModel[0])
		require.NoError(t, err, "failed to add delivery1")
		_, err = db.AddDelivery(ctx, deliveryModel[1])
		require.NoError(t, err, "failed to add delivery2")

		deliveries, err := db.GetALLDeliveries(ctx)

		require.NoError(t, err, "failed to get deliveries")
		require.Equal(t, expectedDeliveries, deliveries, "deliveries should be same as expected")
	})
	t.Run("not found", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()

		deliveries, err := db.GetALLDeliveries(ctx)

		require.Error(t, err, "failed to get deliveries")
		require.Nil(t, deliveries, "deliveries should be same as expected")
	})
}

func TestGetAllPayments(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()
		expectedPayment := paymentModel
		expectedPayment[0].ID = 1
		expectedPayment[1].ID = 2
		_, err := db.AddOrder(ctx, ordersModel[0])
		require.NoError(t, err, "failed to add order1")
		_, err = db.AddOrder(ctx, ordersModel[1])
		require.NoError(t, err, "failed to add order2")

		_, err = db.AddPayment(ctx, paymentModel[0])
		require.NoError(t, err, "failed to add payment1")
		_, err = db.AddPayment(ctx, paymentModel[1])
		require.NoError(t, err, "failed to add payment2")

		payments, err := db.GetAllPayments(ctx)

		require.NoError(t, err, "failed to get payments")
		require.Equal(t, expectedPayment, payments, "payments should be same as expected")
	})
	t.Run("not found", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		defer cleanup()

		ctx := context.Background()

		payments, err := db.GetAllPayments(ctx)

		require.Error(t, err, "failed to get payments")
		require.Nil(t, payments, "payments should be same as expected")
	})

	t.Run("error retry", func(t *testing.T) {
		db, cleanup := setupTestDatabase(t)
		cleanup()

		ctx := context.Background()
		_, err := db.AddOrder(ctx, ordersModel[0])
		require.Error(t, err, "failed to add order1")
		_, err = db.AddOrder(ctx, ordersModel[1])
		require.Error(t, err, "failed to add order2")

		_, err = db.AddPayment(ctx, paymentModel[0])
		require.Error(t, err, "failed to add payment1")
		_, err = db.AddPayment(ctx, paymentModel[1])
		require.Error(t, err, "failed to add payment2")

		payments, err := db.GetAllPayments(ctx)

		require.Error(t, err, "failed to get payments")
		require.Nil(t, payments, "payments should be same as expected")
	})

}
