package main

import (
	"fmt"
	"strings"
)

func lenAndUpper(name string) (lenght int, uppercase string) {
	defer fmt.Println("함수 사용 끝!")
	lenght = len(name)
	uppercase = strings.ToUpper(name)
	return
}

func main() {
	totalLenght, up := lenAndUpper("sleeg")
	fmt.Println(totalLenght, up)
}
