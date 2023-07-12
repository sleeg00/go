package main

import "fmt"

func (cli *CLI) createWallet() { //지갑 생성 후 저장 !
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}
