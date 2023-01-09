package main

import "fmt"

func main() {
	sleeg := map[string]string{"name": "sleeg", "age": "12"} //name이 key age는 Value
	//fmt.Println(sleeg)

	for key, value := range sleeg {
		fmt.Println(key, value)
	}

	for _, value := range sleeg {
		fmt.Println(value)
	}

}
