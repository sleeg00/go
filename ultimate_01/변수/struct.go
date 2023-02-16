package main

import "fmt"

type example struct {
	counter int64
	pi      float32
	flag    bool
}

func main() {
	e2 := example{
		flag:    true,
		counter: 10,
		pi:      3.141592,
	}

	fmt.Println("Flag", e2.flag)
	fmt.Println("Counter", e2.counter)
	fmt.Println("Pi", e2.flag)

}
