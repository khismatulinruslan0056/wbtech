package main

import (
	"fmt"
	"sync"
)

type cache struct {
	stor map[int]int
	mu   sync.RWMutex
}

func NewCache() *cache {
	return &cache{stor: make(map[int]int)}
}
func (c *cache) Put(k int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stor[k]++
}

func (c *cache) Get(k int) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.stor[k]
	return v, ok
}

func main() {
	c := NewCache()
	var wg sync.WaitGroup
	for _, n := range sl {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Put(n)
		}()
	}
	wg.Wait()
	expM := map[int]int{
		1:  3,
		2:  3,
		4:  5,
		5:  1,
		12: 1,
		8:  1,
		0:  2,
		6:  1,
		7:  1,
		52: 1,
		43: 1,
	}

	for k, v := range expM {
		val, ok := c.Get(k)
		if !ok || val != v {
			fmt.Println("incorrect concurrent work with map")
			break
		}
	}

	sMap := loadSyncMap()
	for k, v := range expM {
		valI, ok := sMap.Load(k)
		val := valI.(int)
		if !ok || val != v {
			fmt.Println("incorrect concurrent work with syncmap")
			break
		}
	}

}

var sl = []int{1, 2, 4, 5, 12, 1, 4, 2, 4, 8, 0, 6, 4, 2, 4, 7, 52, 1, 43, 0}

func loadSyncMap() *sync.Map {
	var m sync.Map
	var wg sync.WaitGroup

	for _, n := range sl {
		wg.Add(1)
		go func(num int) {
			defer wg.Done()
			for {
				val, ok := m.Load(num)

				if !ok {
					actual, actOk := m.LoadOrStore(num, 1)
					if !actOk {
						return
					}
					val = actual
				}

				oldVal := val.(int)

				newVal := oldVal + 1
				if m.CompareAndSwap(num, val, newVal) {
					return
				}
			}

		}(n)
	}

	wg.Wait()
	return &m
}

// запуск  go run t7.go -race
