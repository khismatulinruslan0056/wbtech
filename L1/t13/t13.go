package main

import "fmt"

func main() {
	a, b := 10, 5
	fmt.Println(a, b)
	a += b
	b = a - b
	a -= b
	fmt.Println(a, b)

	a ^= b
	b = a ^ b
	a = a ^ b
	fmt.Println(a, b)
}

//1 2
//3 2
//
