package main

import (
	"fmt"
	"time"
)

func main() {

	sl := []int{1, 12, 32, 54, 12, 87, 12, 80}
	numworkers := 2
	ch := make(chan int)
	for i := 0; i < numworkers; i++ {
		go worker(ch)
	}

	for _, n := range sl {
		ch <- n
	}
	close(ch)

	sleep(5 * time.Second)

	fmt.Println("Успешно")
}

func worker(ch chan int) {
	for n := range ch {
		fmt.Println(n)
	}
}

func sleep(d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	<-timer.C
}
