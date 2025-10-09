package main

import "fmt"

func main() {
	str := "snow dog sun"

	strRune := []rune(str)
	reverse(strRune, 0, len(strRune)-1)
	start := 0
	for i := 0; i <= len(strRune); i++ {
		if i == len(strRune) || strRune[i] == ' ' {
			a := string(strRune[start:i])

			reverse(strRune, start, i-1)
			a = string(strRune[start:i])
			_ = a
			start = i + 1
		}
	}

	fmt.Println(str)
	str = string(strRune)
	fmt.Println(str)
}

func reverse(runes []rune, start, end int) {
	for start < end {
		runes[start], runes[end] = runes[end], runes[start]
		start++
		end--
	}

}
