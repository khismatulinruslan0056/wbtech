package cache

//go:generate go run github.com/vektra/mockery/v2 --name Cacher --output ../mocks/cache --case underscore
type Cacher interface {
	Get(string) (*Order, bool)
	Put(order *Order)
	Load(orders map[string]*Order)
}
