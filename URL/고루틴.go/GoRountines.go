package main

import (
	"fmt"
	"time"
)

func main() { //	main이 끝나면 go루틴도 종료된다 main은 wait해주지 않는다!
	c := make(chan string)
	people := [4]string{"sleeg", "flynn", "POP", "PUSH"}
	for _, person := range people {
		go isSexy(person, c)
	}

	for i := 0; i < len(people); i++ {
		fmt.Println(<-c)
	}
	//go sexyCount("Sleeg")
	//go sexyCount("flynn")
	//time.Sleep(time)

}

/*
func sexyCount(person string) {

	for i := 0; i < 10; i++ {
		fmt.Println(person, "is sexy", i)
		time.Sleep(time.Second)
	}

}
*/

func isSexy(person string, c chan string) {
	time.Sleep((time.Second * 2))
	c <- person + " is sexy"
}
