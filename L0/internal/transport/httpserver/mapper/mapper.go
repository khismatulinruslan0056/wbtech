package mapper

import (
	"L0/internal/service"
	"L0/internal/transport/dto"
	"errors"
	"fmt"
)

var (
	IncorrectDTOOrder     = errors.New("incorrect DTO struct order")
	IncorrectServiceOrder = errors.New("incorrect api struct order")
)

func ServiceToDTO(order *service.Order) (*dto.Order, error) {
	const op = "ServiceToDTO"
	if order == nil || order.Delivery == nil || order.Payment == nil || len(order.Items) == 0 {
		return nil, fmt.Errorf("%s: %w", op, IncorrectServiceOrder)
	}
	orderDTO := &dto.Order{
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

	deliveryDTO := &dto.Delivery{
		Name:    order.Delivery.Name,
		Phone:   order.Delivery.Phone,
		Zip:     order.Delivery.Zip,
		City:    order.Delivery.City,
		Address: order.Delivery.Address,
		Region:  order.Delivery.Region,
		Email:   order.Delivery.Email,
	}

	paymentDTO := &dto.Payment{
		Transaction:  order.Payment.Transaction,
		RequestID:    order.Payment.RequestID,
		Currency:     order.Payment.Currency,
		Provider:     order.Payment.Provider,
		Amount:       order.Payment.Amount,
		PaymentDT:    order.Payment.PaymentDt,
		Bank:         order.Payment.Bank,
		DeliveryCost: order.Payment.DeliveryCost,
		GoodsTotal:   order.Payment.GoodsTotal,
		CustomFee:    order.Payment.CustomFee,
	}

	itemsDTO := make([]dto.Item, 0, len(order.Items))

	for i := 0; i < len(order.Items); i++ {
		itemDTO := dto.Item{
			ChrtID:      order.Items[i].ChrtID,
			TrackNumber: order.Items[i].TrackNumber,
			Price:       order.Items[i].Price,
			RID:         order.Items[i].Rid,
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

func DTOToService(order *dto.Order) (*service.Order, error) {
	const op = "DTOToService"
	if order == nil || order.Delivery == nil || order.Payment == nil || len(order.Items) == 0 {
		return nil, fmt.Errorf("%s: %w", op, IncorrectDTOOrder)
	}

	orderService := &service.Order{
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

	deliveryService := &service.Delivery{
		Name:    order.Delivery.Name,
		Phone:   order.Delivery.Phone,
		Zip:     order.Delivery.Zip,
		City:    order.Delivery.City,
		Address: order.Delivery.Address,
		Region:  order.Delivery.Region,
		Email:   order.Delivery.Email,
	}

	paymentService := &service.Payment{
		Transaction:  order.Payment.Transaction,
		RequestID:    order.Payment.RequestID,
		Currency:     order.Payment.Currency,
		Provider:     order.Payment.Provider,
		Amount:       order.Payment.Amount,
		PaymentDt:    order.Payment.PaymentDT,
		Bank:         order.Payment.Bank,
		DeliveryCost: order.Payment.DeliveryCost,
		GoodsTotal:   order.Payment.GoodsTotal,
		CustomFee:    order.Payment.CustomFee,
	}

	itemsService := make([]service.Item, 0, len(order.Items))

	for i := 0; i < len(order.Items); i++ {
		itemService := service.Item{
			ChrtID:      order.Items[i].ChrtID,
			TrackNumber: order.Items[i].TrackNumber,
			Price:       order.Items[i].Price,
			Rid:         order.Items[i].RID,
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

func ServiceToPublicDTO(order *service.Order) (*dto.PublicOrder, error) {
	const op = "ServiceToPublicDTO"
	if order == nil || order.Delivery == nil || order.Payment == nil || len(order.Items) == 0 {
		return nil, fmt.Errorf("%s: %w", op, IncorrectServiceOrder)
	}
	items := make([]dto.PublicItem, 0, len(order.Items))
	for _, it := range order.Items {
		items = append(items, dto.PublicItem{
			it.NmID,
			it.Name,
			it.Brand,
			it.Size,
			it.Price,
			it.Sale,
			it.TotalPrice,
		})
	}
	return &dto.PublicOrder{
		OrderUID:        order.OrderUID,
		DateCreated:     order.DateCreated,
		Currency:        order.Payment.Currency,
		Amount:          order.Payment.Amount,
		DeliveryCost:    order.Payment.DeliveryCost,
		GoodsTotal:      order.Payment.GoodsTotal,
		DeliveryService: order.DeliveryService,
		Items:           items,
	}, nil
}
