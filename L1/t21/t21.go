package main

import "fmt"

// Паттерн адаптер позволяет использовать несовместимую библиотеку
// через композицию и реализацию необходимых методов у новой структуры.
// Что позволит использовать старые (несовместимые) методы внутри необходимых методов.
// HandlerFunc является адаптером в стандартной библиотеке.
// Также часто используют адаптер для реализации интерфейса io.Reader.
// Плюсы: упрощение тестирования, изоляция сторонних библиотек (при изменении библиотек меняется только
// адаптер)
// Минусы: Производительность, потеря информации о типах (возврат интерфейса, а не конкретного типа),
// сложности с расширением (адаптер должен реализовывать все методы), проблемы с потерей оригинальных ошибок.

func main() {
	item := NewItem("philips")
	iAd := NewItemAdapter(item)
	iAd.Print()
}

type Item struct {
	name string
}

func NewItem(name string) *Item {
	return &Item{name: name}
}

func (i *Item) Name() string {
	return i.name
}

type ItemAdapter struct {
	item *Item
}

type Printer interface {
	Print()
}

func (a *ItemAdapter) Print() {
	fmt.Println(a.item.Name())
}

func NewItemAdapter(item *Item) *ItemAdapter {
	return &ItemAdapter{item: item}
}
