package mapper

import (
	"L0/internal/service"
	"L0/internal/transport/dto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var correctTime, _ = time.Parse(time.RFC3339, "2021-11-26T06:12:07Z")

var (
	ordersDTO = []*dto.Order{
		{
			OrderUID:    "b563feb7b2b84b6test",
			TrackNumber: "WBILMTESTTRACK",
			Entry:       "WBIL",
			Delivery: &dto.Delivery{
				Name:    "Test Testov",
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: "Ploshad Mira 15",
				Region:  "Kraiot",
				Email:   "test@gmail.com",
			},
			Payment: &dto.Payment{
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
			Items: []dto.Item{
				{
					ChrtID:      9934930,
					TrackNumber: "WBILMTESTTRACK",
					Price:       453,
					RID:         "ab4219087a764ae0btest",
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
			Delivery: &dto.Delivery{
				Name:    "Test2 Testov2",
				Phone:   "+9720000002",
				Zip:     "2639809",
				City:    "Mozkinburg",
				Address: "Ploshad Pira 15",
				Region:  "Daleko",
				Email:   "tes2t@gmail.com",
			},
			Payment: nil,
			Items: []dto.Item{
				{
					ChrtID:      9934931,
					TrackNumber: "WBILMTESTTRACK",
					Price:       43,
					RID:         "ab4219087a764ae0btest2",
					Name:        "Mascaras",
					Sale:        0,
					Size:        "1",
					TotalPrice:  43,
					NmID:        238912,
					Brand:       "Vivienne Sabor",
					Status:      200,
				},
			},
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
			Delivery: &dto.Delivery{
				Name:    "Test Testov2",
				Phone:   "+9720000002",
				Zip:     "2639809",
				City:    "Kiryat Moskvin",
				Address: "Ploshad Pira 15",
				Region:  "Kayot",
				Email:   "tes2t@gmail.com",
			},
			Payment: &dto.Payment{
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
			Items: []dto.Item{
				{
					ChrtID:      9933930,
					TrackNumber: "WBILMTESTTRACK",
					Price:       452,
					RID:         "ab4219087a764ae0btest2",
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
	ordersPublicDTO = []*dto.PublicOrder{
		{
			OrderUID:        "b563feb7b2b84b6test",
			DateCreated:     correctTime,
			Currency:        "USD",
			Amount:          1817,
			DeliveryCost:    1500,
			GoodsTotal:      317,
			DeliveryService: "meest",
			Items: []dto.PublicItem{
				{
					NmID:       2389212,
					Name:       "Mascaras",
					Brand:      "Vivienne Sabo",
					Size:       "0",
					Price:      453,
					Sale:       30,
					TotalPrice: 317,
				},
			},
		},
		{
			OrderUID:        "b563feb7b2b84b6test2",
			DateCreated:     correctTime,
			Currency:        "USD",
			Amount:          1817,
			DeliveryCost:    1500,
			GoodsTotal:      317,
			DeliveryService: "meest2",
			Items:           []dto.PublicItem{},
		},
		{
			OrderUID:        "b563feb7b2b84b6test2",
			DateCreated:     correctTime,
			Currency:        "USD",
			Amount:          1812,
			DeliveryCost:    1501,
			GoodsTotal:      3313,
			DeliveryService: "meest1",
			Items: []dto.PublicItem{
				{
					NmID:       2389222,
					Name:       "Mascarad",
					Brand:      "Vivienne Sabor",
					Size:       "0",
					Price:      452,
					Sale:       0,
					TotalPrice: 452,
				},
			},
		},
	}
)

func TestServiceToDTO(t *testing.T) {
	testCases := []struct {
		name          string
		order         *service.Order
		expectedOrder *dto.Order
		expectedError bool
	}{
		{
			name:          "valid",
			order:         ordersService[0],
			expectedOrder: ordersDTO[0],
			expectedError: false,
		},
		{
			name:          "expected error",
			order:         ordersService[1],
			expectedOrder: nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			order, err := ServiceToDTO(tc.order)
			if tc.expectedError {
				assert.Error(t, err, tc.name+" expected error")
			} else {
				assert.NoError(t, err, tc.name+" expected no error")
			}
			assert.Equal(t, tc.expectedOrder, order)
		})
	}

}

func TestDTOToService(t *testing.T) {
	testCases := []struct {
		name          string
		Order         *dto.Order
		expectedOrder *service.Order
		expectedError bool
	}{
		{
			name:          "valid",
			Order:         ordersDTO[0],
			expectedOrder: ordersService[0],
			expectedError: false,
		},
		{
			name:          "expected error",
			Order:         ordersDTO[1],
			expectedOrder: nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			order, err := DTOToService(tc.Order)
			if tc.expectedError {
				assert.Error(t, err, tc.name+" expected error")
			} else {
				assert.NoError(t, err, tc.name+" expected no error")
			}
			assert.Equal(t, tc.expectedOrder, order)
		})
	}
}

func TestServiceToPublicDTO(t *testing.T) {
	testCases := []struct {
		name          string
		order         *service.Order
		expectedOrder *dto.PublicOrder
		expectedError bool
	}{
		{
			name:          "valid",
			order:         ordersService[0],
			expectedOrder: ordersPublicDTO[0],
			expectedError: false,
		},
		{
			name:          "expected error",
			order:         ordersService[1],
			expectedOrder: nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			order, err := ServiceToPublicDTO(tc.order)
			if tc.expectedError {
				assert.Error(t, err, tc.name+" expected error")
			} else {
				assert.NoError(t, err, tc.name+" expected no error")
			}
			assert.Equal(t, tc.expectedOrder, order)
		})
	}

}
