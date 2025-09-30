package service

import (
	"L0/internal/cache"
	mocksC "L0/internal/mocks/cache"
	mocksS "L0/internal/mocks/storage"
	"context"
	"errors"
	"io"
	"log/slog"

	"L0/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceToModel(t *testing.T) {
	testCases := []struct {
		name             string
		order            *Order
		expectedOrder    *models.Order
		expectedDelivery *models.Delivery
		expectedPayment  *models.Payment
		expectedItems    []*models.Item
		expectedError    bool
	}{
		{
			name:             "valid",
			order:            ordersService[0],
			expectedOrder:    ordersModel[0],
			expectedDelivery: deliveryModel[0],
			expectedPayment:  paymentModel[0],
			expectedItems:    itemsModel[0],
			expectedError:    false,
		},
		{
			name:             "expected error",
			order:            ordersService[1],
			expectedOrder:    nil,
			expectedDelivery: nil,
			expectedPayment:  nil,
			expectedItems:    nil,
			expectedError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			order, delivery, payment, items, err := serviceToModel(tc.order)
			if tc.expectedError {
				assert.Error(t, err, tc.name+" expected error")
			} else {
				assert.NoError(t, err, tc.name+" expected no error")
			}
			assert.Equal(t, tc.expectedOrder, order)
			assert.Equal(t, tc.expectedDelivery, delivery)
			assert.Equal(t, tc.expectedPayment, payment)
			assert.Equal(t, tc.expectedItems, items)
		})
	}

}

func TestModelToService(t *testing.T) {
	testCases := []struct {
		name          string
		Order         *models.Order
		Delivery      *models.Delivery
		Payment       *models.Payment
		Items         []*models.Item
		expectedOrder *Order
		expectedError bool
	}{
		{
			name:          "valid",
			Order:         ordersModel[0],
			Delivery:      deliveryModel[0],
			Payment:       paymentModel[0],
			Items:         itemsModel[0],
			expectedOrder: ordersService[0],
			expectedError: false,
		},
		{
			name:          "expected error",
			Order:         ordersModel[1],
			Delivery:      deliveryModel[1],
			Payment:       paymentModel[1],
			Items:         itemsModel[1],
			expectedOrder: nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			order, err := modelToService(tc.Order, tc.Delivery, tc.Payment, tc.Items)
			if tc.expectedError {
				assert.Error(t, err, tc.name+" expected error")
			} else {
				assert.NoError(t, err, tc.name+" expected no error")
			}
			assert.Equal(t, tc.expectedOrder, order)
		})
	}
}

func TestServiceToCache(t *testing.T) {
	testCases := []struct {
		name          string
		order         *Order
		expectedOrder *cache.Order
		expectedError bool
	}{
		{
			name:          "valid",
			order:         ordersService[0],
			expectedOrder: ordersCache[0],
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
			order, err := serviceToCache(tc.order)
			if tc.expectedError {
				assert.Error(t, err, tc.name+" expected error")
			} else {
				assert.NoError(t, err, tc.name+" expected no error")
			}
			assert.Equal(t, tc.expectedOrder, order)
		})
	}

}

func TestCacheToService(t *testing.T) {
	testCases := []struct {
		name          string
		Order         *cache.Order
		expectedOrder *Order
		expectedError bool
	}{
		{
			name:          "valid",
			Order:         ordersCache[0],
			expectedOrder: ordersService[0],
			expectedError: false,
		},
		{
			name:          "expected error",
			Order:         ordersCache[1],
			expectedOrder: nil,
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			order, err := cacheToService(tc.Order)
			if tc.expectedError {
				assert.Error(t, err, tc.name+" expected error")
			} else {
				assert.NoError(t, err, tc.name+" expected no error")
			}
			assert.Equal(t, tc.expectedOrder, order)
		})
	}
}

func TestService_Add(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("serviceToCache error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockCache := new(mocksC.Cacher)
		service := New(mockStorage, mockCache, log)
		order := ordersService[1]
		_, err := service.Add(context.Background(), order)
		require.Error(t, err, "expected error")
		mockStorage.AssertNotCalled(t, "AddOrder", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "AddDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "AddPayment", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "AddItems", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "BeginTx", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "WithTx", mock.Anything)

		mockCache.AssertNotCalled(t, "Put", mock.Anything)
	})

	t.Run("beginTX error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockCache := new(mocksC.Cacher)
		service := New(mockStorage, mockCache, log)
		order := ordersService[0]

		//beginTXErr := errors.New("beginTX error")
		mockStorage.On("BeginTx", mock.Anything, mock.Anything).Return(nil, RetryFailed)
		_, err := service.Add(context.Background(), order)
		require.Error(t, err, "expected error")
		require.ErrorIs(t, err, RetryFailed)

		mockStorage.AssertNotCalled(t, "AddOrder", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "AddDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "AddPayment", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "AddItems", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "BeginTx", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "WithTx", mock.Anything)

		mockCache.AssertNotCalled(t, "Put", mock.Anything)
	})

	t.Run("add order error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockTx := new(mocksS.Tx)
		mockCache := new(mocksC.Cacher)

		service := New(mockStorage, mockCache, log)
		orderService := ordersService[0]
		order := ordersModel[0]
		mockStorage.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockStorage.On("WithTx", mockTx).Return(mockStorage)
		mockStorage.On("AddOrder", mock.Anything, order).Return("", RetryFailed)
		mockTx.On("Rollback").Return(nil)

		_, err := service.Add(context.Background(), orderService)
		require.Error(t, err, "expected error")
		require.ErrorIs(t, err, RetryFailed)

		mockStorage.AssertCalled(t, "AddOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "BeginTx", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "WithTx", mock.Anything)
		mockTx.AssertCalled(t, "Rollback")

		mockStorage.AssertNotCalled(t, "AddDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "AddPayment", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "AddItems", mock.Anything, mock.Anything)
		mockTx.AssertNotCalled(t, "Commit")
		mockCache.AssertNotCalled(t, "Put", mock.Anything)
	})

	t.Run("add delivery error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockTx := new(mocksS.Tx)
		mockCache := new(mocksC.Cacher)

		service := New(mockStorage, mockCache, log)
		orderService := ordersService[0]
		order := ordersModel[0]
		delivery := deliveryModel[0]
		mockStorage.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockStorage.On("WithTx", mockTx).Return(mockStorage)
		mockStorage.On("AddOrder", mock.Anything, order).Return("", nil)
		mockStorage.On("AddDelivery", mock.Anything, delivery).Return(int64(1), RetryFailed)
		mockTx.On("Rollback").Return(nil)

		_, err := service.Add(context.Background(), orderService)
		require.Error(t, err, "expected error")
		require.ErrorIs(t, err, RetryFailed)

		mockStorage.AssertCalled(t, "AddOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "AddDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "BeginTx", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "WithTx", mock.Anything)
		mockTx.AssertCalled(t, "Rollback")

		mockStorage.AssertNotCalled(t, "AddPayment", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "AddItems", mock.Anything, mock.Anything)
		mockTx.AssertNotCalled(t, "Commit")
		mockCache.AssertNotCalled(t, "Put", mock.Anything)
	})

	t.Run("add payment error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockTx := new(mocksS.Tx)
		mockCache := new(mocksC.Cacher)

		service := New(mockStorage, mockCache, log)
		orderService := ordersService[0]
		order := ordersModel[0]
		delivery := deliveryModel[0]
		payment := paymentModel[0]
		mockStorage.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockStorage.On("WithTx", mockTx).Return(mockStorage)
		mockStorage.On("AddOrder", mock.Anything, order).Return("", nil)
		mockStorage.On("AddDelivery", mock.Anything, delivery).Return(int64(1), nil)
		mockStorage.On("AddPayment", mock.Anything, payment).Return(int64(1), RetryFailed)
		mockTx.On("Rollback").Return(nil)

		_, err := service.Add(context.Background(), orderService)
		require.Error(t, err, "expected error")
		require.ErrorIs(t, err, RetryFailed)

		mockStorage.AssertCalled(t, "AddOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "AddDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "BeginTx", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "WithTx", mock.Anything)
		mockStorage.AssertCalled(t, "AddPayment", mock.Anything, mock.Anything)
		mockTx.AssertCalled(t, "Rollback")

		mockStorage.AssertNotCalled(t, "AddItems", mock.Anything, mock.Anything)
		mockTx.AssertNotCalled(t, "Commit")
		mockCache.AssertNotCalled(t, "Put", mock.Anything)
	})

	t.Run("add items error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockTx := new(mocksS.Tx)
		mockCache := new(mocksC.Cacher)

		service := New(mockStorage, mockCache, log)
		orderService := ordersService[0]
		order := ordersModel[0]
		delivery := deliveryModel[0]
		payment := paymentModel[0]
		items := itemsModel[0]
		mockStorage.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockStorage.On("WithTx", mockTx).Return(mockStorage)
		mockStorage.On("AddOrder", mock.Anything, order).Return("", nil)
		mockStorage.On("AddDelivery", mock.Anything, delivery).Return(int64(1), nil)
		mockStorage.On("AddPayment", mock.Anything, payment).Return(int64(1), nil)
		mockStorage.On("AddItems", mock.Anything, items).Return(nil, RetryFailed)
		mockTx.On("Rollback").Return(nil)

		_, err := service.Add(context.Background(), orderService)
		require.Error(t, err, "expected error")
		require.ErrorIs(t, err, RetryFailed)

		mockStorage.AssertCalled(t, "AddOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "AddDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "BeginTx", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "WithTx", mock.Anything)
		mockStorage.AssertCalled(t, "AddPayment", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "AddItems", mock.Anything, mock.Anything)
		mockTx.AssertCalled(t, "Rollback")

		mockTx.AssertNotCalled(t, "Commit")
		mockCache.AssertNotCalled(t, "Put", mock.Anything)
	})

	t.Run("rollback error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockTx := new(mocksS.Tx)
		mockCache := new(mocksC.Cacher)

		service := New(mockStorage, mockCache, log)
		orderService := ordersService[0]
		order := ordersModel[0]
		delivery := deliveryModel[0]
		payment := paymentModel[0]
		items := itemsModel[0]
		rollbackErr := errors.New("rollback error")

		mockStorage.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockStorage.On("WithTx", mockTx).Return(mockStorage)
		mockStorage.On("AddOrder", mock.Anything, order).Return("", nil)
		mockStorage.On("AddDelivery", mock.Anything, delivery).Return(int64(1), nil)
		mockStorage.On("AddPayment", mock.Anything, payment).Return(int64(1), nil)
		mockStorage.On("AddItems", mock.Anything, items).Return(nil, RetryFailed)
		mockTx.On("Rollback").Return(rollbackErr)

		_, err := service.Add(context.Background(), orderService)
		require.Error(t, err, "expected error")
		require.ErrorIs(t, err, RetryFailed)

		mockStorage.AssertCalled(t, "AddOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "AddDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "BeginTx", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "WithTx", mock.Anything)
		mockStorage.AssertCalled(t, "AddPayment", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "AddItems", mock.Anything, mock.Anything)
		mockTx.AssertCalled(t, "Rollback")

		mockTx.AssertNotCalled(t, "Commit")
		mockCache.AssertNotCalled(t, "Put", mock.Anything)
	})

	t.Run("commit error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockTx := new(mocksS.Tx)
		mockCache := new(mocksC.Cacher)

		service := New(mockStorage, mockCache, log)
		orderService := ordersService[0]
		order := ordersModel[0]
		delivery := deliveryModel[0]
		payment := paymentModel[0]
		items := itemsModel[0]
		commitErr := errors.New("commit error")

		mockStorage.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockStorage.On("WithTx", mockTx).Return(mockStorage)
		mockStorage.On("AddOrder", mock.Anything, order).Return("", nil)
		mockStorage.On("AddDelivery", mock.Anything, delivery).Return(int64(1), nil)
		mockStorage.On("AddPayment", mock.Anything, payment).Return(int64(1), nil)
		mockStorage.On("AddItems", mock.Anything, items).Return(nil, nil)
		mockTx.On("Rollback").Return(nil)
		mockTx.On("Commit").Return(commitErr)

		_, err := service.Add(context.Background(), orderService)
		require.Error(t, err, "expected error")
		require.ErrorIs(t, err, RetryFailed)

		mockStorage.AssertCalled(t, "AddOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "AddDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "BeginTx", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "WithTx", mock.Anything)
		mockStorage.AssertCalled(t, "AddPayment", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "AddItems", mock.Anything, mock.Anything)
		mockTx.AssertCalled(t, "Rollback")
		mockTx.AssertCalled(t, "Commit")

		mockCache.AssertNotCalled(t, "Put", mock.Anything)
	})

	t.Run("success retry and put to cache", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockTx := new(mocksS.Tx)
		mockCache := new(mocksC.Cacher)

		service := New(mockStorage, mockCache, log)
		orderService := ordersService[0]
		order := ordersModel[0]
		delivery := deliveryModel[0]
		payment := paymentModel[0]
		items := itemsModel[0]
		orderCache := ordersCache[0]
		commitErr := errors.New("commit error")

		mockStorage.On("BeginTx", mock.Anything, mock.Anything).Return(mockTx, nil)
		mockStorage.On("WithTx", mockTx).Return(mockStorage)
		mockStorage.On("AddOrder", mock.Anything, order).Return("", nil)
		mockStorage.On("AddDelivery", mock.Anything, delivery).Return(int64(1), nil)
		mockStorage.On("AddPayment", mock.Anything, payment).Return(int64(1), nil)
		mockStorage.On("AddItems", mock.Anything, items).Return(nil, nil)
		mockTx.On("Rollback").Return(nil)
		mockTx.On("Commit").Return(commitErr).Twice()
		mockTx.On("Commit").Return(nil).Once()
		mockCache.On("Put", orderCache)

		_, err := service.Add(context.Background(), orderService)
		require.NoError(t, err, "expected  no error")

		mockStorage.AssertCalled(t, "AddOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "AddDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "BeginTx", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "WithTx", mock.Anything)
		mockStorage.AssertCalled(t, "AddPayment", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "AddItems", mock.Anything, mock.Anything)
		mockTx.AssertCalled(t, "Rollback")
		mockTx.AssertCalled(t, "Commit")
		mockCache.AssertCalled(t, "Put", mock.Anything)

	})
}

func TestService_Get(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("Get order from cash ", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockCache := new(mocksC.Cacher)
		service := New(mockStorage, mockCache, log)
		orderID := "qweasd"

		mockCache.On("Get", orderID).Return(nil, true)

		_, err := service.Get(context.Background(), orderID)
		require.Error(t, err, "expected error")
		mockStorage.AssertNotCalled(t, "GetOrder", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "GetDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "GetPayment", mock.Anything, mock.Anything)
		mockStorage.AssertNotCalled(t, "GetItemsByID", mock.Anything, mock.Anything)
		mockCache.AssertNotCalled(t, "Put", mock.Anything)
		mockCache.AssertCalled(t, "Get", orderID)
	})
	t.Run("Get error ", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockCache := new(mocksC.Cacher)
		service := New(mockStorage, mockCache, log)
		orderID := "qweasd"
		getError := errors.New("get error")
		mockCache.On("Get", orderID).Return(nil, false)
		mockStorage.On("GetOrder", mock.Anything, mock.Anything).Return(nil, getError)
		mockStorage.On("GetDelivery", mock.Anything, mock.Anything).Return(nil, getError)
		mockStorage.On("GetPayment", mock.Anything, mock.Anything).Return(nil, getError)
		mockStorage.On("GetItemsByID", mock.Anything, mock.Anything).Return(nil, getError)

		_, err := service.Get(context.Background(), orderID)
		require.Error(t, err, "expected error")

		mockStorage.AssertCalled(t, "GetOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetPayment", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetItemsByID", mock.Anything, mock.Anything)

		mockCache.AssertNotCalled(t, "Put", mock.Anything)
		mockCache.AssertCalled(t, "Get", orderID)
	})
	t.Run("modelToService error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockCache := new(mocksC.Cacher)
		service := New(mockStorage, mockCache, log)
		orderID := "qweasd"
		mockCache.On("Get", orderID).Return(nil, false)
		mockStorage.On("GetOrder", mock.Anything, mock.Anything).Return(nil, nil)
		mockStorage.On("GetDelivery", mock.Anything, mock.Anything).Return(nil, nil)
		mockStorage.On("GetPayment", mock.Anything, mock.Anything).Return(nil, nil)
		mockStorage.On("GetItemsByID", mock.Anything, mock.Anything).Return(nil, nil)

		_, err := service.Get(context.Background(), orderID)
		require.Error(t, err, "expected error")

		mockStorage.AssertCalled(t, "GetOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetPayment", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetItemsByID", mock.Anything, mock.Anything)

		mockCache.AssertCalled(t, "Get", orderID)
		mockCache.AssertNotCalled(t, "Put", mock.Anything)
	})
	t.Run("Get success", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockCache := new(mocksC.Cacher)
		service := New(mockStorage, mockCache, log)
		orderID := "b563feb7b2b84b6test"
		order := ordersModel[0]
		delivery := deliveryModel[0]
		payment := paymentModel[0]
		items := itemsModel[0]
		mockCache.On("Get", orderID).Return(nil, false)
		mockCache.On("Put", mock.Anything)
		mockStorage.On("GetOrder", mock.Anything, mock.Anything).Return(order, nil)
		mockStorage.On("GetDelivery", mock.Anything, mock.Anything).Return(delivery, nil)
		mockStorage.On("GetPayment", mock.Anything, mock.Anything).Return(payment, nil)
		mockStorage.On("GetItemsByID", mock.Anything, mock.Anything).Return(items, nil)

		_, err := service.Get(context.Background(), orderID)
		require.NoError(t, err, "expected no error")

		mockStorage.AssertCalled(t, "GetOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetPayment", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetItemsByID", mock.Anything, mock.Anything)

		mockCache.AssertCalled(t, "Get", orderID)
		mockCache.AssertCalled(t, "Put", mock.Anything)
	})
	t.Run("Cache miss, cache hit", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockCache := new(mocksC.Cacher)
		service := New(mockStorage, mockCache, log)
		orderID := "b563feb7b2b84b6test"
		order := ordersModel[0]
		delivery := deliveryModel[0]
		payment := paymentModel[0]
		items := itemsModel[0]
		orderCache := ordersCache[0]
		mockCache.On("Get", orderID).Return(nil, false).Once()
		mockCache.On("Put", mock.Anything)
		mockStorage.On("GetOrder", mock.Anything, mock.Anything).Return(order, nil).After(100 * time.Millisecond).Once()
		mockStorage.On("GetDelivery", mock.Anything, mock.Anything).Return(delivery, nil)
		mockStorage.On("GetPayment", mock.Anything, mock.Anything).Return(payment, nil)
		mockStorage.On("GetItemsByID", mock.Anything, mock.Anything).Return(items, nil)

		_, err := service.Get(context.Background(), orderID)
		require.NoError(t, err, "expected no error")
		mockCache.On("Get", orderID).Return(orderCache, true)
		start := time.Now()
		_, err = service.Get(context.Background(), orderID)
		dur := time.Since(start)
		require.NoError(t, err, "expected no error")
		require.GreaterOrEqual(t, 100*time.Millisecond, dur, "expected 100ms")
		mockStorage.AssertCalled(t, "GetOrder", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetDelivery", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetPayment", mock.Anything, mock.Anything)
		mockStorage.AssertCalled(t, "GetItemsByID", mock.Anything, mock.Anything)

		mockCache.AssertCalled(t, "Get", orderID)
		mockCache.AssertCalled(t, "Put", mock.Anything)

	})

}

func TestService_WarmUpCache(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("get error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockCache := new(mocksC.Cacher)
		service := New(mockStorage, mockCache, log)

		getErr := errors.New("get error")
		mockCache.On("Load", mock.Anything)
		mockStorage.On("GetAllOrders", mock.Anything).Return(nil, getErr)
		mockStorage.On("GetALLDeliveries", mock.Anything).Return(nil, getErr)
		mockStorage.On("GetAllPayments", mock.Anything).Return(nil, getErr)
		mockStorage.On("GetAllItems", mock.Anything).Return(nil, getErr)

		err := service.WarmUpCache(context.Background())
		require.Error(t, err, "expected  error")

		mockStorage.AssertCalled(t, "GetAllOrders", mock.Anything)
		mockStorage.AssertCalled(t, "GetALLDeliveries", mock.Anything)
		mockStorage.AssertCalled(t, "GetAllPayments", mock.Anything)
		mockStorage.AssertCalled(t, "GetAllItems", mock.Anything)

		mockCache.AssertNotCalled(t, "Load", mock.Anything)
	})

	t.Run("get success", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockCache := new(mocksC.Cacher)
		service := New(mockStorage, mockCache, log)

		orders := append([]*models.Order{ordersModel[0]}, ordersModel[2])
		deliveries := append([]*models.Delivery{deliveryModel[0]}, deliveryModel[2])
		payments := append([]*models.Payment{paymentModel[0]}, paymentModel[2])
		items := append(itemsModel[0], itemsModel[2]...)
		mockCache.On("Load", mock.Anything)
		mockStorage.On("GetAllOrders", mock.Anything).Return(orders, nil)
		mockStorage.On("GetALLDeliveries", mock.Anything).Return(deliveries, nil)
		mockStorage.On("GetAllPayments", mock.Anything).Return(payments, nil)
		mockStorage.On("GetAllItems", mock.Anything).Return(items, nil)

		err := service.WarmUpCache(context.Background())
		require.NoError(t, err, "expected no error")

		mockStorage.AssertCalled(t, "GetAllOrders", mock.Anything)
		mockStorage.AssertCalled(t, "GetALLDeliveries", mock.Anything)
		mockStorage.AssertCalled(t, "GetAllPayments", mock.Anything)
		mockStorage.AssertCalled(t, "GetAllItems", mock.Anything)

		mockCache.AssertCalled(t, "Load", mock.Anything)
	})

	t.Run("get modelToService error", func(t *testing.T) {
		mockStorage := new(mocksS.Storage)
		mockCache := new(mocksC.Cacher)
		service := New(mockStorage, mockCache, log)

		mockCache.On("Load", mock.Anything)
		mockStorage.On("GetAllOrders", mock.Anything).Return(nil, nil)
		mockStorage.On("GetALLDeliveries", mock.Anything).Return(nil, nil)
		mockStorage.On("GetAllPayments", mock.Anything).Return(nil, nil)
		mockStorage.On("GetAllItems", mock.Anything).Return(nil, nil)

		err := service.WarmUpCache(context.Background())
		require.NoError(t, err, "expected error")

		mockStorage.AssertCalled(t, "GetAllOrders", mock.Anything)
		mockStorage.AssertCalled(t, "GetALLDeliveries", mock.Anything)
		mockStorage.AssertCalled(t, "GetAllPayments", mock.Anything)
		mockStorage.AssertCalled(t, "GetAllItems", mock.Anything)

		mockCache.AssertCalled(t, "Load", mock.Anything)
	})

}

var correctTime, _ = time.Parse(time.RFC3339, "2021-11-26T06:12:07Z")
var (
	ordersService = []*Order{
		{
			OrderUID:    "b563feb7b2b84b6test",
			TrackNumber: "WBILMTESTTRACK",
			Entry:       "WBIL",
			Delivery: &Delivery{
				Name:    "Test Testov",
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: "Ploshad Mira 15",
				Region:  "Kraiot",
				Email:   "test@gmail.com",
			},
			Payment: &Payment{
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
			Items: []Item{
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
			Delivery: &Delivery{
				Name:    "Test2 Testov2",
				Phone:   "+9720000002",
				Zip:     "2639809",
				City:    "Mozkinburg",
				Address: "Ploshad Pira 15",
				Region:  "Daleko",
				Email:   "tes2t@gmail.com",
			},
			Payment: nil,
			Items: []Item{
				{
					ChrtID:      9934931,
					TrackNumber: "WBILMTESTTRACK",
					Price:       43,
					Rid:         "ab4219087a764ae0btest2",
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
			Delivery: &Delivery{
				Name:    "Test Testov2",
				Phone:   "+9720000002",
				Zip:     "2639809",
				City:    "Kiryat Moskvin",
				Address: "Ploshad Pira 15",
				Region:  "Kayot",
				Email:   "tes2t@gmail.com",
			},
			Payment: &Payment{
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
			Items: []Item{
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
	ordersCache = []*cache.Order{
		{
			OrderUID:    "b563feb7b2b84b6test",
			TrackNumber: "WBILMTESTTRACK",
			Entry:       "WBIL",
			Delivery: &cache.Delivery{
				Name:    "Test Testov",
				Phone:   "+9720000000",
				Zip:     "2639809",
				City:    "Kiryat Mozkin",
				Address: "Ploshad Mira 15",
				Region:  "Kraiot",
				Email:   "test@gmail.com",
			},
			Payment: &cache.Payment{
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
			Items: []cache.Item{
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
			Delivery: &cache.Delivery{
				Name:    "Test2 Testov2",
				Phone:   "+9720000002",
				Zip:     "2639809",
				City:    "Mozkinburg",
				Address: "Ploshad Pira 15",
				Region:  "Daleko",
				Email:   "tes2t@gmail.com",
			},
			Payment: &cache.Payment{
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
			Delivery: &cache.Delivery{
				Name:    "Test Testov2",
				Phone:   "+9720000002",
				Zip:     "2639809",
				City:    "Kiryat Moskvin",
				Address: "Ploshad Pira 15",
				Region:  "Kayot",
				Email:   "tes2t@gmail.com",
			},
			Payment: &cache.Payment{
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
			Items: []cache.Item{
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

	ordersModel = []*models.Order{
		{
			OrderUID:          "b563feb7b2b84b6test",
			TrackNumber:       "WBILMTESTTRACK",
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
			TrackNumber:       "WBILMTESTTRACK",
			Entry:             "WBIL",
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
			OrderUID:          "b563feb7b2b84b6test2",
			TrackNumber:       "WBILMTESTTRACK",
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
			Name:     "Test2 Testov2",
			Phone:    "+9720000002",
			Zip:      "2639809",
			City:     "Mozkinburg",
			Address:  "Ploshad Pira 15",
			Region:   "Daleko",
			Email:    "tes2t@gmail.com",
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
		nil,
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
				ChrtID:     9934931,
				Price:      43,
				RID:        "ab4219087a764ae0btest2",
				Name:       "Mascaras",
				Sale:       0,
				Size:       "1",
				TotalPrice: 43,
				NmID:       238912,
				Brand:      "Vivienne Sabor",
				Status:     200,
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
