package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand/v2"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	numWorkers = flag.Int("numWorkers", 1, "put number of workers")
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	// также можно использовать
	//chSignal := make(chan os.Signal)
	//signal.Notify(chSignal, syscall.SIGINT)
	//...
	//...
	//<-chSignal

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
// go run t4.go -numWorkers=4
