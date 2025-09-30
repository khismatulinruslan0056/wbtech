package importer

import (
	"errors"
	"math/rand"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	validTpl = NewTemplate()
	errorTpl = template.Must(template.New("exec error").Parse("{{.NonExistentField}}"))

	generator = NewGenerator(rand.New(rand.NewSource(42)), validTpl)
)

func TestGenerateOrder(t *testing.T) {

	testCases := []struct {
		name          string
		valid         bool
		expectedKey   []byte
		expectedValue []byte
		expectedErr   bool
		err           error
		tpl           *template.Template
	}{
		{ //паршивый тест с рандомом, падает через раз изз особенностей рандома
			name:          "Valid",
			valid:         true,
			expectedKey:   []byte("I0QbPx#zAW*E+CHT-0"),
			expectedValue: []byte("{\n  \"order_uid\": \"I0QbPx#zAW*E+CHT-0\",\n  \"track_number\": \"TRK-FXAe9\",\n  \"entry\": \"WBIL\",\n  \"delivery\": {\n    \"name\": \"MGfS3fEUuy\",\n    \"phone\": \"+79745640358\",\n    \"zip\": \"452670\",\n    \"city\": \"x29_^$\",\n    \"address\": \"rkazQ1I%XSLjZcA\",\n    \"region\": \"WMeKFjOY\",\n    \"email\": \"QTZ5z@test.ru\"\n  },\n  \"payment\": {\n    \"transaction\": \"I0QbPx#zAW*E+CHT-0\",\n    \"request_id\": \"Jj(2o\",\n    \"currency\": \"USD\",\n    \"provider\": \"wbpay\",\n    \"amount\": 15327,\n        \"bank\": \"*_$0mn\",\n    \"delivery_cost\": 368,\n    \"goods_total\": 31,\n    \"custom_fee\": 639\n  },\n  \"items\": [\n    {\n      \"chrt_id\": 9286472,\n      \"track_number\": \"ITM-6Tl(\",\n      \"price\": 1179,\n      \"rid\": \"ko7QlSVksnqj\",\n      \"name\": \"!TPP*Qs\",\n      \"sale\": 28,\n      \"size\": \"2\",\n      \"total_price\": 43,\n      \"nm_id\": 7152384,\n      \"brand\": \"VCz2wgxD\",\n      \"status\": 378\n    }\n  ],\n  \"locale\": \"ru\",\n  \"internal_signature\": \"hip_s\",\n  \"customer_id\": \"Geg@Qe\",\n  \"delivery_service\": \"meest\",\n  \"shardkey\": \"2\",\n  \"sm_id\": 713,\n    \"oof_shard\": \"2\"\n}"),
			expectedErr:   false,
			tpl:           validTpl,
		},

		{
			name:          "Incorrect template",
			valid:         true,
			expectedKey:   nil,
			expectedValue: nil,
			expectedErr:   true,
			tpl:           errorTpl,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			generator.tmpl = tc.tpl
			key, value, err := generator.GenerateOrder(i, tc.valid)
			if tc.expectedErr && err != nil {
				return
			}
			if tc.expectedErr && err == nil {
				t.Errorf("Expected result %q, but got %q", tc.err.Error(), err.Error())
			}
			if !tc.expectedErr && err != nil {
				t.Errorf("Expected result %q, but got %q", tc.err.Error(), err.Error())
			}

			if string(key) != string(tc.expectedKey) {
				t.Errorf("Expected result %q, but got %q", string(tc.expectedKey), string(key))
			}

			val, err := splitJSONOrder(value)
			if err != nil {
				t.Errorf("Expected no error, but got %q", err.Error())
			}

			if val != string(tc.expectedValue) {
				t.Errorf("Expected result \n%q,\n but got \n%q", string(tc.expectedValue), val)
			}
		})
	}
}

func splitJSONOrder(jsonOrder []byte) (string, error) {
	path := strings.Split(string(jsonOrder), "\"payment_dt\":")
	if len(path) != 2 {
		return "", errors.New("invalid json order")
	}
	valueBeforePayment := path[0]
	path = strings.Split(path[1], "\"date_created\":")
	if len(path) != 2 {
		return "", errors.New("invalid json order")
	}
	valueBeforeDate := path[0][13:]
	valueAfterDate := path[1][30:]
	return valueBeforePayment + valueBeforeDate + valueAfterDate, nil
}

func TestGenerate(t *testing.T) {
	testCases := []struct {
		name      string
		valid     bool
		checkFunc func(t *testing.T, order *Order)
	}{
		{
			name:  "invalid phone",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Delivery.Phone, "Phone should be empty for this invalid case")
			},
		},
		{
			name:  "invalid date",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				date := o.DateCreated
				_, err := time.Parse(time.RFC3339, date)
				if err == nil {
					t.Error("Expected error for invalid date")
				}
			},
		},
		{
			name:  "invalid email",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Delivery.Email, "Email should be empty for this invalid case")
			},
		},
		{
			name:  "invalid name",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Delivery.Name, "Name should be empty for this invalid case")
			},
		},
		{
			name:  "invalid city",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Delivery.City, "City should be empty for this invalid case")
			},
		},
		{
			name:  "invalid address",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Delivery.Address, "Address should be empty for this invalid case")
			},
		},
		{
			name:  "invalid region",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Delivery.Region, "Region should be empty for this invalid case")
			},
		},
		{
			name:  "invalid requestID",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Payment.RequestID, "RequestID should be empty for this invalid case")
			},
		},
		{
			name:  "invalid bank",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Payment.Bank, "Bank should be empty for this invalid case")
			},
		},
		{
			name:  "invalid itemTrackNumber",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, "ITM-", o.Items[0].TrackNumber, "TrackNumber should be \"ITM-\" for this invalid case")
			},
		},
		{
			name:  "invalid rid",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Items[0].Rid, "Rid should be empty for this invalid case")
			},
		},
		{
			name:  "invalid itemName",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Items[0].Name, "Name should be empty for this invalid case")
			},
		},
		{
			name:  "invalid brand",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Items[0].Brand, "Brand should be empty for this invalid case")
			},
		},
		{
			name:  "invalid locale",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, "asd", o.Locale, "Locale should be \"asd\" for this invalid case")
			},
		},
		{
			name:  "invalid internalSignature",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.InternalSignature, "InternalSignature should be empty for this invalid case")
			},
		},
		{
			name:  "invalid customerID",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.CustomerID, "CustomerID should be empty for this invalid case")
			},
		},
		{
			name:  "invalid shardKey",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, "-1", o.ShardKey, "ShardKey should be \"-1\" for this invalid case")
			},
		},
		{
			name:  "invalid oofShard",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, "-1", o.OofShard, "OofShard should be \"-1\" for this invalid case")
			},
		},
		{
			name:  "invalid trackNumber",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.TrackNumber, "TrackNumber should be empty for this invalid case")
			},
		},
		{
			name:  "invalid zip",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Delivery.Zip, "Zip should be empty for this invalid case")
			},
		},
		{
			name:  "invalid amount",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.Payment.Amount, "Amount should be -1 for this invalid case")
			},
		},
		{
			name:  "invalid deliveryCost",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.Payment.DeliveryCost, "DeliveryCost should be -1 for this invalid case")

			},
		},
		{
			name:  "invalid goodsTotal",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.Payment.GoodsTotal, "GoodsTotal should be -1 for this invalid case")

			},
		},
		{
			name:  "invalid customFee",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.Payment.CustomFee, "CustomFee should be -1 for this invalid case")

			},
		},
		{
			name:  "invalid chrtID",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.Items[0].ChrtID, "ChrtID should be -1 for this invalid case")

			},
		},
		{
			name:  "invalid price",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.Items[0].Price, "Price should be -1 for this invalid case")

			},
		},
		{
			name:  "invalid sale",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.Items[0].Sale, "Sale should be -1 for this invalid case")
			},
		},
		{
			name:  "invalid size",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.Items[0].Size, "Size should be empty for this invalid case")
			},
		},
		{
			name:  "invalid totalPrice",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.Items[0].TotalPrice, "TotalPrice should be -1 for this invalid case")
			},
		},
		{
			name:  "invalid nmID",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.Items[0].NmID, "NmID should be -1 for this invalid case")

			},
		},
		{
			name:  "invalid status",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.Items[0].Status, "Status should be -1 for this invalid case")
			},
		},
		{
			name:  "invalid smID",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Equal(t, -1, o.SmID, "SmID should be -1 for this invalid case")
			},
		},
		{
			name:  "invalid orderUID",
			valid: false,
			checkFunc: func(t *testing.T, o *Order) {
				assert.Empty(t, o.OrderUID, "OrderUID should be empty for this invalid case")
			},
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, orderRes := generator.generate(i, tc.valid)
			tc.checkFunc(t, orderRes)
		})
	}
}
