package main

import "fmt"

func main() {
	sl := []int{1, 45, 12, 7, 43, 2, 9, 0, 5, 32, 75, 333, 3, 3, 3}
	fmt.Println(sl)
	sl = quickSort(sl)
	fmt.Println(sl)
	sl2 := []int{1, 45, 12, 7, 43, 2, 9, 0, 5, 32, 75, 333, 3, 3, 3}
	fmt.Println(sl2)
	quickSort2(sl2)
	fmt.Println(sl2)
}

func quickSort(initSl []int) []int {
	if len(initSl) <= 1 {
		return initSl
	}
	pivot := len(initSl) / 2
	left := make([]int, 0, len(initSl))
	right := make([]int, 0, len(initSl))
	center := make([]int, 0, len(initSl))

	for i := 0; i < len(initSl); i++ {
		if initSl[i] < initSl[pivot] {
			left = append(left, initSl[i])
		} else if initSl[i] > initSl[pivot] {
			right = append(right, initSl[i])
		} else {
			center = append(center, initSl[i])
		}
	}

	left = quickSort(left)
	right = quickSort(right)
	return append(left, append(center, right...)...)
}

func quickSort2(initSl []int) {
	quickSortHelper(initSl, 0, len(initSl)-1)
}

func quickSortHelper(arr []int, low, high int) {
	if low < high {
		pivotIndex := partition(arr, low, high)

		quickSortHelper(arr, low, pivotIndex-1)
		quickSortHelper(arr, pivotIndex+1, high)
	}
}

func partition(arr []int, low, high int) int {
	pivot := arr[high]
	i := low - 1

	for j := low; j < high; j++ {
		var a, b int
		a = arr[j]
		if i >= 0 {
			b = arr[i]
		}
		_, _ = a, b
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}

	arr[i+1], arr[high] = arr[high], arr[i+1]
	return i + 1
}
