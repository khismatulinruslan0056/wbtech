package dto

import (
	"time"
)

type ErrorResponse struct {
	ErrMsg string `json:"error"`
}

func (er *ErrorResponse) Error() string {
	return er.ErrMsg
}

type Delivery struct {
	Name    string `json:"name" validate:"required"`
	Phone   string `json:"phone" validate:"required,e164"`
	Zip     string `json:"zip" validate:"required"`
	City    string `json:"city" validate:"required"`
	Address string `json:"address" validate:"required"`
	Region  string `json:"region"`
	Email   string `json:"email" validate:"required,email"`
}

type Payment struct {
	Transaction  string `json:"transaction" validate:"required"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency" validate:"required,len=3,alpha"`
	Provider     string `json:"provider" validate:"required"`
	Amount       int    `json:"amount" validate:"gte=0"`
	PaymentDT    int    `json:"payment_dt" validate:"required"`
	Bank         string `json:"bank" validate:"required"`
	DeliveryCost int    `json:"delivery_cost" validate:"gte=0"`
	GoodsTotal   int    `json:"goods_total" validate:"gte=0"`
	CustomFee    int    `json:"custom_fee" validate:"gte=0"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id" validate:"required,gt=0"`
	TrackNumber string `json:"track_number" validate:"required"`
	Price       int    `json:"price" validate:"gte=0"`
	RID         string `json:"rid" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Sale        int    `json:"sale" validate:"gte=0,lte=100"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price" validate:"gte=0"`
	NmID        int    `json:"nm_id" validate:"required,gt=0"`
	Brand       string `json:"brand" validate:"required"`
	Status      int    `json:"status" validate:"required"`
}

type Order struct {
	OrderUID          string    `json:"order_uid" validate:"required"`
	TrackNumber       string    `json:"track_number" validate:"required"`
	Entry             string    `json:"entry" validate:"required"`
	Delivery          *Delivery `json:"delivery" validate:"required"`
	Payment           *Payment  `json:"payment" validate:"required"`
	Items             []Item    `json:"items" validate:"required,dive"`
	Locale            string    `json:"locale" validate:"required,len=2,alpha"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id" validate:"required"`
	DeliveryService   string    `json:"delivery_service" validate:"required"`
	Shardkey          string    `json:"shardkey" validate:"required"`
	SmID              int       `json:"sm_id" validate:"required,gt=0"`
	DateCreated       time.Time `json:"date_created" validate:"required"`
	OofShard          string    `json:"oof_shard" validate:"required"`
}

type PublicOrder struct {
	OrderUID        string       `json:"order_uid"`
	DateCreated     time.Time    `json:"date_created"`
	Currency        string       `json:"currency"`
	Amount          int          `json:"amount"`
	DeliveryCost    int          `json:"delivery_cost"`
	GoodsTotal      int          `json:"goods_total"`
	DeliveryService string       `json:"delivery_service,omitempty"`
	Items           []PublicItem `json:"items"`
}
type PublicItem struct {
	NmID       int    `json:"nm_id"`
	Name       string `json:"name"`
	Brand      string `json:"brand"`
	Size       string `json:"size"`
	Price      int    `json:"price"`
	Sale       int    `json:"sale" validate:"gte=0,lte=100"`
	TotalPrice int    `json:"total_price"`
}
