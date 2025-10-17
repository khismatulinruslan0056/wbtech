package main

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func test() *customError {
	// ... do something
	return &customError{msg: "ass"}
}

func main() {
	var err error
	err = test()
	// test возвращает указатель на структуру, тк у нас customError реализует интерфейс Error()
	// и переменная err объявлена как тип error, то у нас происходит каст к интерфейсу error
	// у которого data = nil, а tab содержит информацию о типах и методах
	// поэтому у условия err != nil всегда будет true
	// вывод будет: error
	if err != nil {
		println("error")
		return
	}
	println("ok")
}
