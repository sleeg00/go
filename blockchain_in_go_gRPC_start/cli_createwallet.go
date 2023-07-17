package main

import (
	"context"
	"fmt"
	"log"

	blockchain "github.com/sleeg00/blockchain_go/proto"
)

func (cli *CLI) createWallet(nodeID string) {
	request := &blockchain.CreateWalletRequest{
		NodeId: nodeID,
	}

	response, err := cli.blockchain.CreateWallet(context.Background(), request)

	if err != nil {
		log.Printf("Failed to call CreateWallet RPC: %v", err)

	}
	if len(response.Address) <= 0 {
		log.Println("CreateWallet response is nil")

	}

	if len(response.Address) > 0 {
		fmt.Println(response.Address)
	} else {
		fmt.Println("failed : Not Fount Address?")
	}
}

/*
func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}
*/
