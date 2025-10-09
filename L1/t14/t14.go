package main

import (
	"fmt"
	"reflect"
)

func main() {
	var (
		a int    = 1
		b string = "asd"
		c bool
		d chan interface{}
		e [1]int
	)

	Type(a)
	Type(b)
	Type(c)
	Type(d)
	Type(e)
}

func Type(p interface{}) {
	var t string
	switch p.(type) {
	case int:
		t = "int"
	case string:
		t = "string"
	case bool:
		t = "bool"
	default:
		if reflect.TypeOf(p).Kind() == reflect.Chan {
			t = "chan"
		} else {
			t = "unknown"
		}
	}
	fmt.Printf("%v is %s\n", p, t)
}
