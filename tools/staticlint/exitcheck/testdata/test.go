package main

import (
	"fmt"
	"os"
)

func main() {
	// формулируем ожидания: анализатор должен находить ошибку,
	// описанную в комментарии want
	i := 0
	fmt.Println(i)
	os.Exit(0) // want "not allowed using of os.Exit()"
}
