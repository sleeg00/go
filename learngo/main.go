package main

import (
	"fmt"
)

func supperAdd(numbers ...int) int {
	defer.f
	total := 0
	for _, number := range numbers {
		total += number
	}
	return total
}
func main() {
	result := supperAdd(1, 2, 3, 4, 5, 6)
	fmt.Println(result)
}
