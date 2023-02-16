package main

import "fmt"

type user struct {
	name     string
	username string
}

func main() {
	users1 := make(map[string]user) //string type을 key value = user

	users1["Roy"] = user{"Rob", "Roy"}
	users1["Ford"] = user{"Henry", "Ford"}
	users1["Mouse"] = user{"Mickey", "Mouse"}
	users1["Jackson"] = user{"michae1", "Jackson"}

	//map 순회
	fmt.Println("\n => Iterate over map\n")
	for key, value := range users1 {
		fmt.Println(key, value)
	}

	u1, found1 := users1["Roy"]
	fmt.Println("Roy", found1, u1)
}
