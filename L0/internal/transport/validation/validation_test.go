package validation

import (
	"L0/internal/transport/dto"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateOrder(t *testing.T) {
	correctTime, _ := time.Parse(time.RFC3339, "2021-11-26T06:12:07Z")
	orders := []*dto.Order{
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
			Payment: &dto.Payment{
				Transaction:  "b563feb7b2b84b6test2",
				RequestID:    "",
				Currency:     "USD",
				Provider:     "wbpay",
				Amount:       1815,
				PaymentDT:    1637907327,
				Bank:         "alpha",
				DeliveryCost: 1501,
				GoodsTotal:   313,
				CustomFee:    0,
			},
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
			OrderUID:    "b563feb7b2b84b6test",
			TrackNumber: "WBILMTESTTRACK",
			Entry:       "WBIL",
			Delivery: &dto.Delivery{
				Name:    "Test Testov",
				Phone:   "",
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
					Sale:        -1,
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
			OrderUID:    "",
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
			Locale:            "asd",
			InternalSignature: "",
			CustomerID:        "test",
			DeliveryService:   "meest",
			Shardkey:          "9",
			SmID:              -1,
			DateCreated:       correctTime,
			OofShard:          "1",
		},
	}
	testcases := []struct {
		name        string
		order       *dto.Order
		expectedErr bool
		msg         string
	}{
		{
			name:        "valid 1",
			order:       orders[0],
			expectedErr: false,
			msg:         "",
		},
		{
			name:        "valid 2",
			order:       orders[1],
			expectedErr: false,
			msg:         "",
		},
		{
			name:        "invalid sale phone",
			order:       orders[2],
			expectedErr: true,
			msg:         "Order isn't valid:\n\t-Field Phone isn't valid, validation tag - required.\n\t-Field Sale isn't valid, validation tag - gte.",
		},
		{
			name:        "invalid orderUID locale smID",
			order:       orders[3],
			expectedErr: true,
			msg:         "Order isn't valid:\n\t-Field OrderUID isn't valid, validation tag - required.\n\t-Field Locale isn't valid, validation tag - len.\n\t-Field SmID isn't valid, validation tag - gt.",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOrder(tc.order)
			if tc.expectedErr {
				require.Error(t, err, "expected error")
				require.Equal(t, tc.msg, err.Error(), "expected error")

			} else {
				require.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestValidateOrderInvalid(t *testing.T) {
	var order *dto.Order
	err := ValidateOrder(order)

	require.Error(t, err, "expected error")
}
