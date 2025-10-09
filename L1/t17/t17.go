package main

import "fmt"

func main() {
	sl := []int{0, 1, 2, 3, 3, 3, 5, 7, 9, 12, 32, 43, 45, 75, 333}
	n := 222

	fmt.Println(sl, n)
	fmt.Println(binarySearch(sl, n))
}

func binarySearch(initSl []int, s int) int {
	start := 0
	end := len(initSl) - 1

	for start <= end {
		mid := start + (end-start)/2

		if initSl[mid] < s {
			start = mid + 1
		} else if initSl[mid] > s {
			end = mid - 1
		} else {
			return mid
		}
	}

	return -1
}
