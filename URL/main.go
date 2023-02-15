package main

import (
	"fmt"
	"net/http"
)

type result struct {
	url    string
	status string
}

func main() {

	c := make(chan result)
	urls := []string{"https://www.google.com",
		"https://www.reddit.com"}

	for _, url := range urls {
		go hitURL(url, c)
	}
	for range urls {
		fmt.Println(<-c)
	}
}

func hitURL(url string, c chan<- result) { //Send Only! <-
	results, err := http.Get(url)

	if err != nil || results.StatusCode >= 400 {
		c <- result{url: url, status: "FAILED"}
	} else {
		c <- result{url: url, status: "SUCCESS"}
	}
}
