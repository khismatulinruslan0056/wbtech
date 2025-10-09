package main

import "fmt"

func main() {
	A := []string{"cat", "cat", "dog", "cat", "tree"}
	set := make([]string, 0)
	s := make(map[string]struct{})

	for _, str := range A {
		if _, ok := s[str]; !ok {
			set = append(set, str)
		}
		s[str] = struct{}{}
	}

	fmt.Println(set)
}
