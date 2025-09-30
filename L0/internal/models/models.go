package models

import "time"

type Delivery struct {
	ID       int
	OrderUID string
	Name     string
	Phone    string
	Zip      string
	City     string
	Address  string
	Region   string
	Email    string
}

type Payment struct {
	ID           int
	OrderUID     string
	Transaction  string
	RequestID    string
	Currency     string
	Provider     string
	Amount       int
	PaymentDT    int
	Bank         string
	DeliveryCost int
	GoodsTotal   int
	CustomFee    int
}

type Item struct {
	ID         int
	OrderUID   string
	ChrtID     int
	Price      int
	RID        string
	Name       string
	Sale       int
	Size       string
	TotalPrice int
	NmID       int
	Brand      string
	Status     int
}

type Order struct {
	OrderUID          string
	TrackNumber       string
	Entry             string
	Locale            string
	InternalSignature string
	CustomerID        string
	DeliveryService   string
	Shardkey          string
	SmID              int
	DateCreated       time.Time
	OofShard          string
}
