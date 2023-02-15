package main

import (
	"fmt"
	"time"
)

func say(j int, c chan string) {
	var result string = ""
	for i := 0; i < j; i++ {

		result += "*"

	}
	time.Sleep(time.Second * 1)
	c <- result
	fmt.Println()

}

func main() {
	c := make(chan string)
	// 함수를 동기적으로 실행

	// 함수를 비동기적으로 실행
	go say(1, c)
	go say(2, c)
	go say(3, c)

	for i := 0; i < 3; i++ {
		fmt.Println(<-c)
	}
	// 3초 대기

}
