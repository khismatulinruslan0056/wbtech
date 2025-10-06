package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	slInt := []int{1, 23, 34, 1, 5, 3, 2, 4362, 123, 67, 12, 86, 32, 95, 123, 56, 78, 34, 56, 12}
	ch := make(chan int)
	var wg sync.WaitGroup
	defer wg.Wait()
	defer close(ch)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for n := range ch {
			fmt.Println("even: ", n%2 == 0)
		}
	}()

	for _, m := range slInt {
		select {
		case <-time.After(100 * time.Millisecond):
			return
		case ch <- m:
		}
	}
}
