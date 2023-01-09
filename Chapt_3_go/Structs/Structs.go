package main

import "fmt"

type person struct {
	name string
	age  int
	food []string
}

func main() {
	food := []string{"burger", "king"}
	sleeg := person{"sleeg", 15, food}
	//sleeg := person{name: "nico", age: 18, food: food}
	//이런식으로 코딩할 경우 person{name : "nico", 18, food}는 불가
	fmt.Println(sleeg.name)
	//fmt.Println(sleeg)
}
