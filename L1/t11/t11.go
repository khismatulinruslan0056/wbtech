package main

import "fmt"

func main() {
	A := []int{1, 2, 3}
	B := []int{2, 3, 4}

	set := make([]int, 0)

	s := make(map[int]struct{})

	for _, n := range A {
		s[n] = struct{}{}
	}

	for _, n := range B {
		if _, ok := s[n]; ok {
			set = append(set, n)
		}
	}

	fmt.Println(set)

}
