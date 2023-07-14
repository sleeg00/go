package main

import (
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := CreateBlockchain(address)

	UTXOSet := UTXOSet{bc}
	UTXOSet.Reindex() //새로운 블록체인에 생기면 재색인을 한다

	bc.db.Close()
	fmt.Println("Done!")
}
