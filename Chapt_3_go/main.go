package main

import (
	"fmt"

	"github.com/sleeg00/Chapt_3_go/account"
)

func main() {
	account := account.Account(Owner : "sleeg", Balance : 10)
	
	account := account.NewAccount("nico")
	account.Deposit(10)
	fmt.Println(account.Balance())

	err := account.Withdraw(20)
	if err != nil {
		//log.Fatalln(err)
		fmt.Println(err)
	}
	fmt.Println(account.Balance())

	//결과 값 : &{nico 0} -->address이고 복사본이 아니라 객체입니다
}
