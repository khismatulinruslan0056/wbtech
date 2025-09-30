package importer

import "text/template"

const jsonTemplate = `{
  "order_uid": "{{.OrderUID}}",
  "track_number": "{{.TrackNumber}}",
  "entry": "{{.Entry}}",
  "delivery": {
    "name": "{{.Delivery.Name}}",
    "phone": "{{.Delivery.Phone}}",
    "zip": "{{.Delivery.Zip}}",
    "city": "{{.Delivery.City}}",
    "address": "{{.Delivery.Address}}",
    "region": "{{.Delivery.Region}}",
    "email": "{{.Delivery.Email}}"
  },
  "payment": {
    "transaction": "{{.Payment.Transaction}}",
    "request_id": "{{.Payment.RequestID}}",
    "currency": "{{.Payment.Currency}}",
    "provider": "{{.Payment.Provider}}",
    "amount": {{.Payment.Amount}},
    "payment_dt": {{.Payment.PaymentDT}},
    "bank": "{{.Payment.Bank}}",
    "delivery_cost": {{.Payment.DeliveryCost}},
    "goods_total": {{.Payment.GoodsTotal}},
    "custom_fee": {{.Payment.CustomFee}}
  },
  "items": [{{range $index, $item := .Items}}{{if $index}},{{end}}
    {
      "chrt_id": {{$item.ChrtID}},
      "track_number": "{{$item.TrackNumber}}",
      "price": {{$item.Price}},
      "rid": "{{$item.Rid}}",
      "name": "{{$item.Name}}",
      "sale": {{$item.Sale}},
      "size": "{{$item.Size}}",
      "total_price": {{$item.TotalPrice}},
      "nm_id": {{$item.NmID}},
      "brand": "{{$item.Brand}}",
      "status": {{$item.Status}}
    }{{end}}
  ],
  "locale": "{{.Locale}}",
  "internal_signature": "{{.InternalSignature}}",
  "customer_id": "{{.CustomerID}}",
  "delivery_service": "{{.DeliveryService}}",
  "shardkey": "{{.ShardKey}}",
  "sm_id": {{.SmID}},
  "date_created": "{{.DateCreated}}",
  "oof_shard": "{{.OofShard}}"
}`

type Order struct {
	OrderUID          string   `json:"order_uid"`
	TrackNumber       string   `json:"track_number"`
	Entry             string   `json:"entry"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Item   `json:"items"`
	Locale            string   `json:"locale"`
	InternalSignature string   `json:"internal_signature"`
	CustomerID        string   `json:"customer_id"`
	DeliveryService   string   `json:"delivery_service"`
	ShardKey          string   `json:"shardkey"`
	SmID              int      `json:"sm_id"`
	DateCreated       string   `json:"date_created"`
	OofShard          string   `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDT    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func NewTemplate() *template.Template {
	tpl := template.Must(template.New("json").Parse(jsonTemplate))

	return tpl
}
