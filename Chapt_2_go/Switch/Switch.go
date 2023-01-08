package main

import "fmt"

func canIDrink(age int) bool { //if , else if 남발 생략
	switch koreanAge := age + 2; koreanAge {
	case 10:
		return false
	case 18:
		return true
	case 50:
		return false
	}
	return false
}
func main() {
	fmt.Println(canIDrink(16))
}
