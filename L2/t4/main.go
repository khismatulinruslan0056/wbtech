package main

func main() {
	ch := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
		}
	}()
	for n := range ch {
		println(n)
	}
}

// Вывод
// 0
// 1
// 2
// 3
// 4
// 5
// 6
// 7
// 8
// 9
// fatal error: all goroutines are asleep - deadlock!
// это происходит потому что горутина, пишущая в канал, после записи не закрыла его,
// main горутина блокируется на чтении и ждет следующего значения, а тк его никто не шлет и нет сигнала о закрытии
// канала, мы ловим deadlock
