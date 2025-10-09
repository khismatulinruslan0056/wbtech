package main

import "fmt"

func main() {
	str := "главрыба"
	strRune := []rune(str)
	for i := 0; i < len(strRune)/2; i++ {
		strRune[i], strRune[len(strRune)-1-i] = strRune[len(strRune)-1-i], strRune[i]
	}
	fmt.Println(str)
	str = string(strRune)
	fmt.Println(str)
}
