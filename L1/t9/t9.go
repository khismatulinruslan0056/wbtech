package main

import (
	"fmt"
)

//func main() {
//	sl := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
//	ch1, ch2 := make(chan int), make(chan int)
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		for n := range ch1 {
//			ch2 <- n * n
//		}
//		close(ch2)
//
//	}()
//
//	wg.Add(1)
//	go func() {
//		defer wg.Done()
//		for n := range ch2 {
//			fmt.Println(n)
//		}
//	}()
//
//	for _, n := range sl {
//		ch1 <- n
//	}
//	close(ch1)
//
//	wg.Wait()
//}

func main() {
	sl := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ch1, ch2 := make(chan int), make(chan int)
	go func() {
		for n := range ch1 {
			ch2 <- n * n
		}
		close(ch2)

	}()

	go func() {
		for _, n := range sl {
			ch1 <- n
		}
		close(ch1)

	}()

	for n := range ch2 {
		fmt.Println(n)
	}

}
