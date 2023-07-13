package main

import (
	"fmt"
	"log"
)

func (cli *CLI) createWallet() { //지갑 생성 후 저장 !

	wallets, err := NewWallets()
	if err != nil {
		log.Panic(err)
	}

	address := wallets.CreateWallet()
	wallets.SaveToFile()
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Your new address: %s\n", address)
}
