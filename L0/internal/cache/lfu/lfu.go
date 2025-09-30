package lfu

import (
	"L0/internal/cache"
	"L0/internal/config"
	"context"
	"math"
	"sync"
	"time"
)

type Cache struct {
	mu       sync.RWMutex
	stor     map[string]*cache.Order
	freq     map[string]*Info
	capacity int
	length   int
	t        time.Duration
}

func NewCache(cfg *config.Cache) *Cache {
	const op = "lfu.cache.NewCache"
	return &Cache{
		stor:     make(map[string]*cache.Order, cfg.Capacity),
		freq:     make(map[string]*Info, cfg.Capacity),
		capacity: cfg.Capacity,
		length:   0,
		t:        cfg.TTL,
	}
}

func (c *Cache) Get(orderID string) (*cache.Order, bool) {
	const op = "lfu.cache.Get"

	c.mu.Lock()
	defer c.mu.Unlock()
	order, ok := c.stor[orderID]
	if ok {
		c.freq[orderID].count++
	}
	return order, ok
}

func (c *Cache) Put(order *cache.Order) {
	const op = "lfu.cache.Put"
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.stor[order.OrderUID]; ok {
		c.stor[order.OrderUID] = order
		c.freq[order.OrderUID].ttl = time.Now().Add(10 * time.Minute)
		return
	}
	if c.length >= c.capacity {
		c.deleteNotPopular()
	}
	c.stor[order.OrderUID] = order
	c.length++
	c.freq[order.OrderUID] = &Info{
		ttl:   time.Now().Add(10 * time.Minute),
		count: 1,
	}
}

func (c *Cache) deleteNotPopular() {
	const op = "lfu.cache.deleteNotPopular"

	id := ""
	minCount := math.MaxInt32

	for orderID, info := range c.freq {
		if info.count < minCount {
			id = orderID
			minCount = info.count
		}
	}
	if id != "" {
		c.deleteInternal(id)

	}
}

func (c *Cache) CheckTTL(ctx context.Context) {
	const op = "lfu.cache.CheckTTL"

	tick := time.NewTicker(c.t)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			keysToDelete := make([]string, 0)

			c.mu.RLock()
			for orderID, info := range c.freq {
				if time.Now().After(info.ttl) {
					keysToDelete = append(keysToDelete, orderID)
				}
			}
			c.mu.RUnlock()

			if len(keysToDelete) > 0 {
				c.mu.Lock()
				for _, orderID := range keysToDelete {
					c.deleteInternal(orderID)
				}
				c.mu.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}

}

func (c *Cache) deleteInternal(orderID string) {
	const op = "lfu.cache.deleteInternal"

	if _, ok := c.stor[orderID]; ok {
		delete(c.stor, orderID)
		delete(c.freq, orderID)
		c.length--
	}
}

func (c *Cache) Load(orders map[string]*cache.Order) {
	const op = "lfu.cache.Load"

	for _, order := range orders {
		c.Put(order)
	}
}

type Info struct {
	ttl   time.Time
	count int
}
