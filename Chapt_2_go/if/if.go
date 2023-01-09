package main

import "fmt"

func canIDrink(age int) bool {
	if koreanAge := age + 2; koreanAge < 18 { // age < 18 도 가능 {
		return false
	}
	return true
}

func main() {
	fmt.Println(canIDrink(16))
}
