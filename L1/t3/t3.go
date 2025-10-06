package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

var (
	numWorkers = flag.Int("numWorkers", 1, "put number of workers")
	timeWork   = flag.Duration("timeWork", 10*time.Second, "put the program's running time")
)

func worker(ctx context.Context, wg *sync.WaitGroup, iWorker int, ch <-chan int) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case n, ok := <-ch:
			if !ok {
				return
			}
			fmt.Printf("Worker %d: msg - %d.\n", iWorker, n)
		}
	}
}

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), *timeWork)
	defer cancel()
	ch := make(chan int)
	var wg sync.WaitGroup
	defer wg.Wait()
	defer close(ch)
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, &wg, i+1, ch)
	}

	for {
		select {
		case <-ctx.Done():
			//close(ch)
			//wg.Wait()
			return
		case ch <- rand.IntN(100):
			time.Sleep(250 * time.Millisecond)
		}
	}
}

// запуск
// go run t3.go -numWorkers=4 -timeWork=15s
