package main

import "fmt"

func main() {
	count := 10
	fmt.Println("count:\tValue of[", count, "]\t Add of [", &count, "]")
	increment(count)

	fmt.Println("count:\tValue Of[", count, "]\tAddr Of[", &count, "]")
}

func increment(inc int) {
	inc++
	fmt.Println("inc1:\tValue Of[", inc, "]\tAddr Of[", &inc, "]")
}
