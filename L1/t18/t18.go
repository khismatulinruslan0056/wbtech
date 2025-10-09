package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type counter struct {
	count int64
	mu    sync.Mutex
}

func (c *counter) Increment() {
	c.count++
}

func (c *counter) IncrementWithMutex() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count++
}

func (c *counter) IncrementAtomic() {
	atomic.AddInt64(&c.count, 1)
}

func (c *counter) Reset() {
	c.count = 0
}

func main() {
	c := &counter{}
	wg := sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Increment()
		}()
	}

	wg.Wait()
	fmt.Println(c.count, c.count == 1000)
	c.Reset()

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.IncrementWithMutex()
		}()
	}

	wg.Wait()
	fmt.Println(c.count, c.count == 1000)
	c.Reset()

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.IncrementAtomic()
		}()
	}

	wg.Wait()
	fmt.Println(c.count, c.count == 1000)
	c.Reset()

}
