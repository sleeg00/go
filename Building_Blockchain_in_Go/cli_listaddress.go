package main

import (
	"fmt"
	"log"
)

func (cli *CLI) listAddresses() { //지갑들 출력
	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}
