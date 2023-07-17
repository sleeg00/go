package main

import (
	"fmt"
	"log"
)

/*
	func (cli *CLI) startNode(nodeID, minerAddress string) {
		_, err := cli.blockchain.StartNode(context.Background(),
			&blockchain.StartNodeRequest{
				NodeId:       nodeID,
				MinerAddress: minerAddress,
			})
		if err != nil {
			log.Fatalf("Failed to call StartNode RPC: %v", err)
		}
	}
*/
func (cli *CLI) startNode(nodeID, minerAddress string) {
	fmt.Printf("Starting node %s\n", nodeID)
	if len(minerAddress) > 0 {
		if ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	StartServer2(nodeID, minerAddress)
}
