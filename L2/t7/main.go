package main

import (
	"fmt"
	"math/rand"
	"time"
)

// функция принимает на вход слайс интов и возвращает канал интов
func asChan(vs ...int) <-chan int {
	c := make(chan int) // создаем канал интов
	go func() {
		// в отдельной горутине записываем значения из слайса
		//в канал и спим произвольное количество времени
		for _, v := range vs {
			c <- v
			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		}
		// после записи всех элементов в канал закрываем канал
		close(c)
	}()
	return c
}

// функция объединяет два канала в один
func merge(a, b <-chan int) <-chan int {
	c := make(chan int)
	go func() {
		// в отдельной горутине запускается бесконечный цикл
		// с условия выхода оба исходных канала должны быть nil
		for {
			// select случайным образом определяет (если канал не nil)
			// из какого канала читать (какой кейс выполнит)
			// если канал закрыт, то каналу присваивается nil
			// после вычита данных из исходных каналов, закрывается канал объединения
			// и завершается выполнение горутины
			select {
			case v, ok := <-a:
				if ok {
					c <- v
				} else {
					a = nil
				}
			case v, ok := <-b:
				if ok {
					c <- v
				} else {
					b = nil
				}
			}
			if a == nil && b == nil {
				close(c)
				return
			}
		}
	}()
	return c
}

func main() {
	rand.Seed(time.Now().Unix())
	a := asChan(1, 3, 5, 7)
	b := asChan(2, 4, 6, 8)
	c := merge(a, b)
	for v := range c {
		fmt.Print(v)
	}
}
