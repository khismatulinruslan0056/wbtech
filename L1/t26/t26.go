package main

import (
	"fmt"
	"strings"
)

func main() {
	strs := []string{"abcd", "abCdefAaf", "aabcd"}
	for _, str := range strs {
		fmt.Printf("%q -> %v\n", str, uniqStr(str))
	}
}

func uniqStr(str string) bool {
	str = strings.ToLower(str)
	mapUn := make(map[rune]struct{})
	strB := []rune(str)
	for _, r := range strB {
		if _, ok := mapUn[r]; ok {
			return false
		}
		mapUn[r] = struct{}{}
	}

	return true
}
