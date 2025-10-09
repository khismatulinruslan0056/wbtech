package main

import "fmt"

func main() {

	// устанавливаем 1
	fmt.Println(setBit(5, 0, 1))
	// устанавливаем 0
	fmt.Println(setBit(5, 0, 0))

}

func setBit(num, i, oneOrZero int) int {
	mask := 1 << i
	if oneOrZero == 1 {
		return num | mask
	}

	return num &^ mask
}
