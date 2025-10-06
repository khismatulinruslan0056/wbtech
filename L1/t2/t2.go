package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func PowerTwo(n int) int {
	return n * n
}

func main() {
	sl := [...]int{2, 4, 6, 8, 10}
	//counter := 0
	//for _, n := range sl {
	//	go func() {
	//		counter++
	//		fmt.Print(PowerTwo(n), " ")
	//	}()
	//}
	//fmt.Println()
	//fmt.Println("counter:", counter, "\ncounter == len[sl]:", counter == len(sl))
	//time.Sleep(1 * time.Second)
	//fmt.Println()
	//counter = 0
	//for _, n := range sl {
	//	go func() {
	//		counter++
	//		fmt.Print(PowerTwo(n), " ")
	//	}()
	//}
	//fmt.Println()
	//time.Sleep(10 * time.Millisecond)
	//fmt.Println("counter:", counter, "\ncounter == len[sl]:", counter == len(sl))
	//
	//fmt.Println()
	//counter = 0
	//wg := sync.WaitGroup{}
	//for _, n := range sl {
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		counter++
	//		fmt.Print(PowerTwo(n), " ")
	//	}()
	//}
	//wg.Wait()
	//fmt.Println()
	//fmt.Println("counter:", counter, "\ncounter == len[sl]:", counter == len(sl))
	//
	//fmt.Println()
	//counter = 0
	//ch := make(chan struct{})
	//for _, n := range sl {
	//	go func() {
	//		counter++
	//		fmt.Print(PowerTwo(n), " ")
	//		ch <- struct{}{}
	//	}()
	//}
	//
	//for _, _ = range sl {
	//	<-ch
	//}
	//close(ch)
	//fmt.Println()
	//fmt.Println("counter:", counter, "\ncounter == len[sl]:", counter == len(sl))

	//numworkers := 2
	//wg := sync.WaitGroup{}
	//chData := make(chan int)
	//chRes := make(chan int)
	//for i := 0; i < numworkers; i++ {
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		for n := range chData {
	//			chRes <- PowerTwo(n)
	//		}
	//	}()
	//}
	//
	//go func() {
	//	wg.Wait()
	//	close(chRes)
	//}()
	//
	//var wgR sync.WaitGroup
	//wgR.Add(1)
	//go func() {
	//	defer wgR.Done()
	//	for n := range chRes {
	//		fmt.Print(n, " ")
	//	}
	//
	//}()
	//
	//for _, n := range sl {
	//	chData <- n
	//}
	//close(chData)
	//wgR.Wait()
	//fmt.Println()

	//numworkers := 2
	//wg := sync.WaitGroup{}
	//chData := make(chan int)
	//chRes := make(chan int)
	//ctx, cancel := context.WithCancel(context.Background())
	//for i := 0; i < numworkers; i++ {
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		for {
	//			select {
	//			case <-ctx.Done():
	//				return
	//			case n, ok := <-chData:
	//				if !ok {
	//					return
	//				}
	//				chRes <- PowerTwo(n)
	//			}
	//		}
	//	}()
	//}
	//
	//go func() {
	//	wg.Wait()
	//	close(chRes)
	//}()
	//
	//var wgR sync.WaitGroup
	//wgR.Add(1)
	//go func() {
	//	defer wgR.Done()
	//	for n := range chRes {
	//		fmt.Print(n, " ")
	//	}
	//}()
	//
	//for i, n := range sl {
	//	chData <- n
	//	if i == 3 {
	//		cancel()
	//		break
	//	}
	//}
	//
	//close(chData)
	//wgR.Wait()
	//fmt.Println()

	numworkers := 2
	wg := sync.WaitGroup{}
	chData := make(chan int)
	chRes := make(chan int)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	for i := 0; i < numworkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case n, ok := <-chData:
					if !ok {
						return
					}
					chRes <- PowerTwo(n)

				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(chRes)
	}()

	var wgR sync.WaitGroup
	wgR.Add(1)
	go func() {
		defer wgR.Done()
		for n := range chRes {
			fmt.Print(n, " ")
		}
	}()

	for i, n := range sl {
		select {
		case chData <- n:
		case <-ctx.Done():
			break
		}

		if i == 3 {
			cancel()
		}
		time.Sleep(5 * time.Millisecond)
	}

	close(chData)
	wgR.Wait()
	fmt.Println()
}
