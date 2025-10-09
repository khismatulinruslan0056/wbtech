package main

import (
	"fmt"
	"unsafe"
)

func main() {

	sl := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println("Удаление с утечкой")
	fmt.Println(sl)
	fmt.Println("len cap pointer")
	fmt.Printf(" %d  %d %v\n", len(sl), cap(sl), unsafe.SliceData(sl))

	sl = deleteElem(sl, 3)
	fmt.Println(sl)
	fmt.Println("len cap pointer")
	fmt.Printf("  %d  %d %v\n\n", len(sl), cap(sl), unsafe.SliceData(sl))

	fmt.Println("Удаление без утечки 1")
	sl = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println(sl)
	fmt.Println("len cap pointer")
	fmt.Printf(" %d  %d %v\n", len(sl), cap(sl), unsafe.SliceData(sl))

	sl = deleteElem2(sl, 3)
	fmt.Println(sl)
	fmt.Println("len cap pointer")
	fmt.Printf("  %d   %d %v\n\n", len(sl), cap(sl), unsafe.SliceData(sl))

	fmt.Println("Удаление без утечки 2")
	sl = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println(sl)
	fmt.Println("len cap pointer")
	fmt.Printf(" %d  %d %v\n", len(sl), cap(sl), unsafe.SliceData(sl))

	sl = deleteElem3(sl, 3)
	fmt.Println(sl)
	fmt.Println("len cap pointer")
	fmt.Printf("  %d   %d %v\n\n", len(sl), cap(sl), unsafe.SliceData(sl))

}

func deleteElem(sl []int, i int) []int {
	if len(sl) == 0 || i < 0 {
		return nil
	}
	return append(sl[:i], sl[i+1:]...)
}

func deleteElem2(sl []int, i int) []int {
	res := make([]int, len(sl)-1)
	copy(res, sl[:i])
	copy(res[i:], sl[i+1:])

	return res
}

func deleteElem3(sl []int, i int) []int {
	res := make([]int, 0, len(sl)-1)
	res = append(res, sl[:i]...)
	res = append(res, sl[i+1:]...)

	return res
}
