package lfu

import (
	"L0/internal/cache"
	"L0/internal/config"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var correctTime, _ = time.Parse(time.RFC3339, "2021-11-26T06:12:07Z")
var orders = []*cache.Order{
	{
		OrderUID:    "b563feb7b2b84b6test1",
		TrackNumber: "WBILMTESTTRACK1",
		Entry:       "WBIL",
		Delivery: &cache.Delivery{
			Name:    "Test User1",
			Phone:   "+9720000001",
			Zip:     "2639801",
			City:    "Tel Aviv",
			Address: "Street 1",
			Region:  "Central",
			Email:   "test1@gmail.com",
		},
		Payment: &cache.Payment{
			Transaction:  "b563feb7b2b84b6test1",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1000,
			PaymentDt:    1637907721,
			Bank:         "alpha",
			DeliveryCost: 100,
			GoodsTotal:   900,
		},
		Items: []cache.Item{
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
		Delivery: &cache.Delivery{
			Name:    "Test User2",
			Phone:   "+9720000002",
			Zip:     "2639802",
			City:    "Haifa",
			Address: "Street 2",
			Region:  "North",
			Email:   "test2@gmail.com",
		},
		Payment: &cache.Payment{
			Transaction:  "b563feb7b2b84b6test2",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1500,
			PaymentDt:    1637907722,
			Bank:         "alpha",
			DeliveryCost: 200,
			GoodsTotal:   1300,
		},
		Items: []cache.Item{
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
		Delivery: &cache.Delivery{
			Name:    "Test User3",
			Phone:   "+9720000003",
			Zip:     "2639803",
			City:    "Jerusalem",
			Address: "Street 3",
			Region:  "South",
			Email:   "test3@gmail.com",
		},
		Payment: &cache.Payment{
			Transaction:  "b563feb7b2b84b6test3",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       2000,
			PaymentDt:    1637907723,
			Bank:         "leumi",
			DeliveryCost: 250,
			GoodsTotal:   1750,
		},
		Items: []cache.Item{
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
		Delivery: &cache.Delivery{
			Name:    "Test User4",
			Phone:   "+9720000004",
			Zip:     "2639804",
			City:    "Eilat",
			Address: "Street 4",
			Region:  "South",
			Email:   "test4@gmail.com",
		},
		Payment: &cache.Payment{
			Transaction:  "b563feb7b2b84b6test4",
			Currency:     "USD",
			Provider:     "paypal",
			Amount:       800,
			PaymentDt:    1637907724,
			Bank:         "hapoalim",
			DeliveryCost: 100,
			GoodsTotal:   700,
		},
		Items: []cache.Item{
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
		Delivery: &cache.Delivery{
			Name:    "Test User5",
			Phone:   "+9720000005",
			Zip:     "2639805",
			City:    "Nazareth",
			Address: "Street 5",
			Region:  "North",
			Email:   "test5@gmail.com",
		},
		Payment: &cache.Payment{
			Transaction:  "b563feb7b2b84b6test5",
			Currency:     "USD",
			Provider:     "stripe",
			Amount:       2200,
			PaymentDt:    1637907725,
			Bank:         "discount",
			DeliveryCost: 300,
			GoodsTotal:   1900,
		},
		Items: []cache.Item{
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
		Delivery: &cache.Delivery{
			Name:    "Test User6",
			Phone:   "+9720000006",
			Zip:     "2639806",
			City:    "Ashdod",
			Address: "Street 6",
			Region:  "South",
			Email:   "test6@gmail.com",
		},
		Payment: &cache.Payment{
			Transaction:  "b563feb7b2b84b6test6",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       600,
			PaymentDt:    1637907726,
			Bank:         "alpha",
			DeliveryCost: 80,
			GoodsTotal:   520,
		},
		Items: []cache.Item{
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
		Delivery: &cache.Delivery{
			Name:    "Test User7",
			Phone:   "+9720000007",
			Zip:     "2639807",
			City:    "Beer Sheva",
			Address: "Street 7",
			Region:  "South",
			Email:   "test7@gmail.com",
		},
		Payment: &cache.Payment{
			Transaction:  "b563feb7b2b84b6test7",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       400,
			PaymentDt:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 50,
			GoodsTotal:   350,
		},
		Items: []cache.Item{
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
		Delivery: &cache.Delivery{
			Name:    "Test User8",
			Phone:   "+9720000008",
			Zip:     "2639808",
			City:    "Acre",
			Address: "Street 8",
			Region:  "North",
			Email:   "test8@gmail.com",
		},
		Payment: &cache.Payment{
			Transaction:  "b563feb7b2b84b6test8",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1200,
			PaymentDt:    1637907728,
			Bank:         "alpha",
			DeliveryCost: 150,
			GoodsTotal:   1050,
		},
		Items: []cache.Item{
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
		Delivery: &cache.Delivery{
			Name:    "Test User9",
			Phone:   "+9720000009",
			Zip:     "2639809",
			City:    "Safed",
			Address: "Street 9",
			Region:  "North",
			Email:   "test9@gmail.com",
		},
		Payment: &cache.Payment{
			Transaction:  "b563feb7b2b84b6test9",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       3000,
			PaymentDt:    1637907729,
			Bank:         "alpha",
			DeliveryCost: 400,
			GoodsTotal:   2600,
		},
		Items: []cache.Item{
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
		Delivery: &cache.Delivery{
			Name:    "Test User10",
			Phone:   "+9720000010",
			Zip:     "2639810",
			City:    "Tiberias",
			Address: "Street 10",
			Region:  "North",
			Email:   "test10@gmail.com",
		},
		Payment: &cache.Payment{
			Transaction:  "b563feb7b2b84b6test10",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       500,
			PaymentDt:    1637907730,
			Bank:         "alpha",
			DeliveryCost: 70,
			GoodsTotal:   430,
		},
		Items: []cache.Item{
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

func TestCache(t *testing.T) {
	cfg := &config.Cache{
		TTL:      1 * time.Second,
		Capacity: 4,
	}

	c := NewCache(cfg)

	t.Run("put length <= capacity", func(t *testing.T) {
		expectedStor := make(map[string]*cache.Order, cfg.Capacity)
		expectedFreq := make(map[string]*Info, cfg.Capacity)
		expectedLength := 3
		expectedStor[orders[0].OrderUID] = orders[0]
		expectedStor[orders[1].OrderUID] = orders[1]
		expectedStor[orders[2].OrderUID] = orders[2]
		expectedFreq[orders[0].OrderUID] = &Info{count: 1}
		expectedFreq[orders[1].OrderUID] = &Info{count: 1}
		expectedFreq[orders[2].OrderUID] = &Info{count: 1}

		c.Put(orders[0])
		c.Put(orders[1])
		c.Put(orders[2])

		require.Equal(t, expectedLength, c.length, "incorrect length")
		require.Equal(t, expectedStor, c.stor, "incorrect stor")
		for k, v := range expectedFreq {
			require.Equal(t, v.count, c.freq[k].count, "incorrect freq")
			require.False(t, c.freq[k].ttl.IsZero())
		}
		c.freq[orders[1].OrderUID].ttl = time.Now().Add(10 * time.Millisecond)
	})
	t.Run("get", func(t *testing.T) {
		expectedFreq := make(map[string]*Info, cfg.Capacity)
		expectedFreq[orders[0].OrderUID] = &Info{count: 3}
		expectedFreq[orders[1].OrderUID] = &Info{count: 2}
		expectedFreq[orders[2].OrderUID] = &Info{count: 1}
		expectedOrder := orders[0]
		order, ok := c.Get(expectedOrder.OrderUID)
		require.True(t, ok)
		require.Equal(t, expectedOrder, order)
		c.Get(expectedOrder.OrderUID)
		c.Get(orders[1].OrderUID)
		_, ok = c.Get(orders[4].OrderUID)
		require.False(t, ok)

		for k, v := range expectedFreq {
			require.Equal(t, v.count, c.freq[k].count, "incorrect freq")
			require.False(t, c.freq[k].ttl.IsZero())
		}
	})
	t.Run("checkTTL", func(t *testing.T) {
		expectedStor := make(map[string]*cache.Order, cfg.Capacity)
		expectedFreq := make(map[string]*Info, cfg.Capacity)
		expectedLength := 2
		expectedStor[orders[0].OrderUID] = orders[0]
		expectedStor[orders[2].OrderUID] = orders[2]
		expectedFreq[orders[0].OrderUID] = &Info{count: 3}
		expectedFreq[orders[2].OrderUID] = &Info{count: 1}
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			c.CheckTTL(ctx)
		}()
		time.Sleep(5 * time.Second)
		cancel()

		require.Equal(t, expectedLength, c.length, "incorrect length")
		require.Equal(t, expectedStor, c.stor, "incorrect stor")
		for k, v := range expectedFreq {
			require.Equal(t, v.count, c.freq[k].count, "incorrect freq")
			require.False(t, c.freq[k].ttl.IsZero())
		}
	})

	t.Run("put length > capacity", func(t *testing.T) {
		expectedStor := make(map[string]*cache.Order, cfg.Capacity)
		expectedFreq := make(map[string]*Info, cfg.Capacity)
		expectedLength := 4
		expectedStor[orders[0].OrderUID] = orders[0]
		expectedStor[orders[1].OrderUID] = orders[1]
		expectedStor[orders[3].OrderUID] = orders[3]
		expectedStor[orders[4].OrderUID] = orders[4]
		expectedFreq[orders[0].OrderUID] = &Info{count: 4}
		expectedFreq[orders[1].OrderUID] = &Info{count: 2}
		expectedFreq[orders[3].OrderUID] = &Info{count: 2}
		expectedFreq[orders[4].OrderUID] = &Info{count: 1}

		c.Put(orders[0])
		c.Put(orders[1])
		c.Put(orders[3])
		c.Get(orders[0].OrderUID)
		c.Get(orders[1].OrderUID)
		c.Get(orders[3].OrderUID)

		c.Put(orders[4])

		require.Equal(t, expectedLength, c.length, "incorrect length")
		require.Equal(t, expectedStor, c.stor, "incorrect stor")
		for k, v := range expectedFreq {
			require.Equal(t, v.count, c.freq[k].count, "incorrect freq")
			require.False(t, c.freq[k].ttl.IsZero())
		}
	})
}

func TestCache_Load(t *testing.T) {
	cfg := &config.Cache{
		TTL:      1 * time.Second,
		Capacity: 10,
	}
	ordersForLoad := make(map[string]*cache.Order, cfg.Capacity)
	expectedStor := make(map[string]*cache.Order, cfg.Capacity)
	expectedFreq := make(map[string]*Info, cfg.Capacity)

	for _, order := range orders {
		ordersForLoad[order.OrderUID] = order

		expectedStor[order.OrderUID] = order
		expectedFreq[order.OrderUID] = &Info{count: 1}
	}
	c := NewCache(cfg)
	expectedLength := 10

	c.Load(ordersForLoad)

	require.Equal(t, expectedLength, c.length, "incorrect length")
	require.Equal(t, expectedStor, c.stor, "incorrect stor")
	for k, v := range expectedFreq {
		require.Equal(t, v.count, c.freq[k].count, "incorrect freq")
		require.False(t, c.freq[k].ttl.IsZero())
	}
}
