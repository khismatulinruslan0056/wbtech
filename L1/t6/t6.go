package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	ch := make(chan int)
	var wg sync.WaitGroup
	//ctx, cancel := context.WithCancel(context.Background()) // простая отмена контекста
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // отмена по таймауту
	//ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT) // по сис вызову
	defer cancel()

	wg.Add(1)
	go goroutineStopCtx(ctx, &wg, ch)

	for n := range ch {
		fmt.Println(n)
		time.Sleep(150 * time.Millisecond)
	}

	wg.Wait()

	fmt.Println()
	var wgCh sync.WaitGroup
	ch = make(chan int)
	chDone := make(chan struct{})
	wgCh.Add(1)
	go goroutineStopChan(&wg, ch, chDone)
	for n := range ch {
		fmt.Println(n)
		if n > 25 {
			fmt.Println("stop")
			chDone <- struct{}{}
			break
		}
		time.Sleep(150 * time.Millisecond)
	}

	var wgSig sync.WaitGroup
	ch = make(chan int)
	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGINT)
	wgSig.Add(1)
	go goroutineStopSig(&wgSig, ch)
LOOP:
	for {
		select {
		case ch <- rand.IntN(100):
		case <-chSig:
			close(ch)
			break LOOP
		}
	}
	wgSig.Done()

	var flag atomic.Bool
	var wgFlag sync.WaitGroup
	wgFlag.Add(1)
	go goroutineStopFlagAtomic(&wgFlag, &flag)

	time.Sleep(2 * time.Second)
	flag.Store(true)

	wgFlag.Wait()

	var wgItem sync.WaitGroup
	chItem := make(chan *item)
	wgItem.Add(1)
	go goroutineStopOnCondition(&wgItem, chItem)
	names := []string{"McLaren", "Mercedes AMG F1", "Scuderia Ferrari", "Red Bull", "Williams",
		"Racing Bulls", "Aston Martin", "Kick Sauber", "Haas", "Alpine"}
	for i, name := range names {
		if i == 5 {
			chItem <- nil
			break
		}

		chItem <- &item{name: name}
	}

	close(chItem)
	wgItem.Wait()

	var wgExit sync.WaitGroup
	wgExit.Add(1)
	go goroutineStopGoExit(&wgExit)
	wgExit.Wait()

}

func goroutineStopCtx(ctx context.Context, wg *sync.WaitGroup, ch chan<- int) {
	defer wg.Done()
	defer close(ch)
	for {
		select {
		case <-ctx.Done():
			return
		case ch <- rand.IntN(100):
		}
	}

}

func goroutineStopChan(wg *sync.WaitGroup, ch chan<- int, chDone <-chan struct{}) {
	defer wg.Done()
	defer close(ch)
	for {
		select {
		case <-chDone:
			return
		case ch <- rand.IntN(100):
		}
	}

}

func goroutineStopSig(wg *sync.WaitGroup, ch <-chan int) {
	defer wg.Done()
	for {
		for n := range ch {
			fmt.Println(n)
			time.Sleep(100 * time.Millisecond)
		}
	}

}

func goroutineStopFlagAtomic(wg *sync.WaitGroup, flag *atomic.Bool) {
	defer wg.Done()
	for {
		if flag.Load() {
			fmt.Println("Остановка по флагу")
			return
		}
		fmt.Println(rand.IntN(100))
		time.Sleep(100 * time.Millisecond)
	}

}

func goroutineStopOnCondition(wg *sync.WaitGroup, items chan *item) {
	defer wg.Done()
	for it := range items {
		if it == nil {
			fmt.Println("Пустой айтем прекращаем работу")
			return
		}
		fmt.Println(it.name)
		time.Sleep(100 * time.Millisecond)
	}

}

func goroutineStopGoExit(wg *sync.WaitGroup) {
	defer wg.Done()
	defer fmt.Println("Горутина завершилась через Goexit")

	for a := 0; a < 10; a++ {
		if a+1 == 5 {
			runtime.Goexit()
		}
		fmt.Println(a + 1)
	}

	fmt.Println("Горутина не завершилась через Goexit")
}

type item struct {
	name string
}
