package main

import (
	"fmt"
	"math"
	"math/big"
)

func main() {
	a := int64(1 << 50)
	b := int64(1 << 23)
	for {
		fmt.Println("Введите операцию +, -, *, / или 0 для выхода")
		var op string
		fmt.Scanln(&op)

		switch op {
		case "+":
			if math.MaxInt64-b < a {
				fmt.Println("b")
				aBig := big.NewInt(a)
				bBig := big.NewInt(b)
				sum := new(big.Int).Add(aBig, bBig)
				fmt.Printf("a + b = %s\n", sum.String())
			} else {
				fmt.Println("i")
				fmt.Printf("a + b = %d\n", a+b)
			}
		case "-":
			if math.MinInt64-b < a {
				fmt.Println("b")
				aBig := big.NewInt(a)
				bBig := big.NewInt(b)
				diff := new(big.Int).Sub(aBig, bBig)
				fmt.Printf("a - b = %s\n", diff.String())
			} else {
				fmt.Println("i")
				fmt.Printf("a - b = %d\n", a-b)
			}
		case "*":
			if math.MaxInt64/b < a {
				fmt.Println("b")
				aBig := big.NewInt(a)
				bBig := big.NewInt(b)
				prod := new(big.Int).Mul(aBig, bBig)
				fmt.Printf("a * b = %s\n", prod.String())
			} else {
				fmt.Println("i")
				fmt.Printf("a * b = %d\n", a*b)
			}
		case "/":
			fmt.Printf("a / b = %d\n", a/b)
		case "0":
			return
		default:
			fmt.Println("Неизвестная операция")
		}
	}

}
