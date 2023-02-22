package main

import (
	"fmt"
	"time"
)

func main() {
	time.Sleep(time.Second * 2)
	c := make(chan string)
	for i := 0; i < 10; i++ {
		go User(c)
		go admin(c)
		fmt.Print(<-c + " " + <-c + "\n")
		time.Sleep(time.Second)
	}

}

func User(c chan string) {
	c <- "User"
}

func admin(c chan string) {
	c <- "Admin"
}
