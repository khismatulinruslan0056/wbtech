package cache

import (
	"time"
)

type Order struct {
	OrderUID          string
	TrackNumber       string
	Entry             string
	Delivery          *Delivery
	Payment           *Payment
	Items             []Item
	Locale            string
	InternalSignature string
	CustomerID        string
	DeliveryService   string
	Shardkey          string
	SmID              int
	DateCreated       time.Time
	OofShard          string
}

type Item struct {
	ChrtID      int
	TrackNumber string
	Price       int
	Rid         string
	Name        string
	Sale        int
	Size        string
	TotalPrice  int
	NmID        int
	Brand       string
	Status      int
}

type Payment struct {
	Transaction  string
	RequestID    string
	Currency     string
	Provider     string
	Amount       int
	PaymentDt    int
	Bank         string
	DeliveryCost int
	GoodsTotal   int
	CustomFee    int
}

type Delivery struct {
	Name    string
	Phone   string
	Zip     string
	City    string
	Address string
	Region  string
	Email   string
}
