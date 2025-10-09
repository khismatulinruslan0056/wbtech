package main

import "fmt"

var temperatures = []float64{-25.4, -27.0, 13.0, 19.0, 15.5, 24.5, -21.0, 32.5}

func main() {
	temps := make(map[int][]float64)

	for _, temperature := range temperatures {
		t := int(temperature) / 10 * 10
		temps[t] = append(temps[t], temperature)
	}

	for k, v := range temps {
		fmt.Printf("%v:[", k)
		for i, t := range v {
			fmt.Printf("%.1f", t)
			if i != len(v)-1 {
				fmt.Print(" ")
			}
		}
		fmt.Print("] ")
	}
}
