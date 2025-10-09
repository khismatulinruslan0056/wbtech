package main

import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"
)

var justString string

func someFunc() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	before := m.Alloc
	v := createHugeString(1 << 10)
	justString = v[:100]
	runtime.ReadMemStats(&m)
	after := m.Alloc
	fmt.Printf("Выделено памяти: %d байт\n", after-before)

	// проблема с утечкой памяти
	// justString будет ссылаться на массив байт v, поэтому
	// у нас строка будет занимать все те же 1024 байт,
	// а не 100 байт
	a := unsafe.StringData(v)
	b := unsafe.StringData(justString)
	fmt.Println("указатель на v - %v", a)
	fmt.Println("указатель на justString - %v", b)

	//1 решение
	fmt.Println("=============================")
	fmt.Println("1 решение")
	runtime.ReadMemStats(&m)
	before = m.Alloc
	justString = string([]byte(v[:100]))

	runtime.ReadMemStats(&m)
	after = m.Alloc
	fmt.Printf("Выделено памяти: %d байт\n", after-before)
	//  замена базового массива
	//  при приведении к строке создается новый указатель
	// 	ee := []byte{92, 93, 94}
	//	ees := string(ee)
	//	es := string(ee)
	//
	//	eep := unsafe.SliceData(ee)
	//	eesp := unsafe.StringData(ees)
	//	esp := unsafe.StringData(es)
	//	fmt.Println("указатель на eep - %v", eep)
	//	fmt.Println("указатель на ees - %v", eesp)
	//	fmt.Println("указатель на es - %v", esp)
	//  указатель на eep - %v 0x1400010200a
	//	указатель на ees - %v 0x1400010200d
	//	указатель на es - %v 0x14000102020

	//2 решение
	fmt.Println("=============================")
	fmt.Println("2 решение")
	runtime.ReadMemStats(&m)
	before = m.Alloc
	sb := strings.Builder{}
	sb.WriteString(v[:100])
	justString = sb.String()

	runtime.ReadMemStats(&m)
	after = m.Alloc
	fmt.Printf("Выделено памяти: %d байт\n", after-before)

	//3 решение
	fmt.Println("=============================")
	fmt.Println("3 решение")
	runtime.ReadMemStats(&m)
	before = m.Alloc
	newSl := make([]byte, 100)
	copy(newSl, v[:100])
	justString = string(newSl)

	runtime.ReadMemStats(&m)
	after = m.Alloc
	fmt.Printf("Выделено памяти: %d байт\n", after-before)
	//4 решение
	//3 решение
	fmt.Println("=============================")
	fmt.Println("4 решение")
	runtime.ReadMemStats(&m)
	before = m.Alloc
	justString = strings.Clone(v[:100])

	runtime.ReadMemStats(&m)
	after = m.Alloc
	fmt.Printf("Выделено памяти: %d байт\n", after-before)

}

func createHugeString(length int) string {
	sb := strings.Builder{}
	sb.Grow(length)
	for i := 0; i < length; i++ {
		sb.WriteString("a")
	}
	return sb.String()
}

func main() {
	someFunc()
}
