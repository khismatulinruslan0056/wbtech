package main

import (
	"fmt"
)

type Human struct {
	Name string
	Age  int
}

type Action struct {
	Human
}

func NewHuman(name string, age int) Human {
	return Human{
		Name: name,
		Age:  age,
	}
}

func (h Human) Greet() string {
	return fmt.Sprintf("Hi, I'm %s, I'm %d.", h.Name, h.Age)
}

func main() {
	a := Action{
		Human: NewHuman("John", 40),
	}
	fmt.Println(a.Greet())
}
