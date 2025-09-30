package handlers

import (
	"L0/internal/cache/lfu"
	"L0/internal/config"
	mocks "L0/internal/mocks/transport/handlers"
	"L0/internal/service"
	"L0/internal/storage"
	"L0/internal/storage/pq"
	"L0/internal/transport/dto"
	h "L0/internal/transport/httpserver/common"
	"L0/internal/transport/httpserver/mapper"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var correctTime, _ = time.Parse(time.RFC3339, "2021-11-26T06:12:07Z")

var (
	ordersService = []*service.Order{
		{
			OrderUID:    "b563feb7b2b84b6test",
			TrackNumber: "WBILMTESTTRACK",
			Entry:       "WBIL",
			Delivery: &service.Delivery{
				Name:    "Test Testov",
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: "Ploshad Mira 15",
				Region:  "Kraiot",
				Email:   "test@gmail.com",
			},
			Payment: &service.Payment{
				Transaction:  "b563feb7b2b84b6test",
				RequestID:    "",
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1817,
				PaymentDt:    1637907727,
				Bank:         "alpha",
				DeliveryCost: 1500,
				GoodsTotal:   317,
				CustomFee:    0,
			},
			Items: []service.Item{
				{
					ChrtID:      9934930,
					TrackNumber: "WBILMTESTTRACK",
					Price:       453,
					Rid:         "ab4219087a764ae0btest",
					Name:        "Mascaras",
					Sale:        30,
					Size:        "0",
					TotalPrice:  317,
					NmID:        2389212,
					Brand:       "Vivienne Sabo",
					Status:      202,
				},
			},
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
			OrderUID:    "b563feb7b2b84b6test2",
			TrackNumber: "WBILMTESTTRACK",
			Entry:       "WBIL",
			Delivery: &service.Delivery{
				Name:    "Test2 Testov2",
				Phone:   "+9720000002",
				Zip:     "2639809",
				City:    "Mozkinburg",
				Address: "Ploshad Pira 15",
				Region:  "Daleko",
				Email:   "tes2t@gmail.com",
			},
			Payment: &service.Payment{
				Transaction:  "b563feb7b2b84b6test2",
				RequestID:    "",
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1815,
				PaymentDt:    1637907327,
				Bank:         "alpha",
				DeliveryCost: 1501,
				GoodsTotal:   313,
				CustomFee:    0,
			},
			Items:             nil,
			Locale:            "ru",
			InternalSignature: "",
			CustomerID:        "test2",
			DeliveryService:   "meest2",
			Shardkey:          "2",
			SmID:              98,
			DateCreated:       correctTime,
			OofShard:          "2",
		},
		{
			OrderUID:    "b563feb7b2b84b6test2",
			TrackNumber: "WBILMTESTTRACK",
			Entry:       "WBIL",
			Delivery: &service.Delivery{
				Name:    "Test Testov2",
				Phone:   "+9720000002",
				Zip:     "2639809",
				City:    "Kiryat Moskvin",
				Address: "Ploshad Pira 15",
				Region:  "Kayot",
				Email:   "tes2t@gmail.com",
			},
			Payment: &service.Payment{
				Transaction:  "b563feb7b2b84b6test2",
				RequestID:    "",
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1812,
				PaymentDt:    1632907727,
				Bank:         "alpha",
				DeliveryCost: 1501,
				GoodsTotal:   347,
				CustomFee:    0,
			},
			Items: []service.Item{
				{
					ChrtID:      9933930,
					TrackNumber: "WBILMTESTTRACK",
					Price:       452,
					Rid:         "ab4219087a764ae0btest2",
					Name:        "Mascarad",
					Sale:        0,
					Size:        "0",
					TotalPrice:  452,
					NmID:        2389222,
					Brand:       "Vivienne Sabor",
					Status:      202,
				},
			},
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
)

var ordersForLoadDB = []*service.Order{
	{
		OrderUID:    "b563feb7b2b84b6test1",
		TrackNumber: "WBILMTESTTRACK1",
		Entry:       "WBIL",
		Delivery: &service.Delivery{
			Name:    "Test User1",
			Phone:   "+9720000001",
			Zip:     "2639801",
			City:    "Tel Aviv",
			Address: "Street 1",
			Region:  "Central",
			Email:   "test1@gmail.com",
		},
		Payment: &service.Payment{
			Transaction:  "b563feb7b2b84b6test1",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1000,
			PaymentDt:    1637907721,
			Bank:         "alpha",
			DeliveryCost: 100,
			GoodsTotal:   900,
		},
		Items: []service.Item{
			{
				ChrtID:      100001,
				TrackNumber: "WBILMTESTTRACK1",
				Price:       900,
				Rid:         "ridtest1",
				Name:        "Shoes",
				Sale:        0,
				Size:        "42",
				TotalPrice:  900,
				NmID:        200001,
				Brand:       "Nike",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust1",
		DeliveryService: "meest",
		Shardkey:        "1",
		SmID:            101,
		DateCreated:     correctTime,
		OofShard:        "1",
	},
	{
		OrderUID:    "b563feb7b2b84b6test2",
		TrackNumber: "WBILMTESTTRACK2",
		Entry:       "WBIL",
		Delivery: &service.Delivery{
			Name:    "Test User2",
			Phone:   "+9720000002",
			Zip:     "2639802",
			City:    "Haifa",
			Address: "Street 2",
			Region:  "North",
			Email:   "test2@gmail.com",
		},
		Payment: &service.Payment{
			Transaction:  "b563feb7b2b84b6test2",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1500,
			PaymentDt:    1637907722,
			Bank:         "alpha",
			DeliveryCost: 200,
			GoodsTotal:   1300,
		},
		Items: []service.Item{
			{
				ChrtID:      100002,
				TrackNumber: "WBILMTESTTRACK2",
				Price:       1300,
				Rid:         "ridtest2",
				Name:        "Jacket",
				Sale:        10,
				Size:        "L",
				TotalPrice:  1170,
				NmID:        200002,
				Brand:       "Adidas",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust2",
		DeliveryService: "meest",
		Shardkey:        "2",
		SmID:            102,
		DateCreated:     correctTime,
		OofShard:        "2",
	},
	{
		OrderUID:    "b563feb7b2b84b6test3",
		TrackNumber: "WBILMTESTTRACK3",
		Entry:       "WBIL",
		Delivery: &service.Delivery{
			Name:    "Test User3",
			Phone:   "+9720000003",
			Zip:     "2639803",
			City:    "Jerusalem",
			Address: "Street 3",
			Region:  "South",
			Email:   "test3@gmail.com",
		},
		Payment: &service.Payment{
			Transaction:  "b563feb7b2b84b6test3",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       2000,
			PaymentDt:    1637907723,
			Bank:         "leumi",
			DeliveryCost: 250,
			GoodsTotal:   1750,
		},
		Items: []service.Item{
			{
				ChrtID:      100003,
				TrackNumber: "WBILMTESTTRACK3",
				Price:       1750,
				Rid:         "ridtest3",
				Name:        "Watch",
				Sale:        5,
				Size:        "M",
				TotalPrice:  1662,
				NmID:        200003,
				Brand:       "Casio",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust3",
		DeliveryService: "meest",
		Shardkey:        "3",
		SmID:            103,
		DateCreated:     correctTime,
		OofShard:        "3",
	},
	{
		OrderUID:    "b563feb7b2b84b6test4",
		TrackNumber: "WBILMTESTTRACK4",
		Entry:       "WBIL",
		Delivery: &service.Delivery{
			Name:    "Test User4",
			Phone:   "+9720000004",
			Zip:     "2639804",
			City:    "Eilat",
			Address: "Street 4",
			Region:  "South",
			Email:   "test4@gmail.com",
		},
		Payment: &service.Payment{
			Transaction:  "b563feb7b2b84b6test4",
			Currency:     "USD",
			Provider:     "paypal",
			Amount:       800,
			PaymentDt:    1637907724,
			Bank:         "hapoalim",
			DeliveryCost: 100,
			GoodsTotal:   700,
		},
		Items: []service.Item{
			{
				ChrtID:      100004,
				TrackNumber: "WBILMTESTTRACK4",
				Price:       700,
				Rid:         "ridtest4",
				Name:        "Bag",
				Sale:        15,
				Size:        "M",
				TotalPrice:  595,
				NmID:        200004,
				Brand:       "Puma",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust4",
		DeliveryService: "meest",
		Shardkey:        "4",
		SmID:            104,
		DateCreated:     correctTime,
		OofShard:        "4",
	},
	{
		OrderUID:    "b563feb7b2b84b6test5",
		TrackNumber: "WBILMTESTTRACK5",
		Entry:       "WBIL",
		Delivery: &service.Delivery{
			Name:    "Test User5",
			Phone:   "+9720000005",
			Zip:     "2639805",
			City:    "Nazareth",
			Address: "Street 5",
			Region:  "North",
			Email:   "test5@gmail.com",
		},
		Payment: &service.Payment{
			Transaction:  "b563feb7b2b84b6test5",
			Currency:     "USD",
			Provider:     "stripe",
			Amount:       2200,
			PaymentDt:    1637907725,
			Bank:         "discount",
			DeliveryCost: 300,
			GoodsTotal:   1900,
		},
		Items: []service.Item{
			{
				ChrtID:      100005,
				TrackNumber: "WBILMTESTTRACK5",
				Price:       1900,
				Rid:         "ridtest5",
				Name:        "Laptop",
				Sale:        20,
				Size:        "15inch",
				TotalPrice:  1520,
				NmID:        200005,
				Brand:       "Dell",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust5",
		DeliveryService: "meest",
		Shardkey:        "5",
		SmID:            105,
		DateCreated:     correctTime,
		OofShard:        "5",
	},
	{
		OrderUID:    "b563feb7b2b84b6test6",
		TrackNumber: "WBILMTESTTRACK6",
		Entry:       "WBIL",
		Delivery: &service.Delivery{
			Name:    "Test User6",
			Phone:   "+9720000006",
			Zip:     "2639806",
			City:    "Ashdod",
			Address: "Street 6",
			Region:  "South",
			Email:   "test6@gmail.com",
		},
		Payment: &service.Payment{
			Transaction:  "b563feb7b2b84b6test6",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       600,
			PaymentDt:    1637907726,
			Bank:         "alpha",
			DeliveryCost: 80,
			GoodsTotal:   520,
		},
		Items: []service.Item{
			{
				ChrtID:      100006,
				TrackNumber: "WBILMTESTTRACK6",
				Price:       520,
				Rid:         "ridtest6",
				Name:        "Headphones",
				Sale:        25,
				Size:        "Standard",
				TotalPrice:  390,
				NmID:        200006,
				Brand:       "Sony",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust6",
		DeliveryService: "meest",
		Shardkey:        "6",
		SmID:            106,
		DateCreated:     correctTime,
		OofShard:        "6",
	},
	{
		OrderUID:    "b563feb7b2b84b6test7",
		TrackNumber: "WBILMTESTTRACK7",
		Entry:       "WBIL",
		Delivery: &service.Delivery{
			Name:    "Test User7",
			Phone:   "+9720000007",
			Zip:     "2639807",
			City:    "Beer Sheva",
			Address: "Street 7",
			Region:  "South",
			Email:   "test7@gmail.com",
		},
		Payment: &service.Payment{
			Transaction:  "b563feb7b2b84b6test7",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       400,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 50,
			GoodsTotal:   350,
		},
		Items: []service.Item{
			{
				ChrtID:      100007,
				TrackNumber: "WBILMTESTTRACK7",
				Price:       350,
				Rid:         "ridtest7",
				Name:        "Mouse",
				Sale:        10,
				Size:        "Standard",
				TotalPrice:  315,
				NmID:        200007,
				Brand:       "Logitech",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust7",
		DeliveryService: "meest",
		Shardkey:        "7",
		SmID:            107,
		DateCreated:     correctTime,
		OofShard:        "7",
	},
	{
		OrderUID:    "b563feb7b2b84b6test8",
		TrackNumber: "WBILMTESTTRACK8",
		Entry:       "WBIL",
		Delivery: &service.Delivery{
			Name:    "Test User8",
			Phone:   "+9720000008",
			Zip:     "2639808",
			City:    "Acre",
			Address: "Street 8",
			Region:  "North",
			Email:   "test8@gmail.com",
		},
		Payment: &service.Payment{
			Transaction:  "b563feb7b2b84b6test8",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1200,
			PaymentDt:    1637907728,
			Bank:         "alpha",
			DeliveryCost: 150,
			GoodsTotal:   1050,
		},
		Items: []service.Item{
			{
				ChrtID:      100008,
				TrackNumber: "WBILMTESTTRACK8",
				Price:       1050,
				Rid:         "ridtest8",
				Name:        "Keyboard",
				Sale:        20,
				Size:        "Full",
				TotalPrice:  840,
				NmID:        200008,
				Brand:       "Razer",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust8",
		DeliveryService: "meest",
		Shardkey:        "8",
		SmID:            108,
		DateCreated:     correctTime,
		OofShard:        "8",
	},
	{
		OrderUID:    "b563feb7b2b84b6test9",
		TrackNumber: "WBILMTESTTRACK9",
		Entry:       "WBIL",
		Delivery: &service.Delivery{
			Name:    "Test User9",
			Phone:   "+9720000009",
			Zip:     "2639809",
			City:    "Safed",
			Address: "Street 9",
			Region:  "North",
			Email:   "test9@gmail.com",
		},
		Payment: &service.Payment{
			Transaction:  "b563feb7b2b84b6test9",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       3000,
			PaymentDt:    1637907729,
			Bank:         "alpha",
			DeliveryCost: 400,
			GoodsTotal:   2600,
		},
		Items: []service.Item{
			{
				ChrtID:      100009,
				TrackNumber: "WBILMTESTTRACK9",
				Price:       2600,
				Rid:         "ridtest9",
				Name:        "Smartphone",
				Sale:        10,
				Size:        "6inch",
				TotalPrice:  2340,
				NmID:        200009,
				Brand:       "Samsung",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust9",
		DeliveryService: "meest",
		Shardkey:        "9",
		SmID:            109,
		DateCreated:     correctTime,
		OofShard:        "9",
	},
	{
		OrderUID:    "b563feb7b2b84b6test10",
		TrackNumber: "WBILMTESTTRACK10",
		Entry:       "WBIL",
		Delivery: &service.Delivery{
			Name:    "Test User10",
			Phone:   "+9720000010",
			Zip:     "2639810",
			City:    "Tiberias",
			Address: "Street 10",
			Region:  "North",
			Email:   "test10@gmail.com",
		},
		Payment: &service.Payment{
			Transaction:  "b563feb7b2b84b6test10",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       500,
			PaymentDt:    1637907730,
			Bank:         "alpha",
			DeliveryCost: 70,
			GoodsTotal:   430,
		},
		Items: []service.Item{
			{
				ChrtID:      100010,
				TrackNumber: "WBILMTESTTRACK10",
				Price:       430,
				Rid:         "ridtest10",
				Name:        "Book",
				Sale:        5,
				Size:        "Standard",
				TotalPrice:  409,
				NmID:        200010,
				Brand:       "Penguin",
				Status:      202,
			},
		},
		Locale:          "en",
		CustomerID:      "cust10",
		DeliveryService: "meest",
		Shardkey:        "10",
		SmID:            110,
		DateCreated:     correctTime,
		OofShard:        "10",
	},
}

func setupTestDatabase(t *testing.T) (*pq.Storage, func()) {
	var (
		err         error
		pqContainer testcontainers.Container
	)

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

	db, err := pq.NewStorage(cfg, log)
	require.NoError(t, err, "failed to connect to database")

	pqDB, _ := db.DB.(*sql.DB)

	err = goose.SetDialect("postgres")
	require.NoError(t, err, "failed to set dialect")

	err = goose.Up(pqDB, filepath.Join("..", "..", "..", "..", "migrations"))
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
		if errC := rows.Close(); errC != nil {
			t.Errorf("failed to close rows")
		}
	}()

	var tables []string
	for rows.Next() {
		var table string
		err = rows.Scan(&table)
		require.NoError(t, err)
		tables = append(tables, table)
	}

	for _, table := range tables {
		_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		require.NoError(t, err)
	}
}

func LoadDB(t *testing.T) (*service.Service, func()) {
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	stor, cleanup := setupTestDatabase(t)
	cfg := &config.Cache{
		TTL:      1 * time.Second,
		Capacity: 4,
	}

	cache := lfu.NewCache(cfg)
	s := service.New(stor, cache, log)

	for _, order := range ordersForLoadDB {
		_, err := s.Add(context.Background(), order)
		if err != nil {
			t.Fatal("failed to add order", order.OrderUID, err)
		}
	}
	return s, cleanup
}

func TestGetOrder_Integration(t *testing.T) {
	expectedOrder, err := mapper.ServiceToPublicDTO(ordersForLoadDB[0])
	require.NoError(t, err, "failed to convert order")
	serviceT, cleanup := LoadDB(t)
	defer cleanup()
	start := time.Now()
	helpReq(t, serviceT, expectedOrder)
	time1 := time.Since(start)

	start = time.Now()
	helpReq(t, serviceT, expectedOrder)
	time2 := time.Since(start)
	require.Less(t, time2, time1, "time 2 should be less than time 1")
}

func helpReq(t *testing.T, serviceT *service.Service, expectedOrder *dto.PublicOrder) {
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	var ctx = context.Background()
	hh := GetOrder(log, serviceT)

	req := httptest.NewRequest(http.MethodGet, "/order/"+ordersForLoadDB[0].OrderUID, nil)

	ctx = context.WithValue(req.Context(), h.OrderIDKey, ordersForLoadDB[0].OrderUID)

	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	hh.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	order := &dto.PublicOrder{}
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &order))

	require.Equal(t, expectedOrder, order, "wrong order")
}

func TestGetOrder(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))

	testCases := []struct {
		name        string
		method      string
		ctxID       any
		wantStatus  int
		shouldCall  bool
		getReturn   *service.Order
		getError    error
		checkJSONCT bool
	}{
		{
			name:        "ok",
			method:      http.MethodGet,
			ctxID:       "b563feb7b2b84b6test",
			wantStatus:  http.StatusOK,
			shouldCall:  true,
			getReturn:   ordersService[0],
			getError:    nil,
			checkJSONCT: true,
		},
		{
			name:       "error method",
			method:     http.MethodPost,
			ctxID:      "b563feb7b2b84b6test",
			wantStatus: http.StatusMethodNotAllowed,
			shouldCall: false,
		},
		{
			name:       "not valid id (empty)",
			method:     http.MethodGet,
			ctxID:      "",
			wantStatus: http.StatusBadRequest,
			shouldCall: false,
		},
		{
			name:       "not valid id (int)",
			method:     http.MethodGet,
			ctxID:      123,
			wantStatus: http.StatusBadRequest,
			shouldCall: false,
		},
		{
			name:       "not found",
			method:     http.MethodGet,
			ctxID:      "notFound",
			wantStatus: http.StatusNotFound,
			shouldCall: true,
			getError:   storage.ErrNotFound,
		},
		{
			name:       "api error",
			method:     http.MethodGet,
			ctxID:      "err",
			wantStatus: http.StatusInternalServerError,
			shouldCall: true,
			getError:   errors.New("api error"),
		},
		{
			name:       "ServiceToDTOl error",
			method:     http.MethodGet,
			ctxID:      "err ServiceToDTO",
			wantStatus: http.StatusInternalServerError,
			getReturn:  nil,
			shouldCall: true,
			getError:   nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ctx = context.Background()
			getter := new(mocks.Getter)
			hh := GetOrder(log, getter)

			if tc.shouldCall {
				getter.On("Get", mock.Anything, mock.MatchedBy(func(id string) bool {
					s, ok := tc.ctxID.(string)
					return ok && s == id
				})).Return(tc.getReturn, tc.getError).Once()
			}
			req := httptest.NewRequest(tc.method, "/order/x", nil)

			ctx = context.WithValue(req.Context(), h.OrderIDKey, tc.ctxID)

			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			hh.ServeHTTP(rr, req)
			require.Equal(t, tc.wantStatus, rr.Code)

			if tc.shouldCall {
				getter.AssertExpectations(t)
			} else {
				getter.AssertNotCalled(t, "Get", mock.Anything, mock.Anything)
			}

			if tc.wantStatus == http.StatusOK {
				if tc.checkJSONCT {
					require.Contains(t, rr.Header().Get("Content-Type"), "application/json")
				}
				order := &dto.Order{}
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &order))
			}

		})
	}
}
