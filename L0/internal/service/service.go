package service

import (
	"L0/internal/cache"
	"L0/internal/models"
	"L0/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/cenkalti/backoff/v4"
	"golang.org/x/sync/errgroup"
)

var (
	IncorrectServiceOrder = errors.New("incorrect api struct order")
	IncorrectCacheOrder   = errors.New("incorrect cache struct order")
	IncorrectModelOrder   = errors.New("incorrect model struct order")
	RetryFailed           = errors.New("failed to execute transaction after multiple retries")
)

type Service struct {
	st    storage.Storage
	cache cache.Cacher
	log   *slog.Logger
}

func New(s storage.Storage, c cache.Cacher, log *slog.Logger) *Service {
	return &Service{st: s, cache: c, log: log}
}

func (s *Service) Add(ctx context.Context, orderService *Order) (string, error) {
	const op = "Service.Add"
	orderCache, err := serviceToCache(orderService)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	order, delivery, payment, items, err := serviceToModel(orderService)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	operation := func() error {

		tx, err := s.st.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		var opErr error
		defer func() {
			if opErr != nil {
				if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
					s.log.Error("rollback execution problem",
						"op", op,
						"err", rbErr)
				}
			}
		}()

		storageWithTx := s.st.WithTx(tx)
		_, opErr = storageWithTx.AddOrder(ctx, order)
		if opErr != nil {
			return fmt.Errorf("%s: %w", op, opErr)
		}

		_, opErr = storageWithTx.AddDelivery(ctx, delivery)
		if opErr != nil {
			return fmt.Errorf("%s: %w", op, opErr)
		}
		_, opErr = storageWithTx.AddPayment(ctx, payment)

		if opErr != nil {
			return fmt.Errorf("%s: %w", op, opErr)
		}
		_, opErr = storageWithTx.AddItems(ctx, items)

		if opErr != nil {
			return fmt.Errorf("%s: %w", op, opErr)
		}
		if opErr = tx.Commit(); opErr != nil {

			return fmt.Errorf("%s: %w", op, opErr)
		}
		return opErr
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 15 * time.Second

	if err = backoff.Retry(operation, bo); err != nil {
		return "", fmt.Errorf("%s:  %w", op, RetryFailed)
	}

	s.cache.Put(orderCache)

	return order.OrderUID, nil
}

func (s *Service) Get(ctx context.Context, orderID string) (*Order, error) {
	const op = "Service.Get"
	var (
		order    *models.Order
		delivery *models.Delivery
		payment  *models.Payment
		items    []*models.Item
		err      error
	)

	if orderCache, ok := s.cache.Get(orderID); ok {
		orderService, err := cacheToService(orderCache)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return orderService, nil
	}

	g, gCtx := errgroup.WithContext(ctx)
	run := func(fns []func() error) {
		for _, fn := range fns {
			g.Go(func() error {
				if err = fn(); err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
				return nil
			})
		}
	}

	run([]func() error{
		func() error {
			order, err = s.st.GetOrder(gCtx, orderID)
			return err
		},
		func() error {
			delivery, err = s.st.GetDelivery(gCtx, orderID)
			return err
		},
		func() error {
			payment, err = s.st.GetPayment(gCtx, orderID)
			return err
		},
		func() error {
			items, err = s.st.GetItemsByID(gCtx, orderID)
			return err
		}})

	if err = g.Wait(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	orderService, err := modelToService(order, delivery, payment, items)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	orderCache, err := serviceToCache(orderService)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	s.cache.Put(orderCache)

	return orderService, nil
}

func modelToService(order *models.Order, delivery *models.Delivery, payment *models.Payment, items []*models.Item) (*Order, error) {
	const op = "modelToService"

	if order == nil || delivery == nil || payment == nil || len(items) == 0 {
		return nil, fmt.Errorf("%s: %w", op, IncorrectModelOrder)
	}

	orderService := &Order{
		OrderUID:          order.OrderUID,
		TrackNumber:       order.TrackNumber,
		Entry:             order.Entry,
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerID:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		Shardkey:          order.Shardkey,
		SmID:              order.SmID,
		DateCreated:       order.DateCreated,
		OofShard:          order.OofShard,
	}

	deliveryService := &Delivery{
		Name:    delivery.Name,
		Phone:   delivery.Phone,
		Zip:     delivery.Zip,
		City:    delivery.City,
		Address: delivery.Address,
		Region:  delivery.Region,
		Email:   delivery.Email,
	}

	paymentService := &Payment{
		Transaction:  payment.Transaction,
		RequestID:    payment.RequestID,
		Currency:     payment.Currency,
		Provider:     payment.Provider,
		Amount:       payment.Amount,
		PaymentDt:    payment.PaymentDT,
		Bank:         payment.Bank,
		DeliveryCost: payment.DeliveryCost,
		GoodsTotal:   payment.GoodsTotal,
		CustomFee:    payment.CustomFee,
	}

	itemsService := make([]Item, 0, len(items))

	for i := 0; i < len(items); i++ {
		itemService := Item{
			ChrtID:      items[i].ChrtID,
			TrackNumber: orderService.TrackNumber,
			Price:       items[i].Price,
			Rid:         items[i].RID,
			Name:        items[i].Name,
			Sale:        items[i].Sale,
			Size:        items[i].Size,
			TotalPrice:  items[i].TotalPrice,
			NmID:        items[i].NmID,
			Brand:       items[i].Brand,
			Status:      items[i].Status,
		}

		itemsService = append(itemsService, itemService)
	}

	orderService.Delivery = deliveryService
	orderService.Payment = paymentService
	orderService.Items = itemsService

	return orderService, nil
}

func serviceToModel(orderService *Order) (*models.Order, *models.Delivery, *models.Payment, []*models.Item, error) {
	const op = "serviceToModel"
	if orderService == nil || orderService.Payment == nil || orderService.Delivery == nil || len(orderService.Items) == 0 {
		return nil, nil, nil, nil, fmt.Errorf("%s: %w", op, IncorrectServiceOrder)
	}
	order := &models.Order{
		OrderUID:          orderService.OrderUID,
		TrackNumber:       orderService.TrackNumber,
		Entry:             orderService.Entry,
		Locale:            orderService.Locale,
		InternalSignature: orderService.InternalSignature,
		CustomerID:        orderService.CustomerID,
		DeliveryService:   orderService.DeliveryService,
		Shardkey:          orderService.Shardkey,
		SmID:              orderService.SmID,
		DateCreated:       orderService.DateCreated,
		OofShard:          orderService.OofShard,
	}
	delivery := &models.Delivery{
		OrderUID: orderService.OrderUID,
		Name:     orderService.Delivery.Name,
		Phone:    orderService.Delivery.Phone,
		Zip:      orderService.Delivery.Zip,
		City:     orderService.Delivery.City,
		Address:  orderService.Delivery.Address,
		Region:   orderService.Delivery.Region,
		Email:    orderService.Delivery.Email,
	}

	payment := &models.Payment{
		OrderUID:     orderService.OrderUID,
		Transaction:  orderService.Payment.Transaction,
		RequestID:    orderService.Payment.RequestID,
		Currency:     orderService.Payment.Currency,
		Provider:     orderService.Payment.Provider,
		Amount:       orderService.Payment.Amount,
		PaymentDT:    orderService.Payment.PaymentDt,
		Bank:         orderService.Payment.Bank,
		DeliveryCost: orderService.Payment.DeliveryCost,
		GoodsTotal:   orderService.Payment.GoodsTotal,
		CustomFee:    orderService.Payment.CustomFee,
	}
	items := make([]*models.Item, 0, len(orderService.Items))

	for i := 0; i < len(orderService.Items); i++ {
		item := &models.Item{
			OrderUID:   orderService.OrderUID,
			ChrtID:     orderService.Items[i].ChrtID,
			Price:      orderService.Items[i].Price,
			RID:        orderService.Items[i].Rid,
			Name:       orderService.Items[i].Name,
			Sale:       orderService.Items[i].Sale,
			Size:       orderService.Items[i].Size,
			TotalPrice: orderService.Items[i].TotalPrice,
			NmID:       orderService.Items[i].NmID,
			Brand:      orderService.Items[i].Brand,
			Status:     orderService.Items[i].Status,
		}
		items = append(items, item)
	}

	return order, delivery, payment, items, nil
}

func serviceToCache(order *Order) (*cache.Order, error) {
	const op = "ServiceToCache"

	if order == nil || order.Delivery == nil || order.Payment == nil || len(order.Items) == 0 {
		return nil, fmt.Errorf("%s: %w", op, IncorrectServiceOrder)
	}

	orderDTO := &cache.Order{
		OrderUID:          order.OrderUID,
		TrackNumber:       order.TrackNumber,
		Entry:             order.Entry,
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerID:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		Shardkey:          order.Shardkey,
		SmID:              order.SmID,
		DateCreated:       order.DateCreated,
		OofShard:          order.OofShard,
	}

	deliveryDTO := &cache.Delivery{
		Name:    order.Delivery.Name,
		Phone:   order.Delivery.Phone,
		Zip:     order.Delivery.Zip,
		City:    order.Delivery.City,
		Address: order.Delivery.Address,
		Region:  order.Delivery.Region,
		Email:   order.Delivery.Email,
	}

	paymentDTO := &cache.Payment{
		Transaction:  order.Payment.Transaction,
		RequestID:    order.Payment.RequestID,
		Currency:     order.Payment.Currency,
		Provider:     order.Payment.Provider,
		Amount:       order.Payment.Amount,
		PaymentDt:    order.Payment.PaymentDt,
		Bank:         order.Payment.Bank,
		DeliveryCost: order.Payment.DeliveryCost,
		GoodsTotal:   order.Payment.GoodsTotal,
		CustomFee:    order.Payment.CustomFee,
	}

	itemsDTO := make([]cache.Item, 0, len(order.Items))

	for i := 0; i < len(order.Items); i++ {
		itemDTO := cache.Item{
			ChrtID:      order.Items[i].ChrtID,
			TrackNumber: order.Items[i].TrackNumber,
			Price:       order.Items[i].Price,
			Rid:         order.Items[i].Rid,
			Name:        order.Items[i].Name,
			Sale:        order.Items[i].Sale,
			Size:        order.Items[i].Size,
			TotalPrice:  order.Items[i].TotalPrice,
			NmID:        order.Items[i].NmID,
			Brand:       order.Items[i].Brand,
			Status:      order.Items[i].Status,
		}
		itemsDTO = append(itemsDTO, itemDTO)
	}

	orderDTO.Delivery = deliveryDTO
	orderDTO.Payment = paymentDTO
	orderDTO.Items = itemsDTO

	return orderDTO, nil
}

func cacheToService(order *cache.Order) (*Order, error) {
	const op = "cacheToService"
	if order == nil || order.Delivery == nil || order.Payment == nil || len(order.Items) == 0 {
		return nil, fmt.Errorf("%s: %w", op, IncorrectCacheOrder)
	}

	orderService := &Order{
		OrderUID:          order.OrderUID,
		TrackNumber:       order.TrackNumber,
		Entry:             order.Entry,
		Locale:            order.Locale,
		InternalSignature: order.InternalSignature,
		CustomerID:        order.CustomerID,
		DeliveryService:   order.DeliveryService,
		Shardkey:          order.Shardkey,
		SmID:              order.SmID,
		DateCreated:       order.DateCreated,
		OofShard:          order.OofShard,
	}

	deliveryService := &Delivery{
		Name:    order.Delivery.Name,
		Phone:   order.Delivery.Phone,
		Zip:     order.Delivery.Zip,
		City:    order.Delivery.City,
		Address: order.Delivery.Address,
		Region:  order.Delivery.Region,
		Email:   order.Delivery.Email,
	}

	paymentService := &Payment{
		Transaction:  order.Payment.Transaction,
		RequestID:    order.Payment.RequestID,
		Currency:     order.Payment.Currency,
		Provider:     order.Payment.Provider,
		Amount:       order.Payment.Amount,
		PaymentDt:    order.Payment.PaymentDt,
		Bank:         order.Payment.Bank,
		DeliveryCost: order.Payment.DeliveryCost,
		GoodsTotal:   order.Payment.GoodsTotal,
		CustomFee:    order.Payment.CustomFee,
	}

	itemsService := make([]Item, 0, len(order.Items))

	for i := 0; i < len(order.Items); i++ {
		itemService := Item{
			ChrtID:      order.Items[i].ChrtID,
			TrackNumber: order.Items[i].TrackNumber,
			Price:       order.Items[i].Price,
			Rid:         order.Items[i].Rid,
			Name:        order.Items[i].Name,
			Sale:        order.Items[i].Sale,
			Size:        order.Items[i].Size,
			TotalPrice:  order.Items[i].TotalPrice,
			NmID:        order.Items[i].NmID,
			Brand:       order.Items[i].Brand,
			Status:      order.Items[i].Status,
		}
		itemsService = append(itemsService, itemService)
	}

	orderService.Delivery = deliveryService
	orderService.Payment = paymentService
	orderService.Items = itemsService

	return orderService, nil
}

func (s *Service) WarmUpCache(ctx context.Context) error {
	const op = "WarmUpCache"

	var (
		orders     []*models.Order
		deliveries []*models.Delivery
		payments   []*models.Payment
		items      []*models.Item
		err        error
	)

	g, gCtx := errgroup.WithContext(ctx)
	run := func(fns []func() error) {
		for _, fn := range fns {
			g.Go(func() error {
				if err = fn(); err != nil {
					return fmt.Errorf("%s: %w", op, err)
				}
				return nil
			})
		}
	}

	run([]func() error{
		func() error {
			orders, err = s.st.GetAllOrders(gCtx)
			return err
		},
		func() error {
			deliveries, err = s.st.GetALLDeliveries(gCtx)
			return err
		},
		func() error {
			payments, err = s.st.GetAllPayments(gCtx)
			return err
		},
		func() error {
			items, err = s.st.GetAllItems(gCtx)
			return err
		}})

	if err = g.Wait(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ordersMap := make(map[string]*models.Order, len(orders))
	deliveriesMap := make(map[string]*models.Delivery, len(deliveries))
	paymentsMap := make(map[string]*models.Payment, len(payments))
	itemsMap := make(map[string][]*models.Item, len(items))
	ordersMapToCache := make(map[string]*cache.Order, len(items))
	for _, order := range ordersMap {
		ordersMap[order.OrderUID] = order
	}

	for _, delivery := range deliveries {
		deliveriesMap[delivery.OrderUID] = delivery
	}

	for _, payment := range payments {
		paymentsMap[payment.OrderUID] = payment
	}

	for _, item := range items {
		itemsMap[item.OrderUID] = append(itemsMap[item.OrderUID], item)
	}

	for orderID, order := range ordersMap {
		orderService, err := modelToService(order, deliveriesMap[orderID], paymentsMap[orderID], itemsMap[orderID])
		if err != nil {
			s.log.Error("incorrect struct order/delivery/payment/items")
			continue
		}
		orderToCache, _ := serviceToCache(orderService)

		ordersMapToCache[orderID] = orderToCache
	}
	s.cache.Load(ordersMapToCache)

	return nil
}
