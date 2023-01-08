package main

import "fmt"

func main() {
	names := []string{"nico", "lyn", "dal"}

	names = append(names, "flynn")
	fmt.Println(names[2])
}
